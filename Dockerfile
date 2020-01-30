FROM golang:1.12-alpine as base

FROM base as builder
# The database backends use cgo.
# Thus, they require a working gcc to compile
RUN apk add --no-cache git gcc libc-dev
# Copy the source files and build the app
COPY . /app/
WORKDIR /app
RUN CGO_ENABLED=1 go build -o PyPiGo .
    

FROM base
RUN adduser -S -D -H -h /app packageserver
USER packageserver
COPY ./templates /app/templates
COPY --from=builder /app/PyPiGo /app/PyPiGo
WORKDIR /app
ENTRYPOINT ["/app/PyPiGo"]
