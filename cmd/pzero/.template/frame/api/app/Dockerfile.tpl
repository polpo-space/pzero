FROM --platform=$BUILDPLATFORM ghcr.io/polpo-space/pzero:latest as builder

ARG TARGETARCH
ARG LDFLAGS

ENV GOPROXY https://goproxy.cn,direct

WORKDIR /usr/local/go/src/app

COPY ./ ./

RUN --mount=type=cache,target=/go/pkg CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -ldflags="$LDFLAGS" -o /dist/app main.go \
    && pzero gen swagger \
    && cp -r etc /dist/etc \
    && mkdir -p /dist/desc && cp -r desc/swagger /dist/desc


FROM --platform=$TARGETPLATFORM alpine:latest

WORKDIR /dist

COPY --from=builder /dist .

EXPOSE 8001

CMD ["./app", "server"]