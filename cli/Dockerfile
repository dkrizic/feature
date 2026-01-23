FROM golang:1.25.6-alpine3.23 AS builder

ARG VERSION=undefined

WORKDIR /app
COPY . .

# Build statically and set the value of meta.Version to the arg VERSION
RUN go build -ldflags "-s -w -X 'github.com/dkrizic/feature/cli/meta.Version=${VERSION}'" -o feature-cli main.go

# Run all tests
RUN go test ./...

FROM alpine:3.23.2
COPY --from=builder /app/feature-cli /app/feature-cli
ENV PATH="/app:${PATH}"
RUN ln -s /app/feature-cli /app/cli
RUN ln -s /app/feature-cli /app/feature
ENTRYPOINT ["sh", "-c", "trap : TERM INT; sleep infinity & wait"]
