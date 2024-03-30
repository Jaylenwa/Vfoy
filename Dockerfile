# 第一阶段：编译 Go 应用程序
FROM golang:1.18 AS builder

WORKDIR /app

COPY . .

# 设置 GOPROXY 环境变量
ENV GOPROXY=https://goproxy.cn,direct

RUN CGO_ENABLED=0 GOOS=linux go build -a -o vfoy .

# 第二阶段：构建最终镜像
FROM alpine:latest

WORKDIR /vfoy

COPY --from=builder /app /vfoy

RUN apk update \
    && apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && chmod +x ./vfoy \
    && mkdir -p /data/aria2 \
    && chmod -R 766 /data/aria2

EXPOSE 1214

VOLUME ["/vfoy/uploads", "/vfoy/avatar", "/data"]

ENTRYPOINT ["./vfoy"]
