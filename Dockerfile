FROM golang:1.15 as builder

# Copy local code to the container image.
WORKDIR /app
COPY . .
t p
# Build the command inside the container.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o brillwtf

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine
RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/brillwtf /brillwtf

# Run the web service on container startup.
CMD ["/brillwtf"]
