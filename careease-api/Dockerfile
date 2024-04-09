# Use the official Golang image to create a build artifact.
FROM golang:1.18 as builder

# Copy local code to the container image.
WORKDIR /app
COPY . .

RUN go mod init careease-api
RUN go mod tidy

RUN go build -v -o careease_receiver

# Run the web service on container startup.
CMD ["/app/careease_receiver"]
