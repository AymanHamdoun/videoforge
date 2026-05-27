package jobs

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"videoforge/internal/ffmpeg"
)

// Job is the status record surfaced to the UI.
type Job struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`   // queued | processing | completed | failed | canceled
	Progress   float64   `json:"progress"` // 0-100
	Type       string    `json:"type"`     // convert | speed | trim | ...
	InputFile  string    `json:"inputFile"`
	OutputFile string    `json:"outputFile,omitempty"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Emitter pushes job events to the frontend (wired to Wails runtime.EventsEmit).
type Emitter func(event string, data ...interface{})

// Manager runs ffmpeg jobs in goroutines, capped by a semaphore, and reports
// progress through the Emitter. It is the desktop replacement for the old
// HTTP job queue + SSE stream.
type Manager struct {
	mu      sync.RWMutex
	jobs    map[string]*Job
	cancels map[string]context.CancelFunc
	sem     chan struct{}
	emit    Emitter
}

func NewManager(maxConcurrent int, emit Emitter) *Manager {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	if emit == nil {
		emit = func(string, ...interface{}) {}
	}
	return &Manager{
		jobs:    make(map[string]*Job),
		cancels: make(map[string]context.CancelFunc),
		sem:     make(chan struct{}, maxConcurrent),
		emit:    emit,
	}
}

// Start registers a job and runs the ffmpeg command asynchronously. It returns
// the job ID immediately; progress arrives via "job:progress" / "job:done".
func (m *Manager) Start(jobType, input, output string, args []string, duration float64) string {
	id := uuid.NewString()
	job := &Job{
		ID:         id,
		Status:     "queued",
		Type:       jobType,
		InputFile:  input,
		OutputFile: output,
		CreatedAt:  time.Now(),
	}
	m.mu.Lock()
	m.jobs[id] = job
	m.mu.Unlock()

	go m.run(job, args, duration)
	return id
}

func (m *Manager) run(job *Job, args []string, duration float64) {
	m.sem <- struct{}{}
	defer func() { <-m.sem }()

	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancels[job.ID] = cancel
	job.Status = "processing"
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		delete(m.cancels, job.ID)
		m.mu.Unlock()
	}()

	err := ffmpeg.Run(ctx, args, duration, func(p float64) {
		m.mu.Lock()
		job.Progress = p
		m.mu.Unlock()
		m.emit("job:progress", map[string]interface{}{"id": job.ID, "percent": p})
	})

	m.mu.Lock()
	switch {
	case ctx.Err() == context.Canceled:
		job.Status = "canceled"
	case err != nil:
		job.Status = "failed"
		job.Error = err.Error()
	default:
		job.Status = "completed"
		job.Progress = 100
	}
	result := *job
	m.mu.Unlock()

	m.emit("job:done", result)
}

// Get returns a snapshot of the job (zero value if unknown).
func (m *Manager) Get(id string) Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if j, ok := m.jobs[id]; ok {
		return *j
	}
	return Job{}
}

// Cancel aborts a running job's ffmpeg process.
func (m *Manager) Cancel(id string) {
	m.mu.RLock()
	cancel := m.cancels[id]
	m.mu.RUnlock()
	if cancel != nil {
		cancel()
	}
}
