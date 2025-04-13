FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/app

RUN CGO_ENABLED=0 GOOS=linux go build -o /pvz

EXPOSE 8080
EXPOSE 3000
EXPOSE 9000

CMD ["/pvz"]