#!/usr/bin/env bash

source ../.env
WEB_HOOK_PATH="/telegram/handler/webhook"
HOST="https://ab2b-2a00-11b7-321e-8800-d940-9ad1-b5dc-226f.ngrok-free.app"
URL="https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook?url=${HOST}${WEB_HOOK_PATH}"
echo "perform request: ${URL}"
curl -X GET "${URL}"
