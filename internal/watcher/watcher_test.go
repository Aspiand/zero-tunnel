package watcher

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestParseLabels(t *testing.T) {
	w := &Watcher{
		defaultDomain: "default.com",
	}

	tests := []struct {
		name          string
		containerName string
		labels        map[string]string
		wantHostname  string
		wantEphemeral bool
		wantEnabled   bool
	}{
		{
			name:          "Basic valid route",
			containerName: "/my-app",
			labels: map[string]string{
				LabelEnable:    "true",
				LabelSubdomain: "hello",
				LabelDomain:    "custom.com",
			},
			wantHostname:  "hello.custom.com",
			wantEphemeral: true,
			wantEnabled:   true,
		},
		{
			name:          "Fallback to default domain",
			containerName: "/my-app",
			labels: map[string]string{
				LabelEnable:    "true",
				LabelSubdomain: "hello",
			},
			wantHostname:  "hello.default.com",
			wantEphemeral: true,
			wantEnabled:   true,
		},
		{
			name:          "Persistent route",
			containerName: "/db",
			labels: map[string]string{
				LabelEnable:    "true",
				LabelSubdomain: "db",
				LabelEphemeral: "false",
			},
			wantHostname:  "db.default.com",
			wantEphemeral: false,
			wantEnabled:   true,
		},
		{
			name:          "Disabled container",
			containerName: "/ignored",
			labels: map[string]string{
				LabelEnable: "false",
			},
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := container.Summary{
				ID:     "test-id",
				Names:  []string{tt.containerName},
				Labels: tt.labels,
			}

			route, ok := w.parseLabels(c.ID, tt.labels, tt.containerName)

			if !tt.wantEnabled {
				assert.False(t, ok)
				return
			}

			assert.True(t, ok)
			assert.Equal(t, tt.wantHostname, route.Hostname())
			assert.Equal(t, tt.wantEphemeral, route.Ephemeral)
		})
	}
}
