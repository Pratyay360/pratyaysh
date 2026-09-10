FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /pratyaysh .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates openssh-keygen

COPY --from=builder /pratyaysh /usr/bin/pratyaysh

CMD ["pratyaysh"]
