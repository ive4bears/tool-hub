package cmd

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun_basic(t *testing.T) {
	result, err := Run(context.Background(), Options{}, "echo", "hello")
	assert.NoError(t, err, "Run failed")
	assert.Contains(t, string(result.Stdout), "hello")
	assert.Equal(t, "", string(result.Stderr))
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestRun_stdin(t *testing.T) {
	in := bytes.NewBufferString("foo\n")
	res, err := Run(context.Background(), Options{Stdin: in}, "cat")
	assert.NoError(t, err, "Run with stdin failed")
	assert.NoError(t, err, "Wait failed")
	assert.Equal(t, "foo\n", string(res.Stdout))
}

func TestRun_timeout(t *testing.T) {
	_, err := Run(context.Background(), Options{Timeout: 10 * time.Millisecond}, "sleep", "1")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestStream_basic(t *testing.T) {
	stream, err := RunStream(context.Background(), StreamOptions{}, "echo", "streaming")
	assert.NoError(t, err, "Stream failed")
	var out bytes.Buffer
	go func() {
		_, err = io.Copy(&out, stream.Stdout)
		assert.NoError(t, err, "Read Stdout failed")
	}()
	io.Copy(stream.Stdin, bytes.NewBufferString("Streaming"))
	stream.Stdin.Close()
	err = stream.Wait()
	assert.NoError(t, err, "Wait failed")
	assert.Contains(t, out.String(), "streaming")
}

func TestStream_stdin(t *testing.T) {
	in := bytes.NewBufferString("foo\n")
	stream, err := RunStream(context.Background(), StreamOptions{}, "cat")
	assert.NoError(t, err, "Stream with stdin failed")
	var out bytes.Buffer
	go func() {
		_, err = io.Copy(&out, stream.Stdout)
		assert.NoError(t, err, "Read Stdout failed")
	}()
	_, err = io.Copy(stream.Stdin, in)
	assert.NoError(t, err, "Write to Stdin failed")
	stream.Stdin.Close()
	err = stream.Wait()
	assert.NoError(t, err, "Wait failed")
	assert.Equal(t, "foo\n", out.String())
}

func TestStream_timeout(t *testing.T) {
	stream, err := RunStream(context.Background(), StreamOptions{Timeout: 10 * time.Millisecond}, "sleep", "1")
	assert.NoError(t, err)
	stream.Stdin.Close()
	err = stream.Wait()
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
