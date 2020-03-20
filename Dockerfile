FROM docker/compose
COPY pico /
ENTRYPOINT ["/pico"]
