FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM debian:bookworm-slim
RUN useradd --system --uid 10001 volunteer
WORKDIR /app
COPY --from=build /out/server /app/server
RUN mkdir -p /app/var/uploads && chown -R volunteer:volunteer /app
USER volunteer
EXPOSE 8080
ENTRYPOINT ["/app/server"]
