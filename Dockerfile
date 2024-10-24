# First Stage: Builder
FROM cgr.dev/chainguard/wolfi-base AS builder

# Install build dependencies
RUN apk add --no-cache go npm make git pnpm

# Set the working directory
WORKDIR /app

# Copy the source code
COPY . .

RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    make in-docker-build

# Second Stage: Final
FROM ubuntu:22.04

RUN apt-get update && apt install -y git tini openssh-server

RUN sed -E 's/^#(PermitRootLogin)no/\1yes/' /etc/ssh/sshd_config -i
RUN ssh-keygen -A
RUN mkdir /run/sshd && /usr/sbin/sshd


# Copy the compiled application from the builder stage
COPY --from=builder /app/bin/otto8 /bin/
COPY --link --from=builder /app/otto8-tools /otto8-tools
COPY --chmod=0755 <<EOF /bin/run.sh
#!/bin/bash
mkdir -p /run/sshd
/usr/sbin/sshd -D &
exec tini -- otto8 server
EOF

EXPOSE 22
ENV HOME=/data
ENV OTTO_SERVER_TOOL_REGISTRY=/otto8-tools
ENV OTTO_SERVER_WORKSPACE_TOOL=/otto8-tools/workspace-provider
WORKDIR /data
VOLUME /data
CMD ["run.sh"]
