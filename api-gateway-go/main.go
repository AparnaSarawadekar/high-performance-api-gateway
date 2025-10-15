package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/AparnaSarawadekar/high-performance-api-gateway/api-gateway-go/internal/cache"
	"github.com/AparnaSarawadekar/high-performance-api-gateway/api-gateway-go/internal/ratelimit"
)

// Health response structure
type healthResponse struct {
	Ok      bool              `json:"ok"`
	Service string            `json:"service"`
	Uptime  int64             `json:"uptime_ms"`
	Targets map[string]string `json:"targets,omitempty"`
}

var startTime = time.Now()

func getenv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func mustParse(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("invalid URL %q: %v", raw, err)
	}
	return u
}

// Reverse proxy with path rewrite
func newPathProxy(base *url.URL, backendPath string) http.HandlerFunc {
	rp := httputil.NewSingleHostReverseProxy(base)
	orig := rp.Director
	rp.Director = func(r *http.Request) {
		orig(r)
		r.URL.Path = backendPath
		r.URL.RawPath = backendPath
		r.Host = base.Host
	}
	return rp.ServeHTTP
}

func main() {
	pyBase := mustParse(getenv("PY_SERVICE_URL", "http://service-python:8001"))
	nodeBase := mustParse(getenv("NODE_SERVICE_URL", "http://service-node:8002"))
	port := getenv("PORT", "8080")

	mux := http.NewServeMux()

	// Health endpoint (bypassed by cache & limiter)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(healthResponse{
			Ok:      true,
			Service: "api-gateway-go",
			Uptime:  time.Since(startTime).Milliseconds(),
			Targets: map[string]string{
				"python": pyBase.String(),
				"node":   nodeBase.String(),
			},
		})
	})

	// Limited echo endpoint (used in Step 10 verification)
	mux.HandleFunc("/limited", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	// Simple GET endpoint to demo caching (idempotent)
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(120 * time.Millisecond) // simulate work
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"path":"/slow"}`))
	})

	// Inference routes (rate-limited; cacheable later if you add GETs)
	mux.HandleFunc("/infer/python", newPathProxy(pyBase, "/infer"))
	mux.HandleFunc("/infer/node", newPathProxy(nodeBase, "/infer"))

	// Fallback root
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "API Gateway MVP: use /healthz, /limited, /slow, /infer/python, /infer/node\n")
	})

	// ---- Middlewares: rate-limit -> cache -> mux ----
	rl := ratelimit.NewManagerFromEnv()
	cached := cache.NewMiddleware(cache.NewFromEnv(), "/healthz") // bypass health
	handler := rl.Middleware(cached.Handler(mux))
	// -------------------------------------------------

	addr := ":" + port
	log.Printf("Gateway listening on %s", port)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
