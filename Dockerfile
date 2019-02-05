FROM alpine:latest

COPY plexnanny /usr/bin/plexnanny

CMD "plexnanny"
