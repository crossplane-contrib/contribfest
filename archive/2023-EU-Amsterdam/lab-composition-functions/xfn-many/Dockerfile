FROM golang:1.20-alpine3.17 as builder

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /function .

###
FROM alpine:3.17.3

COPY crossplane.yaml /crossplane.yaml
COPY --from=builder /function /function

ENTRYPOINT ["/function"]