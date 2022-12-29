package mobile

import "github.com/0xNSHuman/dapp-tools/schedule"

type Job struct {
	job *schedule.Job
}

type Scheduler struct {
	scheduler *schedule.Scheduler
}

func (s *Scheduler) AddJob(config schedule.JobConfig, task func()) (*Job, error) {
	job, err := s.scheduler.AddJob(config, task)
	return &Job{job}, err
}
