# First stage: build the Go application.
FROM golang:1.21 as builder

# Set the working directory inside the container.
WORKDIR /app

# Copy the go.mod and go.sum files first to leverage Docker cache.
COPY go.mod go.sum ./

# Download the dependencies.
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Build the Go app. Adjust the CGO_ENABLED and GOOS as needed.
RUN CGO_ENABLED=0 go build -o app .

# Second stage: create the runtime image.
FROM alpine:latest  

# Set the working directory.
WORKDIR /root/

# Copy the binary from the builder stage.
COPY --from=builder /app/app .

COPY public public
COPY views views

COPY docs ./docs

# Command to run the executable.
CMD ["./app"]
