############################
# STEP 0 get dependencies
############################
FROM golang:1.16.3 AS dependencies
ENV GOPRIVATE=*.countmax.ru
WORKDIR /go/src/commonapi/
COPY go.mod .
COPY go.sum .
ADD ./ssl/git.countmax.ru.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates
RUN go mod download
############################
# STEP 1 build executable binary
############################
FROM dependencies AS builder
LABEL maintainer="it@watcom.ru" version="0.0.4"
ARG BUILD_NUMBER
ARG GIT_HASH
ENV BUILD_NUMBER ${BUILD_NUMBER}
ENV GIT_HASH ${GIT_HASH}
ENV TZ=Europe/Moscow
ENV GOPRIVATE=*.countmax.ru
ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /go/src/commonapi/
COPY . .
RUN go get -u github.com/swaggo/swag/cmd/swag && swag init
RUN go build -ldflags="-X 'main.build=$BUILD_NUMBER' -X 'main.githash=$GIT_HASH'"
############################
# STEP 2 build a small image
############################
FROM alpine
RUN apk add --no-cache tzdata wget
ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
COPY --from=builder /go/src/commonapi/commonapi /go/bin/commonapi/commonapi
COPY --from=builder /go/src/commonapi/config.yaml /go/bin/commonapi/config.yaml
WORKDIR /go/bin/commonapi/
EXPOSE 8000
CMD ["./commonapi"]