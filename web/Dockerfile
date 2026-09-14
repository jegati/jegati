FROM node:24.21.0-bookworm-slim@sha256:2fe369e969550cde8e867afc3fe370b260140cab4a23d467074295b42163d553 AS build
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --ignore-scripts --no-audit --no-fund
COPY web/ ./
RUN npm run build

FROM golang:1.27.1-bookworm@sha256:648f440f42a0958804efb24df176f806f9d353b41f1c0627f666428e40310f6b AS proxy
WORKDIR /src
COPY deploy/caddy/go.mod deploy/caddy/go.sum ./
RUN go mod download all && go mod verify
COPY deploy/caddy/ ./
COPY scripts/runtimeprep/main.go scripts/runtimeprep/licenses.go /tools/runtimeprep/
RUN CGO_ENABLED=0 go build -mod=readonly -trimpath -buildvcs=false -o /caddy .
RUN go run /tools/runtimeprep/*.go licenses /licenses

FROM scratch
COPY --from=proxy /caddy /usr/bin/caddy
COPY --from=proxy /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=proxy /licenses /licenses
COPY LICENSE /licenses/GATI-LICENSE
COPY --from=build /src/web/dist /srv
COPY deploy/Caddyfile /etc/caddy/Caddyfile
ENV XDG_CONFIG_HOME=/tmp/config XDG_DATA_HOME=/tmp/data
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/usr/bin/caddy"]
CMD ["run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
