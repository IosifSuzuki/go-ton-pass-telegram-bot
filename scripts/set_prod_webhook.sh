#!/usr/bin/env bash

source ../.env
WEB_HOOK_PATH="/telegram/handler/webhook"
PORT="88"
HOST="https://ec2-35-156-253-213.eu-central-1.compute.amazonaws.com"
URL="https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook?url=${HOST}:${PORT}${WEB_HOOK_PATH}"
echo "perform request: ${URL}"
curl -v -F certificate=@./../tls/certificate.crt "${URL}"
