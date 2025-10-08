# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /app

# go.mod first for better layer caching
COPY go.mod ./
ENV GOTOOLCHAIN=auto
RUN go mod download

# copy sources and build
COPY . .
ENV CGO_ENABLED=0
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o server

# tiny runtime image
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
ENV PY_SERVICE_URL=http://service-python:8001
ENV NODE_SERVICE_URL=http://service-node:8002
COPY --from=build /app/server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]

