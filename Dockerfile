# FROM golang:alpine as builder

# RUN apk update && apk upgrade && \
#   apk add --no-cache git

# RUN mkdir /app
# WORKDIR /app

# ENV GO111MODULE=on
# ENV GOPROXY=direct
# ENV GOSUMDB=off

# COPY . .

# RUN go mod download
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cleaner-linux-amd64

# Run container
FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN mkdir /app
WORKDIR /app
# COPY --from=builder /app/cleaner-linux-amd64 .
COPY cleaner .

RUN mkdir /data
ENTRYPOINT ["./cleaner"]
#CMD ["--vod_path=/data", "--image_path=/image"]