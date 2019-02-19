# build stage
FROM golang:1.11 AS build-env
# RUN apk update && apk upgrade && apk add --no-cache git
ADD . /src
RUN cd /src && CGO_ENABLED=0 go build

# final stage
FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build-env /src/goprojectdemo /app/
CMD ./goprojectdemo
