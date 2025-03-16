package main

import (
	"context"
	"time"
)

type BackgroundTask struct {
	name     string
	Interval time.Duration
	Fn       func(context.Context)
}

func (tsk *BackgroundTask) Name() string {
	return tsk.name
}

func NewTask(name string, interval time.Duration, fn func(context.Context)) *BackgroundTask {
	return &BackgroundTask{
		Interval: interval,
		Fn:       fn,
	}
}

func (tsk *BackgroundTask) Start(ctx context.Context) error {
	tsk.Fn(ctx)

	timer := time.NewTimer(tsk.Interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			timer.Reset(tsk.Interval)
			tsk.Fn(ctx)
		}
	}
}
