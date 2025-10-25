package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManagerRun(t *testing.T) {
	input := Input{
		Reader:  bytes.NewBuffer([]byte("hello")),
		Options: StreamOptions{},
		Command: []string{"echo", "hello"},
	}

	// basic smoke: replace runStream to a no-op echo so it doesn't spawn real process
	old := runStream
	defer func() { runStream = old }()

	runStream = func(ctx context.Context, options StreamOptions, command ...string) (StreamResult, error) {
		prOut, pwOut := io.Pipe()
		prIn, pwIn := io.Pipe()

		go func() {
			defer pwOut.Close()
			for {
				data, err := readChunk(prIn)
				if err != nil {
					return
				}
				err = writeChunk(pwOut, data)
				if err != nil {
					return
				}
			}
		}()
		return StreamResult{
			Stdin:    pwIn,
			Stdout:   prOut,
			Stderr:   io.NopCloser(bytes.NewReader(nil)),
			waitFunc: func() error { return nil },
		}, nil
	}

	out, err := SharedRunner.Run(input)
	assert.NoError(t, err, "Run error")
	assert.Equal(t, []byte("hello"), out, "unexpected output")

	input.Reader = bytes.NewBuffer([]byte("hello2"))
	out, err = SharedRunner.Run(input)
	assert.NoError(t, err, "Run error")
	assert.Equal(t, []byte("hello2"), out, "unexpected output")
}

func TestSharedRunnerConcurrent(t *testing.T) {
	// ensure a clean SharedRunner
	SharedRunner = manager{cmds: make(map[string]*runner)}

	old := runStream
	defer func() { runStream = old }()

	runStream = func(ctx context.Context, options StreamOptions, command ...string) (StreamResult, error) {
		prOut, pwOut := io.Pipe()
		prIn, pwIn := io.Pipe()

		go func() {
			defer pwOut.Close()
			for {
				data, err := readChunk(prIn)
				if err != nil {
					return
				}
				time.Sleep(50 * time.Millisecond) // simulate some processing delay
				err = writeChunk(pwOut, data)
				if err != nil {
					return
				}
			}
		}()

		return StreamResult{
			Stdin:    pwIn,
			Stdout:   prOut,
			Stderr:   io.NopCloser(bytes.NewReader(nil)),
			waitFunc: func() error { return nil },
		}, nil
	}

	const n = 200
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			data := []byte(fmt.Sprintf("msg-%d", i))
			out, err := SharedRunner.Run(Input{
				Reader:  bytes.NewBuffer(data),
				Options: StreamOptions{},
				Command: []string{"fake"},
			})
			if err != nil {
				errs <- err
				return
			}
			if !bytes.Equal(out, data) {
				errs <- fmt.Errorf("mismatch %d: got %q want %q", i, out, data)
				return
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		assert.NoError(t, e, "Concurrent Run failed")
	}
}
