# Telegram Bot Server (Go)

This project is a simple Go-based server for a Telegram bot. The bot listens for messages and responds to users based on your defined logic.

You should generate tls flies in path `./tls` (`private.key`, `public.pem`). Also you must configure .env file with similar variables like in `example.env`
The project has two stages: `dev`, `prod`.