FROM cgr.dev/chainguard/wolfi-base AS base

RUN apk add --no-cache go make git npm pnpm

FROM base AS bin
WORKDIR /app
COPY . .
RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/uv \
    --mount=type=cache,target=/root/go/pkg/mod \
    make all

FROM base AS tools
RUN apk add --no-cache curl python-3.13 py3.13-pip
WORKDIR /app
COPY . .
RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/uv \
    --mount=type=cache,target=/root/go/pkg/mod \
    UV_LINK_MODE=copy BIN_DIR=/bin make package-tools

FROM cgr.dev/chainguard/wolfi-base AS final
RUN apk add --no-cache git python-3.13 py3.13-pip openssh-server npm bash tini procps libreoffice
COPY --chmod=0755 /tools/package-chrome.sh /
RUN /package-chrome.sh && rm /package-chrome.sh
RUN sed -E 's/^#(PermitRootLogin)no/\1yes/' /etc/ssh/sshd_config -i
RUN ssh-keygen -A
RUN mkdir /run/sshd && /usr/sbin/sshd
COPY encryption.yaml /
COPY --chmod=0755 <<EOF /bin/run.sh
#!/bin/bash
set -e
if [ "\$OPENAI_API_KEY" = "" ]; then
    echo OPENAI_API_KEY env is required to be set
    exit 1
fi
mkdir -p /run/sshd
/usr/sbin/sshd -D &
mkdir -p /data/cache
# This is YAML
export OTTO8_SERVER_VERSIONS="$(cat <<VERSIONS
"github.com/otto8-ai/tools": "$(cd /otto8-tools && git rev-parse HEAD)"
"github.com/gptscript-ai/workspace-provider": "$(cd /otto8-tools/workspace-provider && git rev-parse HEAD)"
"github.com/gptscript-ai/datasets": "$(cd /otto8-tools/datasets && git rev-parse HEAD)"
"github.com/kubernetes-sigs/aws-encryption-provider": "$(cd /otto8-tools/aws-encryption-provider && git rev-parse HEAD)"
# double echo to remove trailing whitespace
"chrome": "$(echo $(/opt/google/chrome/chrome --version))"
VERSIONS
)"
exec tini -- otto8 server
EOF

COPY --link --from=tools /app/otto8-tools /otto8-tools
COPY --from=bin /app/bin/otto8 /bin/

EXPOSE 22
# libreoffice executables
ENV PATH=/otto8-tools/venv/bin:$PATH:/usr/lib/libreoffice/program
ENV HOME=/data
ENV XDG_CACHE_HOME=/data/cache
ENV GPTSCRIPT_SYSTEM_TOOLS_DIR=/otto8-tools/
ENV OTTO8_SERVER_WORKSPACE_TOOL=/otto8-tools/workspace-provider
ENV OTTO8_SERVER_DATASETS_TOOL=/otto8-tools/datasets
ENV OTTO8_SERVER_TOOL_REGISTRY=/otto8-tools
ENV OTTO8_SERVER_ENCRYPTION_CONFIG_FILE=/encryption.yaml
ENV GOMEMLIMIT=1GiB
ENV TERM=vt100
WORKDIR /data
VOLUME /data
CMD ["run.sh"]
