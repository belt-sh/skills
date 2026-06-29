---
name: belt
description: "Install and set up belt — the cloud platform CLI for AI agents"
allowed-tools: Bash(belt *), Bash(which belt), Bash(curl -fsSL https://cli.inference.sh -o *), Bash(sh /tmp/belt-install.sh)
---

## Belt CLI

Belt is the cloud platform for AI agents. Run 250+ AI apps, manage knowledge, search and publish skills, connect MCP servers.

### Install

First check if belt is already installed:

```bash
which belt && belt --version
```

If belt is already installed, skip to Authenticate.

If not installed, download the installer, then run it:

```bash
curl -fsSL https://cli.inference.sh -o /tmp/belt-install.sh
```

Review what the script does, then install:

```bash
sh /tmp/belt-install.sh
```

### Authenticate

```bash
belt login
belt me
```

### Set up Claude Code integration

```bash
belt init claude-code
```

### Quick start

```bash
belt skill list                   # your skills
belt skill store search "web"     # find skills in the store
belt knowledge list               # view your knowledge
belt app store                    # browse AI apps
```

[belt.sh](https://belt.sh) · [docs](https://belt.sh/docs)
