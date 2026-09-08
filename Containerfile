FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /pratyaysh .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates openssh-keygen && \
    mkdir -p /.ssh

# Copy cloudflared from the official Cloudflare image
COPY --from=cloudflare/cloudflared:latest /usr/local/bin/cloudflared /usr/local/bin/cloudflared

COPY --from=builder /pratyaysh /usr/local/bin/pratyaysh

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 2222

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
