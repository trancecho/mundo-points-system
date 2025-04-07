FROM golang:1.24.1-alpine AS builder

LABEL authors="mundo"

# 设置 Go 环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct \
    GOPROXYCONNECTTIMEOUT=300s

# 设置工作目录
WORKDIR /mundo/mundo-points-system

# 将 go.mod 和 go.sum 复制到工作目录
COPY go.mod go.sum ./

# 安装 Git
RUN apk update && apk add --no-cache git

# 设置 GitHub 访问令牌用于认证
ARG GITHUB_TOKEN
RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# 下载 Go 项目的依赖
RUN go mod download && go mod verify

# 将源代码复制到工作目录
COPY . .

# 编译 Go 项目
RUN go build -ldflags="-s -w" -o main .

# 使用更小的 Alpine 镜像作为运行时镜像
FROM alpine:latest

# 安装ca-certificates以支持HTTPS
RUN apk --no-cache add ca-certificates

# 创建工作目录
WORKDIR /app

# 创建配置文件目录
RUN mkdir -p /app/config

# 从构建阶段复制编译好的二进制文件到运行时镜像
COPY --from=builder /mundo/mundo-points-system/main /app/main

# 复制配置文件（确保存在配置文件）
COPY --from=builder /mundo/mundo-points-system/config/* /app/config/

# 暴露 gRPC 服务端口（根据实际端口调整）
EXPOSE 12345

# 设置工作目录为应用目录
WORKDIR /app

# 运行 gRPC 服务
CMD ["./main"]