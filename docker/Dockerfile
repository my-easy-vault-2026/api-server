FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/easy_vault_api main.go
RUN cp -r $(go env GOMODCACHE)/github.com/my-easy-vault-2026/shared-modules@*/lang ./lang

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/easy_vault_api /app/easy_vault_api
COPY --from=builder /app/lang /app/lang
COPY /docker/.env /app/.env

EXPOSE 8081 9443

# CMD ["sleep", "infinity"] // for debugging
ENTRYPOINT ["/app/easy_vault_api"]
CMD ["serve"]
