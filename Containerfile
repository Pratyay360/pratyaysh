FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /pratyaysh

FROM alpine:3.20

COPY --from=builder /pratyaysh /usr/local/bin/pratyaysh

EXPOSE 2222

ENTRYPOINT ["pratyaysh"]
