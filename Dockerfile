FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY pii ./pii
COPY cmd ./cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/pii-guard ./cmd/pii-guard

FROM alpine:3.20
COPY --from=build /out/pii-guard /usr/local/bin/pii-guard
ENTRYPOINT ["/usr/local/bin/pii-guard"]
CMD ["--help"]
