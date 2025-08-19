# ---------- 1) Build stage ----------
FROM golang:1.22-bookworm AS builder
WORKDIR /app

# 모듈 캐시 최적화
COPY go.mod go.sum ./
RUN go mod download

# 소스 복사
COPY . .

# 정적 빌드(작고 빠르게)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o backup-exporter .

# ---------- 2) Run stage (distroless, nonroot) ----------
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

# 바이너리만 복사
COPY --from=builder /app/backup-exporter /app/backup-exporter

# 환경변수 (필요 시 변경)
ENV REPO_ROOT=/backup/borgrepo

EXPOSE 9102

# distroless는 SHELL/패키지 없음(healthcheck 불가 유의)
ENTRYPOINT ["/app/backup-exporter"]