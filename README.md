# limiter
golang concurrency limiter

## Install

```bash
go get github.com/iccolo/limiter
```

## Usage

```go
func main() {
    limiter.Run([]func() error{
        func() error {
            return nil
        },
    }, 10)
}

```