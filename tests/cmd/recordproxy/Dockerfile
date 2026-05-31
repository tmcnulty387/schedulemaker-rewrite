# Build context is the repo root (see schedulemaker/docker-compose.yml).
FROM golang:1.26 AS build
WORKDIR /src
COPY tests/ ./tests/
WORKDIR /src/tests
RUN CGO_ENABLED=0 go build -o /out/recordproxy ./cmd/recordproxy

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/recordproxy /recordproxy
EXPOSE 9000
ENTRYPOINT ["/recordproxy"]
