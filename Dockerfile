FROM golang:1.13 as base

#------------------
# Builder
#------------------
FROM base as builder
# Explicitly enable G11MODULES- This should not be necessary with go 1.14
# go-sqlite3 requires cgo. Explicitly enable it
ENV CGO_ENABLED=1 \
    G11MODULE="on"
COPY . /app/
WORKDIR /app
RUN make build EXECUTBALE=/app/GoatCheese

#------------------
# Production Image
#------------------
FROM base
RUN adduser -S -D -H -h /app packageserver
USER packageserver
COPY ./templates /app/templates
COPY --from=builder /app/GoatCheese /app/GoatCheese
WORKDIR /app
ENTRYPOINT ["/app/GoatCheese"]
