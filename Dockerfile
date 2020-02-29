FROM docker/compose
COPY picobot /bin/picobot
ENTRYPOINT ["picobot"]
