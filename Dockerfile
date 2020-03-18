FROM golang:alpine as builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN go build -o relaybot .

FROM alpine
RUN apk update && apk add ca-certificates
COPY --from=builder /build/relaybot /app/relaybot
RUN chmod +x /app/relaybot
WORKDIR /app
ENTRYPOINT ["./relaybot"]
