package updates

import "go.uber.org/zap"

type task struct {
	name string
	f    func() error
}

type executor struct {
	tasks  chan task
	done   chan struct{}
	logger *zap.Logger
}

func newExecutor(logger *zap.Logger) *executor {
	exec := &executor{
		tasks:  make(chan task, 10),
		done:   make(chan struct{}),
		logger: logger,
	}

	go exec.run()
	return exec
}

func (e *executor) EnqueueTask(name string, f func() error) {
	e.tasks <- task{name, f}
}

func (e *executor) run() {
	defer close(e.done)

	for task := range e.tasks {
		if err := task.f(); err != nil {
			e.logger.Error(task.name, zap.Error(err))
		}
	}
}

func (e *executor) Close() { close(e.tasks); <-e.done }
