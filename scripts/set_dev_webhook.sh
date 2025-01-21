#!/usr/bin/env bash

source ../.env
WEB_HOOK_PATH="/telegram/handler/webhook"
HOST="https://f40c-2a02-8309-b001-4c00-ad69-5c5d-fccd-37fd.ngrok-free.app"
URL="https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook?url=${HOST}${WEB_HOOK_PATH}"
echo "perform request: ${URL}"
curl -X POST "${URL}"
