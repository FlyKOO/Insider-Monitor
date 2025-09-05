# 构建阶段
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

# 复制 go mod 和 go sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -o /insider-monitor ./cmd/monitor

# 最终阶段
FROM alpine:latest

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /insider-monitor .
# 复制示例配置
COPY config.example.json .

# 创建卷以持久化数据
VOLUME ["/app/data"]

# 运行二进制文件
ENTRYPOINT ["./insider-monitor"]
