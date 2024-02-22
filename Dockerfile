FROM golang:1.18.2-alpine3.16 as base
RUN apk update
WORKDIR /src/moviedbcsv
ADD . .
RUN go mod download
RUN go build -o moviedbcsv ./cmd/main.go

FROM alpine:3.16 as binary
WORKDIR /src/app
COPY --from=base /src/moviedbcsv/moviedbcsv .
EXPOSE 3000