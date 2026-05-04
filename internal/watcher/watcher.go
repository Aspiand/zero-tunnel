package watcher

import (
	"context"
	"strconv"
	"strings"

	"github.com/aspiand/zero-tunnel/pkg/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

const (
	LabelEnable    = "zero-tunnel.enable"
	LabelSubdomain = "zero-tunnel.subdomain"
	LabelDomain    = "zero-tunnel.domain"
	LabelPort      = "zero-tunnel.port"
	LabelName      = "zero-tunnel.name"
	LabelScheme    = "zero-tunnel.scheme"
	LabelPath      = "zero-tunnel.path"
	LabelEphemeral = "zero-tunnel.ephemeral"
)

type Watcher struct {
	docker        *client.Client
	defaultDomain string
}

func New(defaultDomain string) (*Watcher, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Watcher{
		docker:        cli,
		defaultDomain: defaultDomain,
	}, nil
}

func (w *Watcher) ListRoutes(ctx context.Context, all bool) ([]models.Route, error) {
	containers, err := w.docker.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, err
	}

	var routes []models.Route
	for _, c := range containers {
		if route, ok := w.parseLabels(c.ID, c.Labels, c.Names[0]); ok {
			route.Running = c.State == "running"
			routes = append(routes, route)
		}
	}
	return routes, nil
}

func (w *Watcher) WatchEvents(ctx context.Context) (<-chan events.Message, <-chan error) {
	f := filters.NewArgs()
	f.Add("type", "container")
	f.Add("event", "start")
	f.Add("event", "die")
	f.Add("event", "stop")
	f.Add("event", "destroy")

	return w.docker.Events(ctx, events.ListOptions{Filters: f})
}

// GetRoute fetches a specific container and parses its labels into a Route.
func (w *Watcher) GetRoute(ctx context.Context, containerID string) (models.Route, bool) {
	c, err := w.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return models.Route{}, false
	}

	route, ok := w.parseLabels(c.ID, c.Config.Labels, c.Name)
	if ok {
		route.Running = c.State.Running
	}
	return route, ok
}

func (w *Watcher) parseLabels(id string, labels map[string]string, containerName string) (models.Route, bool) {
	if labels[LabelEnable] != "true" {
		return models.Route{}, false
	}

	domain := labels[LabelDomain]
	if domain == "" {
		domain = w.defaultDomain
	}

	port, _ := strconv.Atoi(labels[LabelPort])
	if port == 0 {
		port = 80 // default
	}

	name := labels[LabelName]
	cleanName := strings.TrimPrefix(containerName, "/")
	if name == "" {
		name = cleanName
	}

	ephemeral := true
	if val, ok := labels[LabelEphemeral]; ok {
		if b, err := strconv.ParseBool(val); err == nil {
			ephemeral = b
		}
	}

	return models.Route{
		ID:            id,
		ContainerName: cleanName,
		Subdomain:     labels[LabelSubdomain],
		Domain:        domain,
		Port:          port,
		Name:          name,
		Scheme:        labels[LabelScheme],
		Path:          labels[LabelPath],
		Ephemeral:     ephemeral,
	}, true
}
