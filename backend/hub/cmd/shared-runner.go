package cmd

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// Input represents the input data and options for a runner command.
type Input struct {
	Reader  io.Reader
	Options StreamOptions
	Command []string
}

type runner struct {
	key         string
	worker      *StreamResult
	queue       chan inputTask
	idleTimeout time.Duration
}

type result struct {
	out []byte
	err error
}

type inputTask struct {
	Input
	result chan result
}

// SharedRunner is the global manager instance for handling runner commands.
var SharedRunner = manager{cmds: make(map[string]*runner)}

type manager struct {
	cmds map[string]*runner
	lock sync.Mutex
}

// runStream is a package-level variable that points to the real RunStream function
// by default. Tests can replace this with a fake implementation to avoid spawning
// real processes.
var runStream = RunStream

// Run executes the given Input using a managed runner and returns the output and error.
func (m *manager) Run(input Input) (out []byte, err error) {
	m.lock.Lock()
	ctx := context.Background()
	key := input.Options.Key()
	r, ok := m.cmds[key]
	if !ok {
		r = &runner{
			key:         key,
			queue:       make(chan inputTask),
			idleTimeout: 30 * time.Second,
		}
		m.cmds[key] = r
		go r.loop(ctx)
	}
	m.lock.Unlock()

	task := inputTask{
		Input:  input,
		result: make(chan result, 1),
	}
	r.queue <- task
	res := <-task.result
	return res.out, res.err
}

func (m *manager) stopRunner(key string) {
	m.lock.Lock()
	r, ok := m.cmds[key]
	if !ok {
		m.lock.Unlock()
		return
	}
	delete(m.cmds, key)
	// log.Info().Str("key", key).Msg("Runner stopped, queue closed")
	m.lock.Unlock()

	go func(r *runner) {
		if r == nil {
			return
		}
		close(r.queue)
		if r.worker != nil {
			r.worker.Stdin.Close()
			if r.worker.Stdout != nil {
				r.worker.Stdout.Close()
			}
			if r.worker.Stderr != nil {
				r.worker.Stderr.Close()
			}
		}
	}(r)
}

func (r *runner) getWorker(ctx context.Context, task inputTask) (*StreamResult, error) {
	if r.worker == nil {
		worker, err := runStream(ctx, task.Options, task.Command...)
		if err != nil {
			return nil, err
		}
		r.worker = &worker
		go func() {
			err = r.worker.Wait()
			// log.Error().Err(err).Msg("Worker exited")
			r.worker = nil
		}()
	}
	return r.worker, nil
}

// 串行处理 inputTask，保证顺序和 stdin/stdout 串行
func (r *runner) loop(ctx context.Context) error {
	idleTimer := time.NewTimer(r.idleTimeout) // 创建一个闲置计时器
	defer idleTimer.Stop()
	for {
		select {
		case task, ok := <-r.queue:
			{
				if !ok {
					closePendingTasks(r.queue)
					return nil
				}
				worker, err := r.getWorker(ctx, task)
				if err != nil {
					task.result <- result{nil, fmt.Errorf("Failed to get worker: %w", err)}
					continue
				}
				data, err := io.ReadAll(task.Reader)
				if err != nil && err != io.EOF {
					task.result <- result{nil, fmt.Errorf("Failed to read input: %w", err)}
					continue
				}
				if len(data) == 0 {
					task.result <- result{nil, errors.New("No input data")}
					continue
				}
				err = writeChunk(worker.Stdin, data)
				if err != nil {
					task.result <- result{nil, fmt.Errorf("Failed to write to stdin: %w", err)}
					continue
				}
				out, err := readChunk(worker.Stdout)
				if err != nil {
					task.result <- result{nil, fmt.Errorf("Failed to read from stdout: %w", err)}
					continue
				}
				task.result <- result{out, nil}

				idleTimer.Reset(r.idleTimeout)

			}
		case <-idleTimer.C:
			{

				if r.key != "" {
					SharedRunner.stopRunner(r.key)
				}
				return nil
			}
		}
	}
}

func writeChunk(w io.Writer, data []byte) error {
	var lenBuf [4]byte
	n := uint32(len(data))
	binary.BigEndian.PutUint32(lenBuf[:], n)
	if _, err := w.Write(lenBuf[:]); err != nil {
		return err
	}
	if n > 0 {
		_, err := w.Write(data)
		return err
	}
	return nil
}

func readChunk(r io.Reader) (buf []byte, err error) {
	var lenBuf [4]byte
	_, err = io.ReadFull(r, lenBuf[:])
	if err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(lenBuf[:])
	buf = make([]byte, n)
	if n > 0 {
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func closePendingTasks(queue chan inputTask) {
	for {
		select {
		case task, ok := <-queue:
			if !ok {
				return
			}
			// 通知任务队列已关闭
			if task.result != nil {
				task.result <- result{nil, errors.New("Runner queue closed")}
			}
		default:
			return
		}
	}
}
