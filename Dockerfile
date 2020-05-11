FROM docker/compose:1.25.1
COPY pico /
ENTRYPOINT ["/pico"]
