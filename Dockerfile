FROM docker/compose
COPY pico /bin/pico
ENTRYPOINT ["pico"]
