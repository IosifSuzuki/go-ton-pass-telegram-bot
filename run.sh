#!/usr/bin/env bash

function main() {
  docker-compose -f ./docker-compose.yml build --no-cache
  docker-compose -f ./docker-compose.yml up -d
}

main