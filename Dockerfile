FROM golang:1.18-alpine as builder
RUN apk update && \
    apk add --no-cache git
WORKDIR /project
COPY go.mod go.sum  /project
RUN go mod download -x
COPY . /project
RUN CGO_ENABLED=0 go build -o gsm-pubsub .

FROM alpine:3.15
WORKDIR /app
COPY --from=builder /project/gsm-pubsub /app/gsm-pubsub
CMD ["/app/gsm-pubsub"]
