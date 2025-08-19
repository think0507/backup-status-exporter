# ---------- 1) Build stage ----------
FROM golang:1.23-bookworm AS builder
WORKDIR /app

# 모듈 캐시 최적화
COPY go.mod go.sum ./
RUN go mod download

# 소스 복사
COPY . .

# 정적 빌드
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o backup-exporter .

# ---------- 2) Run stage ----------
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /app/backup-exporter /app/backup-exporter
ENV REPO_ROOT=/backup/borgrepo
EXPOSE 9102
ENTRYPOINT ["/app/backup-exporter"]