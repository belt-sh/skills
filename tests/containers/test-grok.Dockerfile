FROM ubuntu:24.04

RUN apt-get update && apt-get install -y curl git jq && rm -rf /var/lib/apt/lists/*

# Install belt
RUN curl -fsSL https://cli.inference.sh | sh

# Install Grok CLI
RUN curl -fsSL https://x.ai/cli/install.sh | bash

# Copy plugin repo
COPY . /opt/belt-plugin
COPY tests/containers/entrypoint-grok.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
