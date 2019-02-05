#!/usr/bin/env bash

GOOS=linux go build
docker build -t cmmarslender/plexnanny .
docker push cmmarslender/plexnanny
