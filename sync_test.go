package sync

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestMaxcc1(t *testing.T) {
	m := NewMaxCc(3)
	for i := 0; i < 10; i++ {
		m.Add()
		go func() {
			defer m.Done()
			fmt.Printf("NumGoroutine: %d\n", runtime.NumGoroutine())
			time.Sleep(time.Millisecond * 100)
		}()
	}
	m.Wait()
}

func TestMaxcc2(t *testing.T) {
	m := NewMaxCc(10)
	for i := 0; i < 3; i++ {
		m.Add()
		go func() {
			defer m.Done()
			fmt.Printf("NumGoroutine: %d\n", runtime.NumGoroutine())
			time.Sleep(time.Millisecond * 100)
		}()
	}
	m.Wait()
}

func TestMaxcc3(t *testing.T) {
	m := NewMaxCc(10)
	m.Add()
	m.Done()
	m.Wait()
	m.Add()
	m.Done()
	m.Wait()
}
