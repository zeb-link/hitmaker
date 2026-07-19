package simulator

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/zeb-link/hitmaker/v2/internal/config"
	"github.com/zeb-link/hitmaker/v2/internal/identity"
	"github.com/zeb-link/hitmaker/v2/internal/proxy"
	"github.com/zeb-link/hitmaker/v2/internal/urlparams"
)

const DefaultMaxWorkers = 128

type Options struct {
	Config     config.Config
	Targets    []string
	MaxWorkers int
	// Follow controls whether the HTTP client follows 3xx redirects. It is off
	// by default: the tool tests redirect services, so the redirect's own status
	// (e.g. 302) is the signal, and not following avoids hammering the real
	// destination sites with synthetic bot traffic.
	Follow bool
	// Seed makes a run reproducible. Zero means seed from the clock.
	Seed int64
}

type Runner struct {
	cfg        config.Config
	targets    []string
	transport  proxy.Transport
	client     *http.Client
	collector  *Collector
	idgen      *identity.Generator
	maxWorkers int
	baseSeed   int64
	personas   map[string]persona

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu     sync.RWMutex
	paused map[string]bool
}

func New(ctx context.Context, opts Options) (*Runner, error) {
	if err := opts.Config.Validate(); err != nil {
		return nil, err
	}
	targets, err := NormalizeTargets(opts.Targets)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no targets supplied")
	}
	maxWorkers := opts.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = DefaultMaxWorkers
	}
	transport, err := proxy.DefaultRegistry().BuildTransportForTargets(ctx, opts.Config.Origin, targets)
	if err != nil {
		return nil, err
	}
	botFilter, err := opts.Config.BotFilter()
	if err != nil {
		return nil, err
	}
	idgen := identity.New(opts.Seed, 4096)
	idgen.UseBots(botFilter)
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(opts.Config.Traffic.TimeoutMs) * time.Millisecond,
	}
	if !opts.Follow {
		client.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	// Personas are drawn from a stable base seed so a fixed run seed reproduces
	// them exactly; a zero seed varies them per run off the clock.
	personaBaseSeed := opts.Seed
	if personaBaseSeed == 0 {
		personaBaseSeed = time.Now().UnixNano()
	}
	runCtx, cancel := context.WithCancel(ctx)
	r := &Runner{
		cfg:        opts.Config,
		targets:    targets,
		transport:  transport,
		client:     client,
		collector:  NewCollector(targets, 80),
		idgen:      idgen,
		maxWorkers: maxWorkers,
		baseSeed:   opts.Seed,
		personas:   buildPersonas(opts.Config, targets, personaBaseSeed),
		ctx:        runCtx,
		cancel:     cancel,
		paused:     map[string]bool{},
	}
	return r, nil
}

func NormalizeTargets(raw []string) ([]string, error) {
	seen := map[string]struct{}{}
	out := []string{}
	for _, target := range raw {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		parsed, err := url.Parse(target)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, fmt.Errorf("invalid target URL %q", target)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return nil, fmt.Errorf("target must be http or https: %q", target)
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		out = append(out, target)
	}
	return out, nil
}

func (r *Runner) Start() {
	requested := len(r.targets) * r.cfg.Traffic.Concurrent
	workers := requested
	if workers > r.maxWorkers {
		workers = r.maxWorkers
	}
	r.collector.SetWorkerCapHit(requested > workers, workers)
	started := 0
	for _, target := range r.targets {
		for i := 0; i < r.cfg.Traffic.Concurrent; i++ {
			if started >= workers {
				return
			}
			started++
			workerID := i + 1
			seed := time.Now().UnixNano() + int64(started)
			if r.baseSeed != 0 {
				seed = r.baseSeed + int64(started)
			}
			r.wg.Add(1)
			go r.worker(target, workerID, seed)
			time.Sleep(80 * time.Millisecond)
		}
	}
}

func (r *Runner) Stop() {
	r.cancel()
	r.transport.CloseIdleConnections()
}

func (r *Runner) Wait() {
	r.wg.Wait()
}

// StopAndWait cancels all workers and waits up to grace for them to return. It
// gives up after grace even if a worker is still stuck (e.g. an in-flight
// proxied connection that ignores cancellation), so a caller never hangs on
// shutdown — the process exits and abandons the stuck goroutine.
func (r *Runner) StopAndWait(grace time.Duration) {
	r.Stop()
	if grace <= 0 {
		r.wg.Wait()
		return
	}
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(grace):
	}
}

func (r *Runner) Snapshot() Snapshot {
	return r.collector.Snapshot()
}

func (r *Runner) Probe(target string) HitResult {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.doHit(target, 0, rng)
	snap := r.collector.Snapshot()
	if len(snap.Recent) == 0 {
		return HitResult{Target: target, Err: "probe produced no result", At: time.Now()}
	}
	return snap.Recent[0]
}

func (r *Runner) TogglePause(target string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paused[target] = !r.paused[target]
	return r.paused[target]
}

func (r *Runner) SetPaused(target string, paused bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paused[target] = paused
}

func (r *Runner) isPaused(target string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.paused[target]
}

func (r *Runner) worker(target string, workerID int, seed int64) {
	defer r.wg.Done()
	rng := rand.New(rand.NewSource(seed))
	p := r.personas[target]
	first := true
	for {
		if err := r.ctx.Err(); err != nil {
			r.collector.SetWorker(WorkerState{Target: target, WorkerID: workerID, Phase: PhaseStopped})
			return
		}
		if r.isPaused(target) {
			r.collector.SetWorker(WorkerState{Target: target, WorkerID: workerID, Phase: PhasePaused, Paused: true})
			if !sleepContext(r.ctx, 250*time.Millisecond) {
				return // context cancelled
			}
			continue // re-check pause state
		}
		activeMinutes := randRange(rng, r.cfg.Schedule.MinActive, r.cfg.Schedule.MaxActive)
		if activeMinutes == 0 {
			activeMinutes = 1
		}
		// Desync the herd: cut the very first burst short by a random amount so
		// links reach their idle roll at staggered times rather than all at once.
		// Traffic still starts immediately.
		if first && p.desync && activeMinutes > 1 {
			activeMinutes = 1 + rng.Intn(activeMinutes)
		}
		first = false
		r.activePhase(target, workerID, rng, activeMinutes, p)
		if rng.Float64() < p.idleOdds {
			idleMinutes := randRange(rng, r.cfg.Schedule.MinIdle, r.cfg.Schedule.MaxIdle)
			r.idlePhase(target, workerID, idleMinutes)
		}
	}
}

func (r *Runner) activePhase(target string, workerID int, rng *rand.Rand, minutes int, p persona) {
	// A link's energy scales its chosen rate: busy (and viral) links push toward
	// and past the top of the range; quiet links stay light.
	rate := scaleRate(randRange(rng, r.cfg.Traffic.MinPerMin, r.cfg.Traffic.MaxPerMin), p.energy)
	if rate <= 0 {
		rate = 1
	}
	until := time.Now().Add(time.Duration(minutes) * time.Minute)
	r.collector.SetWorker(WorkerState{Target: target, WorkerID: workerID, Phase: PhaseActive, Rate: rate, Until: until})
	for time.Now().Before(until) {
		if r.ctx.Err() != nil || r.isPaused(target) {
			return
		}
		start := time.Now()
		r.doHit(target, workerID, rng)
		elapsed := time.Since(start)
		base := time.Minute / time.Duration(rate)
		jitter := time.Duration(float64(base) * (rng.Float64()*0.2 - 0.1))
		delay := base + jitter - elapsed
		if delay < 0 {
			delay = 0
		}
		if !sleepContext(r.ctx, delay) {
			return
		}
	}
}

func (r *Runner) idlePhase(target string, workerID int, minutes int) {
	if minutes <= 0 {
		return
	}
	until := time.Now().Add(time.Duration(minutes) * time.Minute)
	r.collector.SetWorker(WorkerState{Target: target, WorkerID: workerID, Phase: PhaseIdle, Until: until})
	_ = sleepContext(r.ctx, time.Duration(minutes)*time.Minute)
}

func (r *Runner) doHit(target string, workerID int, rng *rand.Rand) {
	// Each link carries its own desktop/mobile skew, so the device breakdown
	// varies across links instead of every link matching the global ratio.
	deviceRatio := r.cfg.Requests.DeviceRatio
	if p, ok := r.personas[target]; ok {
		deviceRatio = p.deviceRatio
	}
	ident := r.idgen.Next(deviceRatio, r.cfg.Requests.UnknownRatio, r.cfg.Requests.UniqueIPProb)
	applied, err := urlparams.Apply(target, r.cfg.Requests.URLParams, rng)
	if err != nil {
		r.collector.Add(HitResult{Target: target, WorkerID: workerID, Err: err.Error(), At: time.Now(), Phase: PhaseActive})
		return
	}
	req, err := http.NewRequestWithContext(r.ctx, strings.ToUpper(r.cfg.Requests.Method), applied.URL, nil)
	if err != nil {
		r.collector.Add(HitResult{Target: target, WorkerID: workerID, Err: err.Error(), At: time.Now(), Phase: PhaseActive})
		return
	}
	if req.Method == http.MethodPost {
		req.Body = io.NopCloser(strings.NewReader(""))
		req.ContentLength = 0
	}
	identity.ApplyBrowserHeaders(req.Header, ident)
	if r.usesVercelGeoHeaders(req.URL) {
		identity.ApplyVercelGeoHeaders(req.Header, ident)
	}

	start := time.Now()
	res, err := r.client.Do(req)
	latency := time.Since(start)
	result := HitResult{
		Target:    target,
		WorkerID:  workerID,
		Latency:   latency,
		At:        time.Now(),
		Phase:     PhaseActive,
		IP:        ident.IP,
		Location:  humanCity(ident.Location.City) + ", " + ident.Location.Country,
		Applied:   applied.Names,
		UserAgent: ident.UserAgent,
	}
	if err != nil {
		result.Err = err.Error()
		r.collector.Add(result)
		return
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	result.Status = res.StatusCode
	if res.StatusCode >= 500 {
		result.Err = res.Status
	}
	r.collector.Add(result)
}

func (r *Runner) usesVercelGeoHeaders(u *url.URL) bool {
	switch r.cfg.Origin.Mode {
	case config.ModeVercel:
		return true
	case config.ModeAuto:
		return !proxy.UsesPaidProxy(u)
	default:
		return false
	}
}

// humanCity decodes the URL-escaped city stored for geo headers (e.g.
// "S%C3%A3o%20Paulo") into a readable form for display output.
func humanCity(city string) string {
	if decoded, err := url.QueryUnescape(city); err == nil {
		return decoded
	}
	return city
}

func randRange(rng *rand.Rand, min, max int) int {
	if max < min {
		max = min
	}
	if max == min {
		return min
	}
	return min + rng.Intn(max-min+1)
}

func sleepContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
