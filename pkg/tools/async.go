package tools

import (
	"context"
)

type AsyncTask struct {
	Fn func() error
}

func (task *AsyncTask) Run() error {
	return task.Fn()
}

type Task interface {
	Run() error
}

type Status struct {
	err     error
	ctxDone bool
}

func (s Status) DoneOrCanceled() (bool, error) {
	return s.ctxDone, s.err
}

type Async struct {
	task Task
	done chan error
}

func RunAsync(fn func() error) *Async {
	task := &AsyncTask{Fn: fn}
	async := NewAsync(task)
	async.Run()
	return async
}

func NewAsync(task Task) *Async {
	return &Async{
		task: task,
		done: make(chan error),
	}
}

func (a *Async) Run() {
	go func() {
		a.done <- a.task.Run()
	}()
}

func (a *Async) WaitContext(ctx context.Context) Status {
	select {
	case <-ctx.Done():
		return Status{err: nil, ctxDone: true}
	case err := <-a.done:
		return Status{err: err, ctxDone: false}
	}
}

func (a *Async) Wait() Status {
	return Status{
		err:     <-a.done,
		ctxDone: false,
	}
}
