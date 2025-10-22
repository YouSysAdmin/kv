FROM alpine:3.20
RUN apk add --no-cache ca-certificates

COPY kv /kv
RUN mkdir /work && chmod +x /kv

WORKDIR /work
ENTRYPOINT ["/kv"]
