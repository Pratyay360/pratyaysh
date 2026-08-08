FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /pratyaysh .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates openssh-keygen && \
    mkdir -p /.ssh

COPY --from=builder /pratyaysh /usr/local/bin/pratyaysh

EXPOSE 2222

ENTRYPOINT ["pratyaysh"]
