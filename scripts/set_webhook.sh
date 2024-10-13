#!/usr/bin/env bash

source ../.env
WEB_HOOK_PATH="/telegram/handler/webhook"
HOST="https://26fb-2a00-11b7-321e-8800-780f-6f3-7753-53b8.ngrok-free.app"
URL="https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook?url=${HOST}${WEB_HOOK_PATH}"
echo "perform request: ${URL}"
curl -X GET "${URL}"