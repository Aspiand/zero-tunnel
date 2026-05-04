package provider

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aspiand/zero-tunnel/pkg/models"
	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zero_trust"
	"github.com/cloudflare/cloudflare-go/v6/zones"
)

const (
	ManagedComment = "managed-by:zero-tunnel"
)

type CloudflareProvider struct {
	client    *cloudflare.Client
	accountID string
	tunnelID  string
}

func New(apiToken, accountID, tunnelID string) *CloudflareProvider {
	client := cloudflare.NewClient(option.WithAPIToken(apiToken))
	return &CloudflareProvider{
		client:    client,
		accountID: accountID,
		tunnelID:  tunnelID,
	}
}

func (p *CloudflareProvider) Sync(ctx context.Context, routes []models.Route, managedDomains []string) error {
	slog.Info("syncing state with cloudflare", "routes", len(routes), "managed_zones", len(managedDomains))

	// 1. Update Tunnel Configuration (Ingress Rules)
	var ingress []zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress
	runningCount := 0
	for _, r := range routes {
		if !r.Running {
			continue
		}
		runningCount++
		hostname := r.Hostname()
		rule := zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{
			Hostname: cloudflare.F(hostname),
			Service:  cloudflare.F(r.ServiceURL()),
		}
		if r.Path != "" {
			rule.Path = cloudflare.F(r.Path)
		}
		ingress = append(ingress, rule)
	}

	ingress = append(ingress, zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{
		Service: cloudflare.F("http_status:404"),
	})

	_, err := p.client.ZeroTrust.Tunnels.Cloudflared.Configurations.Update(ctx, p.tunnelID, zero_trust.TunnelCloudflaredConfigurationUpdateParams{
		AccountID: cloudflare.F(p.accountID),
		Config: cloudflare.F(zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfig{
			Ingress: cloudflare.F(ingress),
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to update tunnel configuration: %w", err)
	}

	// 2. Manage DNS records
	for _, domain := range managedDomains {
		if err := p.syncZoneDNS(ctx, domain, routes); err != nil {
			slog.Error("failed to sync DNS for zone", "zone", domain, "error", err)
		}
	}

	return nil
}

func (p *CloudflareProvider) syncZoneDNS(ctx context.Context, domain string, activeRoutes []models.Route) error {
	z, err := p.client.Zones.List(ctx, zones.ZoneListParams{
		Name: cloudflare.F(domain),
	})
	if err != nil || len(z.Result) == 0 {
		return fmt.Errorf("failed to find zone for domain %s", domain)
	}
	zoneID := z.Result[0].ID

	// 1. Get ALL records in this zone (we filter by comment manually because API filter for comments is limited)
	existingRecords, err := p.client.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
	})
	if err != nil {
		return fmt.Errorf("failed to list records: %w", err)
	}

	// 2. Identify managed records and remove orphans
	desiredHostnames := make(map[string]bool)
	for _, r := range activeRoutes {
		if r.Domain == domain {
			desiredHostnames[r.Hostname()] = true
		}
	}

	for _, record := range existingRecords.Result {
		// Identify if this record is managed by us using the comment field
		if strings.Contains(record.Comment, ManagedComment) {
			if !desiredHostnames[record.Name] {
				slog.Info("[DNS] removing orphaned record", "hostname", record.Name)
				_, err := p.client.DNS.Records.Delete(ctx, record.ID, dns.RecordDeleteParams{
					ZoneID: cloudflare.F(zoneID),
				})
				if err != nil {
					slog.Error("failed to delete orphaned record", "hostname", record.Name, "error", err)
				}
			}
		}
	}

	// 3. Create or Update active records
	tunnelContent := fmt.Sprintf("%s.cfargotunnel.com", p.tunnelID)
	for _, r := range activeRoutes {
		if r.Domain != domain {
			continue
		}

		hostname := r.Hostname()
		var existing *dns.RecordResponse
		for _, rec := range existingRecords.Result {
			if rec.Name == hostname && strings.Contains(rec.Comment, ManagedComment) {
				existing = &rec
				break
			}
		}

		if existing != nil {
			if existing.Content == tunnelContent {
				continue // Already correct
			}
			slog.Info("[DNS] updating record", "hostname", hostname, "container", r.ContainerName)
			_, err = p.client.DNS.Records.Update(ctx, existing.ID, dns.RecordUpdateParams{
				ZoneID: cloudflare.F(zoneID),
				Body: dns.CNAMERecordParam{
					Name:    cloudflare.F(hostname),
					Type:    cloudflare.F(dns.CNAMERecordTypeCNAME),
					Content: cloudflare.F(tunnelContent),
					Proxied: cloudflare.F(true),
					TTL:     cloudflare.F(dns.TTL1),
					Comment: cloudflare.F(ManagedComment),
				},
			})
		} else {
			slog.Info("[DNS] creating record", "hostname", hostname, "container", r.ContainerName)
			_, err = p.client.DNS.Records.New(ctx, dns.RecordNewParams{
				ZoneID: cloudflare.F(zoneID),
				Body: dns.CNAMERecordParam{
					Name:    cloudflare.F(hostname),
					Type:    cloudflare.F(dns.CNAMERecordTypeCNAME),
					Content: cloudflare.F(tunnelContent),
					Proxied: cloudflare.F(true),
					TTL:     cloudflare.F(dns.TTL1),
					Comment: cloudflare.F(ManagedComment),
				},
			})
		}

		if err != nil {
			slog.Error("failed to sync DNS record", "hostname", hostname, "error", err)
		}
	}

	return nil
}
