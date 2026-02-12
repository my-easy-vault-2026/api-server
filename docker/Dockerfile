FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY bin/easy_vault_api /app/easy_vault_api
COPY lang /app/lang

EXPOSE 8081 9443

ENTRYPOINT ["/app/easy_vault_api"]
CMD ["serve"]
