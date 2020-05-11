FROM docker/compose:1.25.5
COPY pico /
ENTRYPOINT ["/pico"]
