FROM golang:alpine

ENV CGO_ENABLED=0

LABEL \
  org.opencontainers.image.title="pzero" \
  org.opencontainers.image.description="pzero framework" \
  org.opencontainers.image.url="https://github.com/polpo-space/pzero" \
  org.opencontainers.image.documentation="https://github.com/polpo-space/pzero#readme" \
  org.opencontainers.image.source="https://github.com/polpo-space/pzero" \
  org.opencontainers.image.licenses="MIT" \
  maintainer="jaronnie <jaron@jaronnie.com>"

WORKDIR /app

COPY dist/pzero_linux_amd64_v1/pzero /dist/pzero_linux_amd64/pzero
COPY dist/pzero_linux_arm64_v8.0/pzero /dist/pzero_linux_arm64/pzero

RUN if [ `go env GOARCH` = "amd64" ]; then \
      cp /dist/pzero_linux_amd64/pzero /usr/local/bin/pzero; \
    elif [ `go env GOARCH` = "arm64" ]; then \
      cp /dist/pzero_linux_arm64/pzero /usr/local/bin/pzero; \
    fi

RUN apk update --no-cache \
  && apk add --no-cache tzdata ca-certificates protoc \
  && pzero check \
  && rm -rf /dist \
  && rm -rf /go/pkg/mod \
  && rm -rf /go/pkg/sumdb

ENTRYPOINT ["pzero"]