package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute_Hostname(t *testing.T) {
	r := Route{
		Subdomain: "blog",
		Domain:    "example.com",
	}
	assert.Equal(t, "blog.example.com", r.Hostname())
}

func TestRoute_ServiceURL(t *testing.T) {
	tests := []struct {
		name     string
		route    Route
		expected string
	}{
		{
			name: "Default HTTP",
			route: Route{
				Name: "web",
				Port: 80,
			},
			expected: "http://web:80",
		},
		{
			name: "Custom Scheme and Port",
			route: Route{
				Name:   "api",
				Scheme: "https",
				Port:   443,
			},
			expected: "https://api:443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.route.ServiceURL())
		})
	}
}
