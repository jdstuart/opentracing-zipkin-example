FROM golang:1.8.1-alpine
EXPOSE 8080
WORKDIR "/go/src/app"
CMD ["go","run","main.go"]