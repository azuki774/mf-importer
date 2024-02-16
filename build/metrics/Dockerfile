FROM golang:1.20.14 as builder

COPY cmd/ /app/cmd/
COPY internal/ /app/internal/
COPY go.mod /app/go.mod
COPY .git /app/.git
WORKDIR /app
RUN go mod tidy
RUN go build -a -tags "netgo" -installsuffix netgo -ldflags="-s -w -extldflags \"-static\" \
    -X main.version=$(git describe --tag --abbrev=0) \
    -X main.revision=$(git rev-list -1 HEAD) \
    -X main.build=$(git describe --tags)" \
    -o bin/ ./...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/bin/metrics /usr/local/bin/metrics
ENTRYPOINT ["/usr/local/bin/metrics", "start"]

EXPOSE 8800
