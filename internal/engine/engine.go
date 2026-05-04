package engine

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/aspiand/zero-tunnel/internal/provider"
	"github.com/aspiand/zero-tunnel/internal/watcher"
	"github.com/aspiand/zero-tunnel/pkg/models"
	"github.com/docker/docker/api/types/events"
)

type Engine struct {
	watcher  *watcher.Watcher
	provider *provider.CloudflareProvider
	interval time.Duration

	mu             sync.Mutex
	routes         map[string]models.Route
	managedDomains map[string]bool
	syncChan       chan struct{}
}

func New(w *watcher.Watcher, p *provider.CloudflareProvider, interval time.Duration) *Engine {
	return &Engine{
		watcher:        w,
		provider:       p,
		interval:       interval,
		routes:         make(map[string]models.Route),
		managedDomains: make(map[string]bool),
		syncChan:       make(chan struct{}, 1),
	}
}

func (e *Engine) Run(ctx context.Context) error {
	slog.Info("starting zero-tunnel engine")

	go e.syncWorker(ctx)

	if err := e.initialSync(ctx); err != nil {
		slog.Error("initial sync failed", "error", err)
	}

	msgChan, errChan := e.watcher.WatchEvents(ctx)

	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case msg := <-msgChan:
			e.handleEvent(ctx, msg)
			e.triggerSync()

		case err := <-errChan:
			if err != nil {
				return err
			}

		case <-ticker.C:
			slog.Debug("periodic reconciliation triggered")
			e.triggerSync()

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (e *Engine) syncWorker(ctx context.Context) {
	for {
		select {
		case <-e.syncChan:
			time.Sleep(1 * time.Second)
			select {
			case <-e.syncChan:
			default:
			}

			if err := e.doSync(ctx); err != nil {
				slog.Error("sync failed", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (e *Engine) handleEvent(ctx context.Context, msg events.Message) {
	id := msg.Actor.ID
	action := msg.Action

	switch action {
	case "start":
		if route, ok := e.watcher.GetRoute(ctx, id); ok {
			slog.Info("route active", "container", route.ContainerName, "hostname", route.Hostname())
			e.mu.Lock()
			e.routes[id] = route
			e.managedDomains[route.Domain] = true
			e.mu.Unlock()
		}
	case "die", "stop":
		e.mu.Lock()
		route, exists := e.routes[id]
		if !exists {
			e.mu.Unlock()
			return // Already handled by 'die' or 'stop'
		}

		if route.Ephemeral {
			slog.Info("ephemeral route stopping", "container", route.ContainerName, "hostname", route.Hostname())
			delete(e.routes, id)
		} else {
			slog.Info("persistent route stopping", "container", route.ContainerName, "hostname", route.Hostname())
			route.Running = false
			e.routes[id] = route
		}
		e.mu.Unlock()

	case "destroy":
		e.mu.Lock()
		if r, ok := e.routes[id]; ok {
			slog.Info("route destroyed", "container", r.ContainerName, "hostname", r.Hostname())
			delete(e.routes, id)
		}
		e.mu.Unlock()
	}
}

func (e *Engine) initialSync(ctx context.Context) error {
	routes, err := e.watcher.ListRoutes(ctx, true)
	if err != nil {
		return err
	}

	e.mu.Lock()
	for _, r := range routes {
		if r.Running || !r.Ephemeral {
			e.routes[r.ID] = r
			e.managedDomains[r.Domain] = true
		}
	}
	e.mu.Unlock()

	e.triggerSync()
	return nil
}

func (e *Engine) triggerSync() {
	select {
	case e.syncChan <- struct{}{}:
	default:
	}
}

func (e *Engine) doSync(ctx context.Context) error {
	e.mu.Lock()
	var routes []models.Route
	for _, r := range e.routes {
		routes = append(routes, r)
	}
	var domains []string
	for d := range e.managedDomains {
		domains = append(domains, d)
	}
	e.mu.Unlock()

	return e.provider.Sync(ctx, routes, domains)
}
