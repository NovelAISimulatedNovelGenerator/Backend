# NovelAI 应用服务 Dockerfile
# 使用多阶段构建减小最终镜像大小

# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装编译依赖
RUN apk add --no-cache git gcc musl-dev

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制所有源代码
COPY . .

# 编译应用
RUN go build -o novelai-server .

# 第二阶段：运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制编译好的应用
COPY --from=builder /app/novelai-server .

# 创建配置目录
RUN mkdir -p /app/configs

# 暴露API服务端口（根据Hertz服务配置调整）
EXPOSE 8888

# 定义启动命令
ENTRYPOINT ["/app/novelai-server"]
