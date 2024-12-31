FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY .env .

RUN CGO_ENABLED=0 GOOS=linux go build -o /inventory_app_server

EXPOSE 8080

CMD ["/inventory_app_server"]

# docker build -t inventory-app-server .
# docker run -d --rm --network host inventory-app-server
