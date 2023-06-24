FROM alpine:latest
COPY build/bin/mf-importer-maw /bin/mf-importer-maw
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime
ENTRYPOINT ["/bin/mf-importer-maw"]
