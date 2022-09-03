package sync

// MaxCc provide control method for max cc goroutine number
type MaxCc interface {
	// Add increase counter once, block if reach max limit
	Add()
	// Done decrease counter once, block if reach zero limit
	Done()
	// Wait wait all goroutine done
	Wait()
}

// NewMaxCc create MaxCc
func NewMaxCc(max int) MaxCc {
	return &maxCc{cc: make(chan struct{}, max)}
}

type maxCc struct {
	cc chan struct{}
}

// Add increase counter once, block if reach max limit
func (m *maxCc) Add() {
	m.cc <- struct{}{}
}

// Done decrease counter once, block if reach zero limit
func (m *maxCc) Done() {
	<-m.cc
}

// Wait wait all goroutine done
func (m *maxCc) Wait() {
	n := cap(m.cc)
	for i := 0; i < n; i++ {
		m.cc <- struct{}{}
	}
	for i := 0; i < n; i++ {
		<-m.cc
	}
}
