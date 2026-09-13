# Pinned upstream image index; default build never enables simulation.
FROM golang:1.27.1-bookworm@sha256:648f440f42a0958804efb24df176f806f9d353b41f1c0627f666428e40310f6b AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o /gati ./cmd/gati

FROM scratch
COPY --from=build /gati /gati
COPY config/gati.yaml /config/gati.yaml
COPY data/tirana/roads.geojson /data/tirana/roads.geojson
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/gati"]
CMD ["-listen", "0.0.0.0:8080", "-config", "/config/gati.yaml"]
