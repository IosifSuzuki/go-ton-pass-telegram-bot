FROM golang:1.20.4-alpine3.18 as BuildStage

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod tidy

COPY . .
RUN go build -o main cmd/main.go

FROM alpine:latest

WORKDIR /

COPY --from=BuildStage app/main /main
COPY --from=BuildStage app/locales /locales

CMD ["./main"]