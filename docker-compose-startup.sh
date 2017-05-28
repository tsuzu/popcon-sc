#! /bin/bash

./ppms/debug/build.sh && ./ppjc/debug/build.sh && ./ppms/debug/build.sh && docker-compose -f docker-compose-debug.yml up
