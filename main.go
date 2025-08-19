package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ns = "backup"

	lastTS = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "last_success_timestamp_seconds",
			Help:      "Last successful backup unix timestamp",
		},
		[]string{"server"},
	)
)

func scanRepo(root string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		log.Println("readdir error:", err)
		return
	}
	for _, e := range entries {
		if !e.IsDir() { continue }
		server := e.Name()
		info, err := os.Stat(filepath.Join(root, server))
		if err != nil {
			log.Println("stat error:", err)
			continue
		}
		lastTS.WithLabelValues(server).Set(float64(info.ModTime().Unix()))
	}
}

func main() {
	prometheus.MustRegister(lastTS)

	repoRoot := os.Getenv("REPO_ROOT")
	if repoRoot == "" {
		repoRoot = "/backup/borgrepo"
	}

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			BeforeServe: func(w http.ResponseWriter, r *http.Request) {
				scanRepo(repoRoot) // 스크랩 직전에 매번 최신 스캔
			},
		},
	))

	log.Println("listen :9102")
	log.Fatal(http.ListenAndServe(":9102", nil))
}
