ARG DOCKER_ALPINE_VERSION=3.23
FROM golang:1.24-alpine${DOCKER_ALPINE_VERSION} AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /mcpedia

COPY go.mod go.sum ./
COPY vendor vendor
COPY cmd cmd
COPY internal internal

RUN CGO_ENABLED=1 go build -mod=vendor -trimpath -ldflags="-s -w" -o mcpedia ./cmd/mcpedia


ARG DOCKER_ALPINE_VERSION=3.23
FROM alpine:${DOCKER_ALPINE_VERSION}
ARG MCPEDIA_VERSION
LABEL org.opencontainers.image.title="mcpedia"
LABEL org.opencontainers.image.description="MCPedia - MCP knowledge base server"
LABEL org.opencontainers.image.url="https://github.com/pouriya/mcpedia"
LABEL org.opencontainers.image.source="https://github.com/pouriya/mcpedia"
LABEL org.opencontainers.image.version="${MCPEDIA_VERSION}"

WORKDIR /
COPY --from=builder /mcpedia/mcpedia /usr/local/bin/mcpedia

ENTRYPOINT ["/usr/local/bin/mcpedia"]
CMD ["serve"]
