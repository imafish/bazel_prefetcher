package main

import (
	"errors"
	"log"
	"time"
)

type scheduler struct {
	interval  int
	startTime time.Time
	endTime   time.Time

	lastRun time.Time
}

func NewScheduler(interval int, startTime, endTime string) (*scheduler, error) {
	if interval <= 0 {
		return nil, errors.New("interval must be greater than 0")
	}

	layout := "15:04"
	start, err := time.Parse(layout, startTime)
	if err != nil {
		return nil, errors.New("invalid startTime format")
	}
	end, err := time.Parse(layout, endTime)
	if err != nil {
		return nil, errors.New("invalid endTime format")
	}

	return &scheduler{
		interval:  interval,
		startTime: start,
		endTime:   end,
	}, nil
}

func (s *scheduler) shouldRun() bool {
	now := time.Now()
	currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), 0, time.UTC)
	log.Printf("current time: %s, last run: %s, interval: %d seconds", currentTime.Format("15:04"), s.lastRun.Format("15:04"), s.interval)

	if currentTime.Before(s.startTime) || currentTime.After(s.endTime) {
		if s.lastRun.IsZero() || time.Since(s.lastRun) >= time.Duration(s.interval)*time.Second {
			return true
		}
	}
	return false
}

func (s *scheduler) sleepTime() time.Duration {
	now := time.Now()
	currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), 0, time.UTC)

	var sleepDuration time.Duration
	if currentTime.After(s.startTime) && currentTime.Before(s.endTime) {
		sleepDuration = s.endTime.Sub(currentTime)
	}
	nextRun := s.lastRun.Add(time.Duration(s.interval) * time.Second)
	sleepDuration2 := time.Until(nextRun)
	if sleepDuration2 > sleepDuration {
		sleepDuration = sleepDuration2
	}
	return sleepDuration
}

func (s *scheduler) Run(f func() error) error {
	for {
		if s.shouldRun() {
			log.Printf("running scheduled function...")
			startTime := time.Now()
			if err := f(); err != nil {
				log.Printf("error running scheduled function: %v", err)
			} else {
				log.Printf("scheduled function executed successfully")
			}
			log.Printf("elapsed time since last run: %v", time.Since(startTime))
			s.lastRun = startTime
		} else {
			log.Printf("not in scheduled time, waiting...")
		}

		sleepDuration := s.sleepTime()
		if sleepDuration > 0 {
			log.Printf("sleeping for %v", sleepDuration)
			time.Sleep(sleepDuration)
		} else {
			log.Printf("no sleep needed, continuing...")
		}
	}
}
