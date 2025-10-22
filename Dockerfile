FROM alpine:3.20
ARG TARGETPLATFORM

RUN apk add --no-cache ca-certificates

COPY $TARGETPLATFORM/kv /kv
RUN mkdir /work && chmod +x /kv

WORKDIR /work
ENTRYPOINT ["/kv"]
