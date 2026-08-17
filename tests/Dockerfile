FROM node:22-slim AS base

RUN apt-get update && apt-get install -y \
    curl git jq unzip python3 python3-pip python3-venv \
    && rm -rf /var/lib/apt/lists/*

# Install Go (for building harness-test)
RUN curl -fsSL https://go.dev/dl/go1.24.5.linux-$(dpkg --print-architecture).tar.gz | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:/root/go/bin:$PATH"

# Install Bun (needed for OpenCode plugins)
RUN curl -fsSL https://bun.sh/install | bash
ENV PATH="/root/.bun/bin:$PATH"

# Install belt
RUN curl -fsSL https://cli.inference.sh | sh

# Install all harness CLIs (each on its own layer for caching)
RUN npm install -g @anthropic-ai/claude-code
RUN npm install -g @openai/codex
RUN npm install -g @github/copilot
RUN npm install -g opencode-ai
RUN npm install -g --ignore-scripts @earendil-works/pi-coding-agent
RUN npm install -g mastracode || echo "mastracode install failed"
RUN curl -fsSL https://x.ai/cli/install.sh | bash
RUN pip install --break-system-packages hermes-agent

# Create non-root user (Claude Code refuses --dangerously-skip-permissions as root)
RUN useradd -m -s /bin/bash testuser

# Copy repo and build harness-test
COPY . /opt/belt-plugin
WORKDIR /opt/belt-plugin/tools/harness-test
RUN go build -o /usr/local/bin/harness-test .

WORKDIR /root
ENTRYPOINT ["harness-test"]
