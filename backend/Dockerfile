FROM golang:tip-alpine3.22
WORKDIR /app
COPY . .
RUN go mod tidy
CMD ["go","run","."]