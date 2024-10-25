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
FROM cgr.dev/chainguard/wolfi-base AS final

# Install build dependencies
RUN apk add --no-cache git py3.12-pip openssh-server nodejs bash tini
RUN ln -s /usr/bin/python3.12 /usr/bin/python3

RUN sed -E 's/^#(PermitRootLogin)no/\1yes/' /etc/ssh/sshd_config -i
RUN ssh-keygen -A
RUN mkdir /run/sshd && /usr/sbin/sshd


# Copy the compiled application from the builder stage
COPY --from=builder /app/bin/otto8 /bin/
COPY --link --from=builder /app/otto8-tools /otto8-tools
RUN <<EOF
for i in $(find /otto8-tools/[^k]* -name requirements.txt -exec cat {} \; -exec echo \; | sort -u); do
    pip install "$i"
done
EOF
COPY --chmod=0755 <<EOF /bin/run.sh
#!/bin/bash
set -e
mkdir -p /run/sshd
/usr/sbin/sshd -D &
mkdir -p /data/cache
ln -sf /otto8-tools /data/cache/system
exec tini -- otto8 server
EOF

EXPOSE 22
ENV HOME=/data
ENV XDG_CACHE_HOME=/data/cache
ENV OTTO_SERVER_TOOL_REGISTRY=/data/cache/system
ENV OTTO_SERVER_WORKSPACE_TOOL=/data/cache/system/workspace-provider
WORKDIR /data
VOLUME /data
CMD ["run.sh"]
