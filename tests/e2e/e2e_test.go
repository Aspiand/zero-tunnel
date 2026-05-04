package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aspiand/zero-tunnel/internal/engine"
	"github.com/aspiand/zero-tunnel/internal/provider"
	"github.com/aspiand/zero-tunnel/internal/watcher"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCF struct {
	mu           sync.Mutex
	ingressRules []interface{}
	dnsRecords   map[string]string // hostname -> recordID
	dnsDeleted   []string          // recordIDs
}

func TestE2ELifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	if os.Getenv("DOCKER_HOST") == "" && !isDockerRunning() {
		t.Skip("Docker daemon not found, skipping E2E test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Setup Mock Cloudflare
	mcf := &mockCF{
		dnsRecords: make(map[string]string),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mcf.mu.Lock()
		defer mcf.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		// Mock Tunnel Configurations Update
		if r.Method == "PUT" && r.URL.Path == "/accounts/acc-123/cfd_tunnel/tun-123/configurations" {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			config := body["config"].(map[string]interface{})
			mcf.ingressRules = config["ingress"].([]interface{})
			w.Write([]byte(`{"success":true,"result":{}}`))
			return
		}

		// Mock Zones List
		if r.Method == "GET" && r.URL.Path == "/zones" {
			w.Write([]byte(`{"success":true,"result":[{"id":"zone-123","name":"example.com"}]}`))
			return
		}

		// Mock DNS Records List
		if r.Method == "GET" && r.URL.Path == "/zones/zone-123/dns_records" {
			var result []map[string]interface{}
			for host, id := range mcf.dnsRecords {
				result = append(result, map[string]interface{}{
					"id":      id,
					"name":    host,
					"type":    "CNAME",
					"content": "tun-123.cfargotunnel.com",
					"comment": "managed-by:zero-tunnel",
				})
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "result": result})
			return
		}

		// Mock DNS Record Creation
		if r.Method == "POST" && r.URL.Path == "/zones/zone-123/dns_records" {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			hostname := body["name"].(string)
			id := fmt.Sprintf("rec-%s", hostname)
			mcf.dnsRecords[hostname] = id
			w.Write([]byte(fmt.Sprintf(`{"success":true,"result":{"id":"%s"}}`, id)))
			return
		}

		// Mock DNS Record Deletion
		if r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/zones/zone-123/dns_records/") {
			id := strings.TrimPrefix(r.URL.Path, "/zones/zone-123/dns_records/")
			mcf.dnsDeleted = append(mcf.dnsDeleted, id)
			for host, rid := range mcf.dnsRecords {
				if rid == id {
					delete(mcf.dnsRecords, host)
					break
				}
			}
			w.Write([]byte(`{"success":true,"result":{"id":"` + id + `"}}`))
			return
		}

		w.WriteHeader(404)
	}))
	defer server.Close()

	// 2. Initialize zero-tunnel
	w, err := watcher.New("example.com")
	require.NoError(t, err)

	p := provider.New("fake-token", "acc-123", "tun-123", option.WithBaseURL(server.URL+"/"))
	eng := engine.New(w, p, 5*time.Minute)

	engineCtx, engineCancel := context.WithCancel(ctx)
	defer engineCancel()
	go eng.Run(engineCtx)

	// 3. Create Docker Container
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	containerName := "zero-tunnel-e2e-test"
	// Cleanup if exists
	cli.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true})

	reader, err := cli.ImagePull(ctx, "docker.io/library/busybox:latest", image.PullOptions{})
	require.NoError(t, err)
	defer reader.Close()
	// Wait for image to be pulled
	io.Copy(io.Discard, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "busybox",
		Cmd:   []string{"sleep", "300"},
		Labels: map[string]string{
			"zero-tunnel.enable":    "true",
			"zero-tunnel.subdomain": "e2e",
		},
	}, nil, nil, nil, containerName)
	require.NoError(t, err)

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	require.NoError(t, err)
	defer cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true})

	// 4. Wait and Verify Ingress & DNS
	t.Log("Waiting for sync...")
	assert.Eventually(t, func() bool {
		mcf.mu.Lock()
		defer mcf.mu.Unlock()

		// Verify Ingress
		foundIngress := false
		for _, rule := range mcf.ingressRules {
			r := rule.(map[string]interface{})
			if r["hostname"] == "e2e.example.com" {
				foundIngress = true
			}
		}

		// Verify DNS
		_, foundDNS := mcf.dnsRecords["e2e.example.com"]

		return foundIngress && foundDNS
	}, 10*time.Second, 1*time.Second, "Expected ingress rule and DNS record to be created")

	// 5. Stop Container
	t.Log("Stopping container...")
	err = cli.ContainerStop(ctx, resp.ID, container.StopOptions{})
	require.NoError(t, err)

	// 6. Wait and Verify Cleanup
	t.Log("Waiting for cleanup...")
	assert.Eventually(t, func() bool {
		mcf.mu.Lock()
		defer mcf.mu.Unlock()

		// Verify Ingress removed (should only have 404 rule left)
		foundIngress := false
		for _, rule := range mcf.ingressRules {
			r := rule.(map[string]interface{})
			if r["hostname"] == "e2e.example.com" {
				foundIngress = true
			}
		}

		// Verify DNS removed
		_, foundDNS := mcf.dnsRecords["e2e.example.com"]

		return !foundIngress && !foundDNS
	}, 10*time.Second, 1*time.Second, "Expected ingress rule and DNS record to be removed")
}

func isDockerRunning() bool {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false
	}
	_, err = cli.Ping(context.Background())
	return err == nil
}
