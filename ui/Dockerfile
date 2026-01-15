FROM golang:1.26rc2-alpine AS builder

ARG VERSION=undefined

WORKDIR /app
COPY . .

# Build statically and set the value of meta.Version to the arg VERSION
RUN go build -ldflags "-s -w -X 'github.com/dkrizic/feature/ui/meta.Version=${VERSION}'" -o feature-ui main.go

# Run all tests
RUN go test ./...

FROM scratch
COPY --from=builder /app/feature-ui /app/feature-ui

ENV LOG_FORMAT=json
ENV LOG_LEVEL=info
ENV PORT=8000

EXPOSE 8000
ENTRYPOINT ["/app/feature-ui"]
CMD ["service"]
