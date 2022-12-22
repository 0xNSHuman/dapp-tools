package schedule

import "github.com/0xNSHuman/dapp-tools/common"

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//								  	ERRORS
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type ScheduleError uint

const (
	Unknown ScheduleError = common.ErrorDomainSchedule + iota
	JobNotFound
	DuplicateJob
)

func (e ScheduleError) Error() string {
	switch e {
	case DuplicateJob:
		return "Job is already scheduled"
	default:
		return "Unknown"
	}
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//									JOB TYPE
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type JobType uint8

const (
	// Runs once and dies
	JobTypePrimitive JobType = iota
	// Runs every period even if the previous execution hasn't finished
	JobTypePeriodic
	// Runs in cycle, always waiting for the previous execution to finish
	JobTypeCyclical
)

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//								  JOB STATE
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type JobState uint8

const (
	JobStateScheduled JobState = iota
	JobStateRunning
	JobStateCompleted
	JobStateCancelled
)
