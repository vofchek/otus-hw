package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("test maxErrorCounter <= 0", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 5

		require.NoError(t, Run(tasks, workersCount, -1))
	})

	t.Run("test with worker count <= 0", func(t *testing.T) {
		tasksCount := 1
		tasks := make([]Task, 0, tasksCount)

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				return nil
			})
		}

		require.ErrorIs(t, ErrErrorsNonPositiveWorkerNumber, Run(tasks, 0, 1))
		require.ErrorIs(t, ErrErrorsNonPositiveWorkerNumber, Run(tasks, -10, 1))
	})

	t.Run("test with empty tasks", func(t *testing.T) {
		require.Nil(t, Run(make([]Task, 0), 0, 1))
	})
}

func TestRunEventually(t *testing.T) {
	t.Run("test with require.Eventually counter locked by mutex", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)
		runTasksCount := 0
		mu := &sync.Mutex{}

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				mu.Lock()
				defer mu.Unlock()
				runTasksCount++
				return nil
			})
		}

		go func() {
			require.NoError(t, Run(tasks, 5, 0), "run returned error")
		}()

		require.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()

			return runTasksCount == tasksCount
		}, time.Second*10, time.Millisecond*100, "not all tasks completed")
	})

	t.Run("test with require.Eventually count through special channel", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)
		runTasksCount := 0
		counterChannel := make(chan int, 1)

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				counterChannel <- 1
				return nil
			})
		}

		go func() {
			require.NoError(t, Run(tasks, 5, 0), "run returned error")
			close(counterChannel)
		}()

		require.Eventually(t, func() bool {
			for {
				_, more := <-counterChannel
				if !more {
					break
				}
				runTasksCount++
			}

			return runTasksCount == tasksCount
		}, time.Second*10, time.Millisecond*100, "not all tasks completed")
	})
}
