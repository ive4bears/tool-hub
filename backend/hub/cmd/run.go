package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Options holds options for running a command.
type Options struct {
	Cwd     string
	Env     map[string]string
	Stdin   io.Reader
	Timeout time.Duration
}

// Result holds the result of a command execution.
type Result struct {
	Stdout   []byte
	Stderr   []byte
	Duration time.Duration
}

// Run executes a command-line tool in the specified working directory and returns its standard output, standard error, and any execution error.
func Run(ctx context.Context, options Options, command ...string) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = options.Cwd

	if options.Env != nil {
		env := os.Environ()
		for k, v := range options.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	if options.Stdin != nil {
		cmd.Stdin = options.Stdin
	}

	start := time.Now()
	var stdout, stderr []byte
	var err error
	stdout, err = cmd.Output()
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = exitErr.Stderr
		if ctx.Err() != nil {
			err = ctx.Err()
		}
	}
	return Result{
		Stdout:   stdout,
		Stderr:   stderr,
		Duration: time.Since(start),
	}, err
}

// StreamOptions holds options for streaming command execution.
// Stdin is exposed in StreamResult
type StreamOptions struct {
	Cwd     string
	Env     map[string]string
	Shell   string
	Timeout time.Duration
}

// Key generates a unique key for the StreamOptions, useful for identification.
func (o *StreamOptions) Key() string {
	var env string
	if o.Env != nil {
		keys := make([]string, 0, len(o.Env))
		for k := range o.Env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%s", k, o.Env[k]))
		}
		env = strings.Join(parts, ";")
	}
	return fmt.Sprintf("cwd=%s,shell=%s,env=%v", o.Cwd, o.Shell, env)
}

// StreamResult holds the output streams of a command execution.
type StreamResult struct {
	Stdin    io.WriteCloser
	Stdout   io.ReadCloser
	Stderr   io.ReadCloser
	waitFunc func() error
}

// Wait waits for the command to exit.
func (r *StreamResult) Wait() error {
	if r.waitFunc != nil {
		return r.waitFunc()
	}
	return nil
}

// RunStream executes a command-line tool with streaming stdin/stdout/stderr, supports options and timeout.
// Call res.Wait() to wait for the command to exit.
func RunStream(ctx context.Context, options StreamOptions, command ...string) (res StreamResult, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var cancel context.CancelFunc
	if options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
	}
	var cmd *exec.Cmd
	if options.Shell != "" {
		shellCmd := strings.Join(command, " ")
		cmd = exec.CommandContext(ctx, options.Shell, "-c", shellCmd)
	} else {
		cmd = exec.CommandContext(ctx, command[0], command[1:]...)
	}
	cmd.Dir = options.Cwd

	if options.Env != nil {
		env := os.Environ()
		for k, v := range options.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	pr, pw := io.Pipe()
	cmd.Stdin = pr
	res.Stdin = pw

	res.Stdout, err = cmd.StdoutPipe()
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return res, err
	}

	res.Stderr, err = cmd.StderrPipe()
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return res, err
	}

	if err = cmd.Start(); err != nil {
		if cancel != nil {
			cancel()
		}
		return res, err
	}

	res.waitFunc = func() error {
		err := cmd.Wait()
		if cancel != nil {
			cancel()
		}
		if err != nil && ctx.Err() != nil {
			err = ctx.Err()
		}
		return err
	}
	return res, nil
}
