FROM golang:1.23-bookworm as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./mealie-ical-proxy .

FROM scratch

WORKDIR /app
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/mealie-ical-proxy .

EXPOSE 3333

CMD ["./mealie-ical-proxy"]
