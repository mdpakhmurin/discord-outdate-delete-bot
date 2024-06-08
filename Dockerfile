# Use the official Golang image as a base
FROM golang:1.22.4-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the source into the working directory
COPY /app .

# Download all dependencies
RUN go mod download

# Add sqlite3 utility
RUN apk add --no-cache sqlite

# Build the application
RUN go build -o main .

# Run the application
CMD ["./main"]
