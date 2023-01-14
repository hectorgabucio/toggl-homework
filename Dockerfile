

## Build
FROM golang:1.19 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /homework cmd/main.go

FROM debian:latest

WORKDIR /

COPY --from=build /homework /homework

#COPY --from=build /app/infrastructure/sql/migrations/* /infrastructure/sql/migrations/

ARG DATABASE_URL
ARG PORT

ENTRYPOINT ["/homework"]