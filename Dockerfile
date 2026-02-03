FROM golang:1.25-alpine AS builder

WORKDIR /app

# 安装 git
RUN apk add --no-cache git ca-certificates

# ---------- 私有仓库配置 ----------
ARG GITHUB_TOKEN

# 告诉 Go 哪些是私有仓库
ENV GOPRIVATE=github.com/trancecho \
    GONOSUMDB=github.com/trancecho

# 使用 token 重写 github 地址
RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# ---------- 依赖 ----------
COPY go.mod go.sum ./
RUN go mod download

# ---------- 源码 ----------
COPY . .

# ---------- 构建 ----------
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o main .

# ---------- 运行时 ----------
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main /app/main