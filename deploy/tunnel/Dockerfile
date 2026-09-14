# Source version and content checksum are also recorded in source.json.
FROM golang:1.27.1-bookworm@sha256:648f440f42a0958804efb24df176f806f9d353b41f1c0627f666428e40310f6b AS build
WORKDIR /bootstrap
RUN go mod download -json github.com/cloudflare/cloudflared@v0.0.0-20260910205223-f11dea9cb707 > /source.json
# Check the content hash independently of the proxy's current response.
RUN grep -F '"Sum": "h1:K6+hRRs5RDxsUKtA2mq1sAm4GjKgA34ZjAxIhATGt+c="' /source.json
WORKDIR /src
RUN cp -a /go/pkg/mod/github.com/cloudflare/cloudflared@v0.0.0-20260910205223-f11dea9cb707/. /src/ && chmod -R u+w /src
COPY deploy/tunnel/go.mod deploy/tunnel/go.sum ./
RUN go mod download all && go mod verify
COPY scripts/runtimeprep/main.go scripts/runtimeprep/licenses.go /tools/runtimeprep/
RUN go run /tools/runtimeprep/*.go /src
RUN CGO_ENABLED=0 go build -mod=readonly -trimpath -buildvcs=false -ldflags="-X main.Version=2026.9.1-gati.1 -X main.BuildTime=reproducible -X github.com/cloudflare/cloudflared/metrics.Runtime=virtual" -o /cloudflared ./cmd/cloudflared
RUN go run /tools/runtimeprep/*.go licenses /licenses

FROM scratch
COPY --from=build /cloudflared /usr/local/bin/cloudflared
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /licenses /licenses
COPY deploy/tunnel/LICENSE /licenses/cloudflared-LICENSE
COPY LICENSE /licenses/GATI-LICENSE
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/cloudflared"]
CMD ["version"]
