FROM iron/base

ADD cleaner-linux-amd64 /

ENTRYPOINT ["./cleaner-linux-amd64", "/data"]