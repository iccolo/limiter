package limiter

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewLimiter(t *testing.T) {
	limit := 5
	l := NewLimiter(limit)
	if l == nil {
		t.Fatal("NewLimiter returned nil")
	}

	lImpl, ok := l.(*limiter)
	if !ok {
		t.Fatal("NewLimiter returned wrong type")
	}

	if cap(lImpl.ch) != limit {
		t.Fatalf("expected channel capacity to be %d, but got %d", limit, cap(lImpl.ch))
	}
}

func TestLimiter_AddDone(t *testing.T) {
	l := NewLimiter(2)

	// test Add does not block
	done := make(chan bool)
	go func() {
		l.Add()
		l.Add()
		done <- true
	}()

	select {
	case <-done:
		// normal
	case <-time.After(time.Second):
		t.Fatal("Add operation timeout")
	}

	// test third Add will be blocked
	blocked := make(chan bool)
	go func() {
		l.Add()
		blocked <- false
	}()

	select {
	case <-blocked:
		t.Fatal("third Add should be blocked")
	case <-time.After(time.Millisecond * 100):
		// normal, Add is blocked
	}

	// test Done
	l.Done()

	select {
	case <-blocked:
		// normal, Add is not blocked
	case <-time.After(time.Second):
		t.Fatal("Done after Add is blocked")
	}
}

func TestLimiter_Wait(t *testing.T) {
	l := NewLimiter(3)
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Add()
			time.Sleep(time.Millisecond * 100)
			l.Done()
		}()
	}

	wg.Wait()

	done := make(chan bool)
	go func() {
		l.Wait()
		done <- true
	}()

	select {
	case <-done:
		// normal
	case <-time.After(time.Second):
		t.Fatal("Wait operation timeout")
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name          string
		handlers      []func() error
		limit         int
		opts          []Option
		expectedError bool
	}{
		{
			name: "all handlers success",
			handlers: []func() error{
				func() error { time.Sleep(time.Millisecond * 10); return nil },
				func() error { time.Sleep(time.Millisecond * 20); return nil },
				func() error { time.Sleep(time.Millisecond * 30); return nil },
			},
			limit:         2,
			expectedError: false,
		},
		{
			name: "one handler failed, not break",
			handlers: []func() error{
				func() error { return nil },
				func() error { return errors.New("错误") },
				func() error { return nil },
			},
			limit:         2,
			expectedError: true,
		},
		{
			name: "one handler failed, break",
			handlers: []func() error{
				func() error { time.Sleep(time.Millisecond * 10); return nil },
				func() error { return errors.New("错误") },
				func() error { time.Sleep(time.Millisecond * 100); return nil },
			},
			limit:         2,
			opts:          []Option{WithBreakOnError(true)},
			expectedError: true,
		},
		{
			name: "handler panic",
			handlers: []func() error{
				func() error { panic("panic") },
				func() error { return nil },
			},
			limit:         1,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(tt.handlers, tt.limit, tt.opts...)
			if (err != nil) != tt.expectedError {
				t.Errorf("Run() error = %v, expected error %v", err, tt.expectedError)
			}
		})
	}
}

func TestWithBreakOnError(t *testing.T) {
	opt := WithBreakOnError(true)
	options := &Options{}
	opt(options)

	if !options.breakOnError {
		t.Error("WithBreakOnError did not set breakOnError option correctly")
	}
}
