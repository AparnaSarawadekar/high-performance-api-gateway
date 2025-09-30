package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

type jsonResponse map[string]any

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func proxyTo(targetBase, overridePath string) *httputil.ReverseProxy {
	u, err := url.Parse(targetBase)
	if err != nil {
		log.Fatalf("invalid target url %q: %v", targetBase, err)
	}
	p := httputil.NewSingleHostReverseProxy(u)

	orig := p.Director
	p.Director = func(req *http.Request) {
		orig(req)
		if overridePath != "" {
			req.URL.Path = overridePath
		}
		// keeping the original query string
	}

	// adding a simple error handler to return JSON
	p.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, e error) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(rw).Encode(jsonResponse{
			"ok":    false,
			"error": e.Error(),
		})
	}
	return p
}

func jsonOK(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func main() {
	// default URLs match docker-compose service names/ports
	pyURL := env("PY_SERVICE_URL", "http://service-python:8001")
	nodeURL := env("NODE_SERVICE_URL", "http://service-node:8002")

	mux := http.NewServeMux()

	// health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		jsonOK(w, jsonResponse{"ok": true, "service": "api-gateway-go"})
	})

	// route: /infer  -> python service /infer
	pyProxy := proxyTo(pyURL, "/infer")
	mux.Handle("/infer", pyProxy)

	// (placeholder) route: /metrics -> node service /metrics
	nodeProxy := proxyTo(nodeURL, "/metrics")
	mux.Handle("/metrics", nodeProxy)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           logging(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("api-gateway-go listening on :8080")
	log.Fatal(srv.ListenAndServe())
}

// simple JSON access log
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &wrap{ResponseWriter: w, status: 200}
		next.ServeHTTP(ww, r)
		log.Printf(`{"ts":"%s","method":"%s","path":"%s","status":%d,"dur_ms":%d}`,
			time.Now().Format(time.RFC3339), r.Method, r.URL.Path, ww.status, time.Since(start).Milliseconds())
	})
}

type wrap struct {
	http.ResponseWriter
	status int
}

func (w *wrap) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

