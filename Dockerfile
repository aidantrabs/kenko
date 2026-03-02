FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/kenko ./cmd/kenko

FROM alpine:3.21

RUN addgroup -S kenko && adduser -S kenko -G kenko

COPY --from=build --chown=kenko:kenko /bin/kenko /bin/kenko
COPY --chown=kenko:kenko configs/config.yaml /etc/kenko/config.yaml

USER kenko

EXPOSE 6969

ENTRYPOINT ["/bin/kenko", "-config", "/etc/kenko/config.yaml"]
