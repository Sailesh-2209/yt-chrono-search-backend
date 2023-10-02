FROM golang
WORKDIR /app
COPY . .
RUN go mod tidy
