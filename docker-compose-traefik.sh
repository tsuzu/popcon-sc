#! /bin/bash

./ppms/debug/build.sh && ./ppjc/debug/build.sh && ./ppweb/debug/build.sh && docker-compose -f docker-compose-traefik.yml up -d
