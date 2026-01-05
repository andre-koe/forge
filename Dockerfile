FROM golang:1.24-bullseye AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/forge ./cmd/forge

FROM scratch
COPY --from=builder /out/forge /usr/local/bin/forge
COPY LICENSE /usr/share/forge/LICENSE
COPY README.md /usr/share/forge/README.md
ENTRYPOINT ["/usr/local/bin/forge"]
