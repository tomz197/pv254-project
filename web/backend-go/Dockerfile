# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 go build -o /out/backend-go ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /srv
ENV PORT=8000
ENV ASSETS_DIR=/srv/assets
COPY --from=build /out/backend-go /usr/local/bin/backend-go
# Expect assets to be mounted or copied to /srv/assets
EXPOSE 8000
USER nonroot:nonroot
CMD ["/usr/local/bin/backend-go"]