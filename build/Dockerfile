FROM alpine:3.21

ARG BINARY_NAME=fileServer-linux-amd64

WORKDIR /app

RUN addgroup -S app && \
    adduser -S app -G app

COPY build/bin/${BINARY_NAME} /app/fileServer
COPY build/config/config.yaml /app/config/config.yaml

RUN chmod +x /app/fileServer && \
    chown -R app:app /app

USER app

EXPOSE 80

ENTRYPOINT ["/app/fileServer", "--config", "/app/config/config.yaml"]