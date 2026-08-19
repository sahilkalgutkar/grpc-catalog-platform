FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY gen ./gen
COPY services/pricing-service ./services/pricing-service
WORKDIR /workspace/services/pricing-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/pricing-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/pricing-service /pricing-service
EXPOSE 9091
ENTRYPOINT ["/pricing-service"]
