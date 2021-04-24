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
ENTRYPOINT [ "/app/domainchecker" ]