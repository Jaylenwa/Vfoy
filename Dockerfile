FROM alpine:latest

WORKDIR /vfoy
COPY vfoy ./vfoy

RUN apk update \
    && apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && chmod +x ./vfoy \
    && mkdir -p /data/aria2 \
    && chmod -R 766 /data/aria2

EXPOSE 5212
VOLUME ["/vfoy/uploads", "/vfoy/avatar", "/data"]

ENTRYPOINT ["./vfoy"]
