FROM golang:1.25.5-alpine AS builder

ARG VERSION=undefined

WORKDIR /app
COPY . .

# Build statically and set the value of meta.Version to the arg VERSION
RUN go build -ldflags "-s -w -X 'meta.Version=${VERSION}'" -o feature-cli main.go

# Run all tests
RUN go test ./...

FROM alpine:2.11.7
COPY --from=builder /app/feature-cli /app/feature-cli
EXPOSE 8000
ENTRYPOINT ["/bin/pause"]
