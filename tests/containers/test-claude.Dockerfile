FROM node:22-slim

RUN apt-get update && apt-get install -y curl git jq && rm -rf /var/lib/apt/lists/*

# Install belt
RUN curl -fsSL https://cli.inference.sh | sh

# Install Claude Code
RUN npm install -g @anthropic-ai/claude-code

# Create non-root user (claude refuses --dangerously-skip-permissions as root)
RUN useradd -m -s /bin/bash testuser
RUN cp /usr/local/bin/belt /usr/local/bin/belt 2>/dev/null || true

# Copy plugin repo
COPY . /opt/belt-plugin
COPY tests/containers/entrypoint-claude.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

USER testuser
WORKDIR /home/testuser

ENTRYPOINT ["/entrypoint.sh"]
