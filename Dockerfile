FROM golang:alpine as builder
RUN echo "Building domain-checcker"
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
WORKDIR /build/cmd/domain-checker
RUN go build -o /build/domainchecker .
FROM alpine
COPY --from=builder /build/domainchecker /app/
COPY cmd/domain-checker/docker-config.json /app/config.json
WORKDIR /app
RUN wget https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh && \
    apk update && \
    apk add bash && \
    chmod +x wait-for-it.sh
CMD [ "/app/wait-for-it.sh", "rabbitmq:5672", "--timeout=60", "--", "/app/wait-for-it.sh", "mongodb:27017", "--timeout=60","--", "/app/domainchecker" ]