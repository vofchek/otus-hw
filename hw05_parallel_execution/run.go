package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

var ErrErrorsNonPositiveWorkerNumber = errors.New("workers number must be positive")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 {
		return nil
	}

	if n <= 0 {
		return ErrErrorsNonPositiveWorkerNumber
	}

	// нужен небуфферизованный канал, чтобы воркер брал задачу когда готов.
	// иначе в канале могут остаться задачи на выполнение, хотя уже накопилось достаточно ошибок.
	taskChannel := make(chan Task)

	safeErrorCounter := newSafeErrorCounter(m)
	wg := &sync.WaitGroup{}

	// consumers
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for task := range taskChannel {
				err := task()
				if err != nil {
					safeErrorCounter.increase()
				}
			}
		}()
	}

	// produce
	for _, task := range tasks {
		if safeErrorCounter.tooMuch() {
			break
		}

		taskChannel <- task
	}
	close(taskChannel)

	wg.Wait()

	if safeErrorCounter.tooMuch() {
		return ErrErrorsLimitExceeded
	}

	return nil
}

type safeErrorCounter struct {
	errorCounter int
	maxErrors    int
	mu           *sync.Mutex
}

func (c *safeErrorCounter) increase() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.errorCounter++
}

func (c *safeErrorCounter) tooMuch() bool {
	if c.maxErrors <= 0 {
		return false
	}

	return c.get() >= c.maxErrors
}

func (c *safeErrorCounter) get() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.errorCounter
}

func newSafeErrorCounter(maxErrors int) *safeErrorCounter {
	return &safeErrorCounter{mu: &sync.Mutex{}, maxErrors: maxErrors}
}
