FROM golang:1.23.5-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

# Копирование зависимостей в образ
COPY ["go.mod", "go.sum", "./"]

RUN go mod download

COPY ./ ./

RUN go build -o ./bin/wallet ./cmd/main.go

FROM alpine 

COPY --from=builder /usr/local/src/bin/wallet /
COPY /config.env /config.env
COPY /666.txt /666.txt

CMD ["/wallet"]