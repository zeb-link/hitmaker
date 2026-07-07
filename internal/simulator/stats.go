package simulator

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Phase string

const (
	PhaseStarting Phase = "starting"
	PhaseActive   Phase = "active"
	PhaseIdle     Phase = "idle"
	PhasePaused   Phase = "paused"
	PhaseStopped  Phase = "stopped"
)

type HitResult struct {
	Target    string        `json:"target"`
	WorkerID  int           `json:"workerId"`
	Status    int           `json:"status,omitempty"`
	Err       string        `json:"error,omitempty"`
	Latency   time.Duration `json:"latency"`
	At        time.Time     `json:"at"`
	Phase     Phase         `json:"phase"`
	Location  string        `json:"location,omitempty"`
	IP        string        `json:"ip,omitempty"`
	Applied   []string      `json:"applied,omitempty"`
	UserAgent string        `json:"userAgent,omitempty"`
}

type WorkerState struct {
	Target   string    `json:"target"`
	WorkerID int       `json:"workerId"`
	Phase    Phase     `json:"phase"`
	Rate     int       `json:"rate"`
	Until    time.Time `json:"until,omitempty"`
	Paused   bool      `json:"paused"`
}

type TargetStats struct {
	Target        string        `json:"target"`
	Hits          int64         `json:"hits"`
	Errors        int64         `json:"errors"`
	LastStatus    int           `json:"lastStatus,omitempty"`
	LastError     string        `json:"lastError,omitempty"`
	LastLatency   time.Duration `json:"lastLatency"`
	LastAt        time.Time     `json:"lastAt,omitempty"`
	CurrentRate   int           `json:"currentRate"`
	ActiveWorkers int           `json:"activeWorkers"`
	IdleWorkers   int           `json:"idleWorkers"`
	Paused        bool          `json:"paused"`
}

type Snapshot struct {
	StartedAt    time.Time     `json:"startedAt"`
	Uptime       time.Duration `json:"uptime"`
	TotalHits    int64         `json:"totalHits"`
	TotalErrors  int64         `json:"totalErrors"`
	WorkerCapHit bool          `json:"workerCapHit"`
	WorkerCount  int           `json:"workerCount"`
	Targets      []TargetStats `json:"targets"`
	Workers      []WorkerState `json:"workers"`
	Recent       []HitResult   `json:"recent"`
}

type Collector struct {
	mu           sync.RWMutex
	startedAt    time.Time
	targets      map[string]*TargetStats
	workers      map[string]WorkerState
	recent       []HitResult
	recentLimit  int
	workerCapHit bool
	workerCount  int
}

func NewCollector(targets []string, recentLimit int) *Collector {
	if recentLimit <= 0 {
		recentLimit = 64
	}
	stats := map[string]*TargetStats{}
	for _, target := range targets {
		stats[target] = &TargetStats{Target: target}
	}
	return &Collector{
		startedAt:   time.Now(),
		targets:     stats,
		workers:     map[string]WorkerState{},
		recentLimit: recentLimit,
	}
}

func (c *Collector) SetWorkerCapHit(hit bool, count int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.workerCapHit = hit
	c.workerCount = count
}

func (c *Collector) SetWorker(state WorkerState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.workers[workerKey(state.Target, state.WorkerID)] = state
	c.recomputeTargetLocked(state.Target)
}

func (c *Collector) Add(result HitResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	stats := c.targets[result.Target]
	if stats == nil {
		stats = &TargetStats{Target: result.Target}
		c.targets[result.Target] = stats
	}
	if result.Err != "" {
		stats.Errors++
		stats.LastError = result.Err
	} else {
		stats.Hits++
		stats.LastStatus = result.Status
		stats.LastError = ""
	}
	stats.LastLatency = result.Latency
	stats.LastAt = result.At
	c.recent = append([]HitResult{result}, c.recent...)
	if len(c.recent) > c.recentLimit {
		c.recent = c.recent[:c.recentLimit]
	}
}

func (c *Collector) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := Snapshot{
		StartedAt:    c.startedAt,
		Uptime:       time.Since(c.startedAt),
		WorkerCapHit: c.workerCapHit,
		WorkerCount:  c.workerCount,
		Recent:       append([]HitResult(nil), c.recent...),
	}
	for _, stats := range c.targets {
		cp := *stats
		out.Targets = append(out.Targets, cp)
		out.TotalHits += cp.Hits
		out.TotalErrors += cp.Errors
	}
	for _, worker := range c.workers {
		out.Workers = append(out.Workers, worker)
	}
	sort.Slice(out.Targets, func(i, j int) bool { return out.Targets[i].Target < out.Targets[j].Target })
	sort.Slice(out.Workers, func(i, j int) bool {
		if out.Workers[i].Target == out.Workers[j].Target {
			return out.Workers[i].WorkerID < out.Workers[j].WorkerID
		}
		return out.Workers[i].Target < out.Workers[j].Target
	})
	return out
}

func (c *Collector) recomputeTargetLocked(target string) {
	stats := c.targets[target]
	if stats == nil {
		return
	}
	stats.CurrentRate = 0
	stats.ActiveWorkers = 0
	stats.IdleWorkers = 0
	stats.Paused = false
	for _, worker := range c.workers {
		if worker.Target != target {
			continue
		}
		if worker.Paused || worker.Phase == PhasePaused {
			stats.Paused = true
			continue
		}
		switch worker.Phase {
		case PhaseActive:
			stats.ActiveWorkers++
			stats.CurrentRate += worker.Rate
		case PhaseIdle:
			stats.IdleWorkers++
		}
	}
}

func workerKey(target string, id int) string {
	return fmt.Sprintf("%s#%d", target, id)
}
