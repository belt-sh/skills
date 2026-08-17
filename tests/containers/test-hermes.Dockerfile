FROM python:3.12-slim

RUN apt-get update && apt-get install -y curl git jq && rm -rf /var/lib/apt/lists/*

# Install belt
RUN curl -fsSL https://cli.inference.sh | sh

# Install Hermes Agent
RUN pip install hermes-agent || echo "hermes install attempted"

# Copy plugin repo
COPY . /opt/belt-plugin
COPY tests/containers/entrypoint-hermes.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
