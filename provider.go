// Package split is a Rollops feature-flag provider plugin backed by Split's
// (Harness FME) Admin API. It drives a split's default rule treatment buckets so
// the on/off percentage split matches a rollout's progressive steps, so a Split
// feature flag tracks a Rollops canary in lockstep.
package split

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.klarlabs.de/rollops/pkg/plugin"
)

// Provider talks to Split's Admin API. BaseURL and Token come from the plugin's
// environment (see Config); Environment is supplied per call by Rollops as the
// Split environment name (e.g. "Production"). The split is treated as a boolean
// on/off split.
type Provider struct {
	BaseURL string // e.g. https://api.split.io
	Token   string // Admin API key (Authorization: Bearer <key>)
	HTTP    *http.Client
}

func (p Provider) client() *http.Client {
	if p.HTTP != nil {
		return p.HTTP
	}
	return http.DefaultClient
}

// ApplyFlag PUTs the split's definition in the environment with a default rule
// whose treatment buckets are the on/off percentage split. When disabled, the
// on bucket is 0 so every key serves off.
func (p Provider) ApplyFlag(ctx context.Context, c plugin.FlagChange) error {
	if p.Token == "" {
		return fmt.Errorf("split: SPLIT_TOKEN is required")
	}
	on := c.Percentage
	if c.Disabled {
		on = 0
	}
	def := map[string]any{
		"treatments":        []any{map[string]any{"name": "on"}, map[string]any{"name": "off"}},
		"defaultTreatment":  "off",
		"baselineTreatment": "off",
		"rules":             []any{},
		"defaultRule": []any{
			map[string]any{"treatment": "on", "size": on},
			map[string]any{"treatment": "off", "size": 100 - on},
		},
		"comment": "rollops canary",
	}
	u := fmt.Sprintf("%s/internal/api/v2/splits/%s/environments/%s",
		p.BaseURL, url.PathEscape(c.Flag), url.PathEscape(c.Environment))
	if err := p.put(ctx, u, def); err != nil {
		return fmt.Errorf("split: update split %q in %q: %w", c.Flag, c.Environment, err)
	}
	return nil
}

func (p Provider) put(ctx context.Context, u string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
