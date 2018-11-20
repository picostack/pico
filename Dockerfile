# -
# Build workspace
# -
FROM golang AS compile

RUN apt-get update -y && apt-get install --no-install-recommends -y -q build-essential ca-certificates

WORKDIR /wadsworth
ADD . .
RUN make static

# -
# Runtime
# -
FROM scratch

COPY --from=compile /wadsworth/wadsworth /bin/wadsworth
COPY --from=compile /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["wadsworth"]
