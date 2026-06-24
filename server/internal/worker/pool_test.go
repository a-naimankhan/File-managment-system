package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type MockTask struct {
	executionTime time.Duration
	executed      *atomic.Bool
	done          chan struct{}
}

func (m *MockTask) Execute(ctx context.Context) error {
	defer close(m.done)
	defer m.executed.Store(true)

	time.Sleep(m.executionTime)
	return nil
}

func TestPoolStopWaitsForTasks(t *testing.T) {

	pool := NewPool(2)

	ctx := context.Background()
	pool.Start(ctx)

	task1 := &MockTask{
		executionTime: 100 * time.Millisecond,
		executed:      &atomic.Bool{},
		done:          make(chan struct{}),
	}

	task2 := &MockTask{
		executionTime: 50 * time.Millisecond,
		executed:      &atomic.Bool{},
		done:          make(chan struct{}),
	}

	pool.Submit(task1)
	pool.Submit(task2)

	time.Sleep(10 * time.Millisecond)

	stopTime := time.Now()
	pool.Stop()
	stopDuration := time.Since(stopTime)

	if !task1.executed.Load() {
		t.Error("task1 was not executed before Stop() returned")
	}

	if !task2.executed.Load() {
		t.Error("task2 was not executed before Stop() returned")
	}

	if stopDuration < 100*time.Millisecond {
		t.Errorf("Stop() returned too quickly: %v, expected at least ~100ms", stopDuration)
	}

	t.Log("Test passed: Stop() correctly waited for all tasks to complete")
}

func TestPoolStopMultipleTasks(t *testing.T) {
	pool := NewPool(2)
	ctx := context.Background()
	pool.Start(ctx)

	taskCount := 5
	var wg sync.WaitGroup

	for i := 0; i < taskCount; i++ {
		task := &MockTask{
			executionTime: 50 * time.Millisecond,
			executed:      &atomic.Bool{},
			done:          make(chan struct{}),
		}

		wg.Add(1)
		go func(t *MockTask) {
			defer wg.Done()
			pool.Submit(t)
		}(task)
	}

	wg.Wait()

	pool.Stop()

	t.Logf("Test passed: All tasks were processed before Stop() returned")
}

func TestPoolStopImmediatelyWithoutTasks(t *testing.T) {
	pool := NewPool(2)
	ctx := context.Background()
	pool.Start(ctx)

	startTime := time.Now()
	pool.Stop()
	duration := time.Since(startTime)

	if duration > 1*time.Second {
		t.Errorf("Stop() took too long without tasks: %v", duration)
	}

	t.Log("Test passed: Stop() works correctly without tasks")
}

func TestPoolStopCannotSubmitAfterStop(t *testing.T) {
	pool := NewPool(2)
	ctx := context.Background()
	pool.Start(ctx)

	initialTask := &MockTask{
		executionTime: 10 * time.Millisecond,
		executed:      &atomic.Bool{},
		done:          make(chan struct{}),
	}
	pool.Submit(initialTask)

	pool.Stop()

	afterStopTask := &MockTask{
		executionTime: 10 * time.Millisecond,
		executed:      &atomic.Bool{},
		done:          make(chan struct{}),
	}

	pool.Submit(afterStopTask)

	if !initialTask.executed.Load() {
		t.Error("initial task was not executed")
	}

	if afterStopTask.executed.Load() {
		t.Error("task submitted after Stop() should not be executed")
	}

	t.Log("Test passed: Cannot submit tasks after Stop()")
}

func TestPoolConcurrentOperations(t *testing.T) {
	pool := NewPool(3)
	ctx := context.Background()
	pool.Start(ctx)

	executedTasks := &atomic.Int32{}
	taskCount := 20

	var wg sync.WaitGroup
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			task := &MockTask{
				executionTime: 20 * time.Millisecond,
				executed:      &atomic.Bool{},
				done:          make(chan struct{}),
			}
			pool.Submit(task)

			<-task.done
			executedTasks.Add(1)
		}()
	}

	wg.Wait()

	pool.Stop()

	if executedTasks.Load() != int32(taskCount) {
		t.Errorf("expected %d executed tasks, got %d", taskCount, executedTasks.Load())
	}

	t.Log("Test passed: All concurrent tasks were executed correctly")
}
