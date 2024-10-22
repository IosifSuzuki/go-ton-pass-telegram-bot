#!/usr/bin/env bash

CMD=$1

function main() {
  if [ "$CMD" == "dev" ]; then
    docker-compose -f ./docker-compose-dev.yml build --no-cache
    docker-compose -f ./docker-compose-dev.yml up -d
  elif [ "$CMD" == "prod" ]; then
    docker-compose -f ./docker-compose-prod.yml build --no-cache
    docker-compose -f ./docker-compose-prod.yml up -d
  fi
}

#start endpoint
main