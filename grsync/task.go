package grsync

import (
	"bufio"
	"context"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Task runs rsync and exposes transfer progress and output.
type Task struct {
	rsync *Rsync

	mtx   sync.RWMutex
	state State
	log   Log
}

// State describes rsync transfer progress.
type State struct {
	Remain   int     `json:"remain"`
	Total    int     `json:"total"`
	Speed    string  `json:"speed"`
	Progress float64 `json:"progress"`
}

// Log contains raw rsync standard output and error.
type Log struct {
	Stderr string `json:"stderr"`
	Stdout string `json:"stdout"`
}

// State returns a snapshot of current rsync progress.
func (t *Task) State() State {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return t.state
}

// Log returns a snapshot of rsync output.
func (t *Task) Log() Log {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return t.log
}

// Run starts rsync and waits for transfer and output processing to finish.
func (t *Task) Run() error {
	stderr, err := t.rsync.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()
	stdout, err := t.rsync.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	scanErrors := make(chan error, 2)
	go func() {
		defer wg.Done()
		scanErrors <- t.processStdout(stdout)
	}()
	go func() {
		defer wg.Done()
		scanErrors <- t.processStderr(stderr)
	}()

	runError := t.rsync.Run()
	wg.Wait()
	close(scanErrors)
	if runError != nil {
		return runError
	}
	for scanError := range scanErrors {
		if scanError != nil {
			return errors.Wrap(scanError, "read rsync output")
		}
	}
	return nil
}

// NewTask constructs an rsync task.
func NewTask(
	ctx context.Context,
	source string,
	destination string,
	rsyncOptions RsyncOptions,
) *Task {
	rsyncOptions.HumanReadable = true
	rsyncOptions.Partial = true
	rsyncOptions.Progress = true
	rsyncOptions.Archive = true
	return &Task{rsync: NewRsync(ctx, source, destination, rsyncOptions)}
}

func (t *Task) processStdout(stdout io.Reader) error {
	const maxPercent = float64(100)
	const minDivider = 1

	progressMatcher := newMatcher(`\(.+-chk=(\d+.\d+)`)
	speedMatcher := newMatcher(`(\d+\.\d+.{2}\/s)`)
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		t.mtx.Lock()
		if progressMatcher.match(line) {
			t.state.Remain, t.state.Total = getTaskProgress(progressMatcher.extract(line))
			copiedCount := float64(t.state.Total - t.state.Remain)
			t.state.Progress = copiedCount / math.Max(float64(t.state.Total), float64(minDivider)) * maxPercent
		}
		if speedMatcher.match(line) {
			t.state.Speed = getTaskSpeed(speedMatcher.extractAllStringSubmatch(line, 2))
		}
		t.log.Stdout += line + "\n"
		t.mtx.Unlock()
	}
	return scanner.Err()
}

func (t *Task) processStderr(stderr io.Reader) error {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		t.mtx.Lock()
		t.log.Stderr += scanner.Text() + "\n"
		t.mtx.Unlock()
	}
	return scanner.Err()
}

func getTaskProgress(remainTotal string) (int, int) {
	const expectedParts = 2
	parts := strings.Split(remainTotal, "/")
	if len(parts) < expectedParts {
		return 0, 0
	}
	remain, _ := strconv.Atoi(parts[0])
	total, _ := strconv.Atoi(parts[1])
	return remain, total
}

func getTaskSpeed(data [][]string) string {
	if len(data) < 2 || len(data[1]) < 2 {
		return ""
	}
	return data[1][1]
}
