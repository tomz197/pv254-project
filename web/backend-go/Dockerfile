# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
	go build -o backend-go ./...

FROM alpine:3.20
WORKDIR /srv
ENV PORT=8000
COPY --from=build /app/backend-go /usr/local/bin/backend-go
EXPOSE 8000
CMD ["/usr/local/bin/backend-go"]