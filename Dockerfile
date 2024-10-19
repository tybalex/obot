# First Stage: Builder
FROM cgr.dev/chainguard/wolfi-base AS builder

# Install build dependencies
RUN apk add --no-cache go build-base npm make

# Set the working directory
WORKDIR /app

# Copy the source code
COPY . .

RUN make ui
# Use build cache for Go modules and build
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go build -o otto8 main.go

# Second Stage: Final
FROM ubuntu:22.04

RUN apt-get update && apt install -y git tini

# Copy the compiled application from the builder stage
COPY --link --from=builder /app/otto8 /bin/

ENV HOME=/data
WORKDIR /data
VOLUME /data
# Command to run the application
CMD ["tini", "--", "otto8", "server"]
