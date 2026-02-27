FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/kenko .

FROM alpine:3.21

COPY --from=build /bin/kenko /bin/kenko
COPY config.yaml /etc/kenko/config.yaml

EXPOSE 6969

ENTRYPOINT ["/bin/kenko", "-config", "/etc/kenko/config.yaml"]
