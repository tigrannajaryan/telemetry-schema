FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY custcol /
EXPOSE 55680 55679
ENTRYPOINT ["/custcol"]
CMD ["--config", "/etc/otel/config.yaml"]
