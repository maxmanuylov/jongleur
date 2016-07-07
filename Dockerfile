FROM alpine
COPY jongleur /opt/bin/jongleur
ENTRYPOINT ["/opt/bin/jongleur"]
