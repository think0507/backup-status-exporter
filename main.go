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
	// 이전 스캔 결과 싹 비우기(사라진 서버 라벨 잔상 제거)
	lastTS.Reset()

	entries, err := os.ReadDir(root)
	if err != nil {
		log.Println("readdir error:", err)
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
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

	// /metrics 요청이 들어올 때마다 최신 스캔 → 메트릭 제공
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		scanRepo(repoRoot)
		// 실제 메트릭 응답
		promhttp.Handler().ServeHTTP(w, r)

		// 오래 걸리면 로그로만 참고(선택)
		if d := time.Since(start); d > 2*time.Second {
			log.Printf("scanRepo took %s\n", d)
		}
	})

	log.Println("listen :9102")
	log.Fatal(http.ListenAndServe(":9102", nil))
}
