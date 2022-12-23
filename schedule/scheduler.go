package schedule

import (
	"math"
	"time"
)

var jobIdCounter uint32 = 0

func nextJobID() uint32 {
	jobIdCounter++
	return jobIdCounter
}

type Scheduler struct {
	// Main table for all scheduled jobs
	jobTable map[uint32]*Job

	// Job state history table
	jobStateTable map[uint32]JobState

	// Next Unix run time for periodic jobs in nanoseconds
	jobSchedule map[int64]*Job

	// The closest scheduled job time in the future (or max int64)
	nextJobTime int64

	// Non-recurring jobs state processing queue
	queue chan *Job

	// The new shortest job period for the scheduler to check
	periodReset chan int64
}

func NewScheduler() *Scheduler {
	scheduler := &Scheduler{
		jobTable:      make(map[uint32]*Job),
		jobStateTable: make(map[uint32]JobState),
		jobSchedule:   make(map[int64]*Job),
		nextJobTime:   math.MaxInt64,
		queue:         make(chan *Job),
		periodReset:   make(chan int64),
	}

	go scheduler.launch()

	return scheduler
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//								PUBLIC METHODS
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func (s *Scheduler) AddJob(config JobConfig, task func()) (*Job, error) {
	job := NewJob(nextJobID(), config, task)

	s.jobTable[job.ID] = job
	s.jobStateTable[job.ID] = JobStateScheduled
	s.queue <- job

	switch job.config.tType {
	case JobTypePeriodic:
		nextPeriodRunTime := time.Now().UnixNano() + job.config.period.Nanoseconds()

		// Place the job in the time schedule
		s.jobSchedule[nextPeriodRunTime] = job

		// If the job is now the earliest in schedule
		// — reset min period tracking for the scheduler
		if nextPeriodRunTime < s.nextJobTime {
			s.nextJobTime = nextPeriodRunTime
			s.periodReset <- job.config.period.Nanoseconds()
		}
	}

	return job, nil
}

func (s *Scheduler) CancelJob(id uint32) {
	job := s.jobTable[id]

	delete(s.jobTable, id)
	s.jobStateTable[id] = JobStateCancelled

	switch job.config.tType {
	case JobTypePeriodic:
		nextPeriodRunTime := time.Now().UnixNano() + job.config.period.Nanoseconds()

		// Place the job in the time schedule
		s.jobSchedule[nextPeriodRunTime] = job

		// If the job is now the earliest in schedule
		// — reset min period tracking for the scheduler
		if nextPeriodRunTime < s.nextJobTime {
			s.nextJobTime = nextPeriodRunTime
			s.periodReset <- job.config.period.Nanoseconds()
		}
	}
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//								PRIVATE METHODS
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func (s *Scheduler) launch() {
	go s.runPeriodWatcher()

	for {
		job := <-s.queue

		if s.jobStateTable[job.ID] == JobStateCancelled {
			// skip cancelled job
			continue
		}

		s.jobStateTable[job.ID] = JobStateRunning
		go s.runJob(job)

		// Before completion
		switch job.config.tType {
		case JobTypePeriodic:
			// job is rescheduled
			s.jobStateTable[job.ID] = JobStateScheduled
		}
	}
}

func (s *Scheduler) runJob(job *Job) {
	job.Run()

	// After completion
	switch job.config.tType {
	case JobTypePrimitive:
		// remove the job from the main table
		delete(s.jobTable, job.ID)
		// but keep the state table growing for the duration of the execution
		s.jobStateTable[job.ID] = JobStateCompleted
	}
}

func (s *Scheduler) runPeriodWatcher() {
	var terminator chan uint8

	for {
		newPeriod := <-s.periodReset // new period signaled

		if terminator != nil {
			terminator <- 0 // invalidate the prev period
		} else {
			terminator = make(chan uint8)
		}

		ticker := time.NewTicker(time.Duration(newPeriod)).C
		go s.processPeriodicSchedule(ticker, terminator) // start tracking updated smallest job period
	}
}

func (s *Scheduler) processPeriodicSchedule(ticker <-chan time.Time, terminator chan uint8) {
	for {
		select {
		case <-ticker: // period passed
			now := time.Now().UnixNano()
			if now >= int64(s.nextJobTime) {
				job := s.jobSchedule[s.nextJobTime]

				// Skip the job if it's cancelled and clean up the schedule
				// Otherwise reschedule it with new run time
				if s.jobStateTable[job.ID] == JobStateCancelled {
					delete(s.jobSchedule, s.nextJobTime)
				} else {
					nextPeriodRunTime := s.nextJobTime + job.config.period.Nanoseconds()
					s.jobSchedule[nextPeriodRunTime] = job

					// Put the job in the processing queue
					s.queue <- s.jobSchedule[s.nextJobTime]
				}

				// Set the next closest job time across schedule
				// to be processed next
				s.nextJobTime = s.lookNextJobTimeUp()
			}
		case <-terminator: // period invalidated
			return
		}
	}
}

// Returns the next closest registered job schedule time
// or max int64 value if nothing is found
func (s *Scheduler) lookNextJobTimeUp() int64 {
	var nextJobTime int64 = math.MaxInt64

	for time := range s.jobSchedule {
		if time < nextJobTime {
			nextJobTime = time
		}
	}

	return nextJobTime
}
