package limiter

// Limiter provide control method for max concurrent goroutine number
type Limiter interface {
	// Add increase counter once, block if reach max limit
	Add()
	// Done decrease counter once, block if reach zero limit
	Done()
	// Wait wait all goroutine done
	Wait()
}

// NewLimiter create Limiter
func NewLimiter(limit int) Limiter {
	return &limiter{ch: make(chan struct{}, limit)}
}

type limiter struct {
	ch chan struct{}
}

// Add increase counter once, block if reach max limit
func (m *limiter) Add() {
	m.ch <- struct{}{}
}

// Done decrease counter once, block if reach zero limit
func (m *limiter) Done() {
	<-m.ch
}

// Wait wait all goroutine done
func (m *limiter) Wait() {
	n := cap(m.ch)
	for i := 0; i < n; i++ {
		m.ch <- struct{}{}
	}
	for i := 0; i < n; i++ {
		<-m.ch
	}
}


// Run execute handlers with concurrency limit
func Run(handlers []func() error, limit int, opts ...Option) error {
	options := defaultOptions
	for _, o := range opts {
		o(&options)
	}
	var (
		limiter = NewLimiter(limit)
		err     atomic.Value
	)
	for _, f := range handlers {
		if options.breakOnError && err.Load() != nil {
			limiter.Wait()
			return err.Load().(error)
		}
		limiter.Add()
		go func(handler func() error) {
			defer func() {
				if e := recover(); e != nil {
					buf := make([]byte, 1024)
					buf = buf[:runtime.Stack(buf, false)]
					log.Errorf("run panic, buf:%s", string(buf))
					err.Store(errs.New(errs.RetServerSystemErr, "panic found in call handlers"))
				}
				limiter.Done()
			}()
			if e := handler(); e != nil {
				err.Store(e)
			}
		}(f)
	}
	limiter.Wait()
	if err.Load() != nil {
		return err.Load().(error)
	}
	return nil
}

// Options options
type Options struct {
	breakOnError bool // if error, stop execution, otherwise wait for all handlers to execute
}

var defaultOptions = Options{
	breakOnError: false,
}

// Option option
type Option func(*Options)

// WithBreakOnError set break on error
func WithBreakOnError(b bool) Option {
	return func(o *Options) {
		o.breakOnError = b
	}
}
