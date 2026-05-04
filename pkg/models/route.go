package models

import "fmt"

type Route struct {
	ID            string
	ContainerName string
	Subdomain     string
	Domain        string
	Port          int
	Name          string
	Scheme        string
	Path          string
	Ephemeral     bool
	Running       bool
}

func (r Route) Hostname() string {
	if r.Subdomain == "" {
		return r.Domain
	}
	return fmt.Sprintf("%s.%s", r.Subdomain, r.Domain)
}

func (r Route) ServiceURL() string {
	scheme := r.Scheme
	if scheme == "" {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, r.Name, r.Port)
}
