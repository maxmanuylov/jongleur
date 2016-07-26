FROM alpine
RUN apk add --no-cache ca-certificates && update-ca-certificates
COPY jongleur /opt/bin/jongleur
ENTRYPOINT ["/opt/bin/jongleur"]
