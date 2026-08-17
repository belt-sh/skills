FROM ubuntu:24.04

RUN apt-get update && apt-get install -y curl git jq npm && rm -rf /var/lib/apt/lists/*

# Install belt
RUN curl -fsSL https://cli.inference.sh | sh

# Install OpenCode
RUN npm install -g opencode-ai || \
    curl -fsSL https://opencode.ai/install | bash || \
    echo "opencode install attempted"

# Copy plugin repo
COPY . /opt/belt-plugin
COPY tests/containers/entrypoint-opencode.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
