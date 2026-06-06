# ── 阶段一：编译 Go 二进制 ──
FROM golang:1.26-alpine AS builder

WORKDIR /build

# 使用国内 Go 代理
ENV GOPROXY=https://goproxy.cn,direct

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o interview-server ./cmd/main.go

# ── 阶段二：最小运行镜像 ──
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/interview-server .

EXPOSE 8080

CMD ["./interview-server"]
