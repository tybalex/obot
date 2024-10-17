# First Stage: Builder
FROM cgr.dev/chainguard/wolfi-base AS builder

# Install build dependencies
RUN apk add --no-cache go build-base

# Set the working directory
WORKDIR /app

# Copy the source code
COPY . .

# Use build cache for Go modules and build
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o otto8 main.go

# Second Stage: Final
FROM cgr.dev/chainguard/wolfi-base

# Set the working directory
WORKDIR /app

# Copy the compiled application from the builder stage
COPY --from=builder /app/otto8 .

VOLUME /data

ENV HOME=/data

WORKDIR /data
# Command to run the application
CMD ["../otto8", "server"]
