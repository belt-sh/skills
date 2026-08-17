FROM node:22-slim

RUN apt-get update && apt-get install -y curl git jq && rm -rf /var/lib/apt/lists/*

# Install belt
RUN curl -fsSL https://cli.inference.sh | sh

# Install Kilo Code CLI
RUN npm install -g @kilocode/cli || echo "kilo install attempted"

# Copy plugin repo
COPY . /opt/belt-plugin
COPY tests/containers/entrypoint-kilo.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
