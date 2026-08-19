FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY gen ./gen
COPY services/catalog-service ./services/catalog-service
WORKDIR /workspace/services/catalog-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/catalog-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/catalog-service /catalog-service
EXPOSE 9090 8080
ENTRYPOINT ["/catalog-service"]
