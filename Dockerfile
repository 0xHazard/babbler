FROM alpine:3.8
COPY babbler /usr/local/bin/babbler
ENTRYPOINT ["/usr/local/bin/babbler"]