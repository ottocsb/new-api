# syntax=docker/dockerfile:1.7
# Dockerfile（使用预构建前端）
# 推荐使用根目录 Makefile：make build 会先打包前端再构建镜像
# 单独执行 docker build 时，需要确保 web/dist 已存在
# 需要启用 BuildKit（现代 docker 默认开启），cache mounts 会让 go build 走增量编译

FROM golang:1.26.1-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS builder
ENV GO111MODULE=on CGO_ENABLED=0

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ENV GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64}
ENV GOEXPERIMENT=greenteagc

WORKDIR /build

ADD go.mod go.sum ./
# relaykit 是独立 go module（go.mod 里用 replace 指向 ./relaykit），
# go mod download 解析 replace 时需要它的 go.mod/go.sum 先就位。
ADD relaykit/go.mod relaykit/go.sum ./relaykit/
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY main.go ./
COPY common ./common
COPY constant ./constant
COPY controller ./controller
COPY dto ./dto
COPY i18n ./i18n
COPY logger ./logger
COPY middleware ./middleware
COPY model ./model
COPY oauth ./oauth
COPY pkg ./pkg
COPY relay ./relay
COPY relaykit ./relaykit
COPY router ./router
COPY service ./service
COPY setting ./setting
COPY types ./types
COPY web/dist ./web/dist
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags "-s -w -X 'newapi/common.Version=${VERSION}'" -o new-api

FROM debian:bookworm-slim@sha256:f06537653ac770703bc45b4b113475bd402f451e85223f0f2837acbf89ab020a

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata libasan8 wget \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

COPY --from=builder /build/new-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/new-api"]
