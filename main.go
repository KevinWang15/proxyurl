package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Config represents the application configuration sourced from config.json.
type Config struct {
	ProxyURL string `json:"proxy_url"`
}

func loadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	if cfg.ProxyURL == "" {
		return Config{}, errors.New("proxy_url is required in config.json")
	}
	return cfg, nil
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load %s: %v", configPath, err)
	}

	proxyEndpoint, err := url.Parse(cfg.ProxyURL)
	if err != nil {
		log.Fatalf("invalid proxy url: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyEndpoint),
		},
		Timeout: 45 * time.Second,
	}

	http.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		routeParam := r.URL.Query().Get("route")
		if routeParam == "" {
			http.Error(w, "missing route query parameter", http.StatusBadRequest)
			return
		}

		targetRaw, err := url.QueryUnescape(routeParam)
		if err != nil {
			http.Error(w, "invalid url encoding for route", http.StatusBadRequest)
			return
		}

		targetURL, err := url.Parse(targetRaw)
		if err != nil || targetURL.Scheme == "" || targetURL.Host == "" {
			http.Error(w, "route must be a valid absolute URL", http.StatusBadRequest)
			return
		}
		if targetURL.Scheme != "http" && targetURL.Scheme != "https" {
			http.Error(w, "unsupported scheme (only http/https allowed)", http.StatusBadRequest)
			return
		}

		username := targetURL.User.Username()
		password, _ := targetURL.User.Password()

		reqURL := *targetURL
		reqURL.User = nil

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, reqURL.String(), nil)
		if err != nil {
			http.Error(w, "failed to build request", http.StatusInternalServerError)
			return
		}

		if username != "" {
			req.SetBasicAuth(username, password)
		}

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "upstream request failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for key, vals := range resp.Header {
			for _, v := range vals {
				w.Header().Add(key, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		if _, err := io.Copy(w, io.LimitReader(resp.Body, 10<<20)); err != nil {
			log.Printf("failed to copy response body: %v", err)
		}
	})

	addr := ":8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		addr = ":" + fromEnv
	}

	log.Printf("proxy fetcher listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
