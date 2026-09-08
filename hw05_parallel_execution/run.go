package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	// нужен небуфферизованный канал, иначе не работает
	taskChannel := make(chan Task)
	safeErrorCounter := newSafeErrorCounter(m)
	wg := &sync.WaitGroup{}

	// produce
	go func() {
		defer close(taskChannel)

		for _, task := range tasks {
			if safeErrorCounter.tooMuch() {
				return
			}

			taskChannel <- task
		}
	}()

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

	wg.Wait()

	if safeErrorCounter.tooMuch() {
		return ErrErrorsLimitExceeded
	}

	return nil
}

type safeErrorCounter struct {
	errorCounter int
	maxErrors    int
	sm           *sync.Mutex
}

func (safeErrorCounter *safeErrorCounter) increase() {
	safeErrorCounter.sm.Lock()
	defer safeErrorCounter.sm.Unlock()

	safeErrorCounter.errorCounter++
}

func (safeErrorCounter *safeErrorCounter) tooMuch() bool {
	if safeErrorCounter.maxErrors <= 0 {
		return false
	}

	return safeErrorCounter.get() >= safeErrorCounter.maxErrors
}

func (safeErrorCounter *safeErrorCounter) get() int {
	safeErrorCounter.sm.Lock()
	defer safeErrorCounter.sm.Unlock()

	return safeErrorCounter.errorCounter
}

func newSafeErrorCounter(maxErrors int) *safeErrorCounter {
	return &safeErrorCounter{sm: &sync.Mutex{}, maxErrors: maxErrors}
}
