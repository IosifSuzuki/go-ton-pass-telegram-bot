#!/usr/bin/env bash

source ../.env
WEB_HOOK_PATH="/telegram/handler/webhook"
HOST="https://1c4b-3-70-223-176.ngrok-free.app"
URL="https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook?url=${HOST}${WEB_HOOK_PATH}"
echo "perform request: ${URL}"
curl -X GET "${URL}"
