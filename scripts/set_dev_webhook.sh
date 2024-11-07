#!/usr/bin/env bash

source ../.env
WEB_HOOK_PATH="/telegram/handler/webhook"
HOST="https://d7ff-2a00-11b7-321e-8800-e9d1-993b-651a-ac05.ngrok-free.app"
URL="https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook?url=${HOST}${WEB_HOOK_PATH}"
echo "perform request: ${URL}"
curl -X POST "${URL}"
