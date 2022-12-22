FROM alpine:3.17.0
RUN apk add git>=2.38
COPY bin/devx /usr/bin/devx
ENTRYPOINT ["devx"]
