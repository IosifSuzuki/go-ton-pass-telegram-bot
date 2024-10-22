FROM golang:1.21-alpine as BuildStage

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
COPY --from=BuildStage app/jsons /jsons

CMD ["./main"]