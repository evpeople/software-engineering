FROM golang:alpine AS builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
# 之所以先做上面的内容，是为了充分利用Docker的缓存。
# ADD 会解压，会从URL拷贝文件，COPY可以多阶段构建的时候传递文件
COPY . .
RUN go build -o main main.go

FROM alpine

ENV TZ Asia/Shanghai

WORKDIR /build

ENV GIN_MODE=release
COPY --from=builder /build/main /build/main

CMD ["./main"]