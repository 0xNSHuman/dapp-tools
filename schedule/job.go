package schedule

import "time"

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//								  JOB CONFIG
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type JobConfig struct {
	tType  JobType
	period time.Duration
}

func NewPrimitiveJobConfig() JobConfig {
	return JobConfig{
		tType:  JobTypePrimitive,
		period: -1,
	}
}

func NewPeriodicJobConfig(period time.Duration) JobConfig {
	return JobConfig{
		tType:  JobTypePeriodic,
		period: period,
	}
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//										JOB
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type Job struct {
	id     uint32
	config JobConfig
	task   func()
}

func NewJob(
	id uint32,
	config JobConfig,
	task func(),
) *Job {
	job := &Job{
		id,
		config,
		task,
	}

	return job
}

func (j *Job) Run() {
	j.task()
}
