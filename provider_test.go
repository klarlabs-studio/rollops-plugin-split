package split

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.klarlabs.de/rollops/pkg/plugin"
)

func TestApplyFlag_PutsDefaultRuleBuckets(t *testing.T) {
	var path, method, auth string
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, method, auth = r.URL.Path, r.Method, r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := Provider{BaseURL: srv.URL, Token: "key", HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "checkout", Environment: "Production", Percentage: 30}); err != nil {
		t.Fatalf("ApplyFlag: %v", err)
	}
	if method != http.MethodPut {
		t.Errorf("method = %s, want PUT", method)
	}
	if !strings.HasSuffix(path, "/internal/api/v2/splits/checkout/environments/Production") {
		t.Errorf("wrong path: %s", path)
	}
	if auth != "Bearer key" {
		t.Errorf("auth = %q", auth)
	}
	rule, _ := body["defaultRule"].([]any)
	on, _ := rule[0].(map[string]any)
	off, _ := rule[1].(map[string]any)
	if on["size"].(float64) != 30 || off["size"].(float64) != 70 {
		t.Errorf("buckets = %v/%v, want 30/70", on["size"], off["size"])
	}
}

func TestApplyFlag_DisabledServesOff(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := Provider{BaseURL: srv.URL, Token: "key", HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "f", Environment: "Production", Percentage: 80, Disabled: true}); err != nil {
		t.Fatalf("ApplyFlag: %v", err)
	}
	rule, _ := body["defaultRule"].([]any)
	on, _ := rule[0].(map[string]any)
	off, _ := rule[1].(map[string]any)
	if on["size"].(float64) != 0 || off["size"].(float64) != 100 {
		t.Errorf("disabled buckets = %v/%v, want 0/100", on["size"], off["size"])
	}
}

func TestApplyFlag_ServerErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(403) }))
	defer srv.Close()
	p := Provider{BaseURL: srv.URL, Token: "key", HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "f", Environment: "p"}); err == nil {
		t.Fatal("403 must error")
	}
}

func TestApplyFlag_RequiresToken(t *testing.T) {
	p := Provider{BaseURL: "http://x"}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "f", Environment: "p"}); err == nil {
		t.Fatal("missing token must error")
	}
}
