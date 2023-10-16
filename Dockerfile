# Atlanta's Weather Service Dockerfile

# Start from the latest golang base image
FROM --platform=linux/amd64 docker.io/golang:latest as builder

# Add Maintainer Info
LABEL maintainer="Carlos Alvarado carlos-alvarado@outlook.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

COPY models ./models
COPY rabbitTypes.go ./rabbitTypes.go
COPY main.go ./main.go


# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download


# Build the Go app
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o weather-service .

RUN chmod +x /app/weather-service

FROM --platform=linux/amd64 docker.io/alpine:latest

RUN mkdir /app

COPY --from=builder /app/weather-service /app

CMD [ "/app/weather-service"]