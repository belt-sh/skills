---
name: belt
description: "Install and set up belt — the cloud platform CLI for AI agents"
allowed-tools: Bash(belt *), Bash(which belt), Bash(brew install belt-sh/tap/belt), Bash(scoop install belt), Bash(npm install -g @belt-sh/cli)
---

## belt cli

belt is the cloud platform cli for ai agents. single ~4mb binary, no runtime dependencies.

using a purpose-built cli means your agent operates through a constrained, typed interface instead of writing raw curl commands or sdk calls. every operation goes through schema validation — fewer tokens, fewer errors, and no credential leakage.

### install

first check if belt is already installed:

```bash
which belt && belt --version
```

if already installed, skip to authenticate.

**package managers (recommended — verified through each registry's trust chain):**

```bash
brew install belt-sh/tap/belt            # macos / linux (homebrew tap, signed)
scoop bucket add belt https://github.com/belt-sh/scoop-belt && scoop install belt  # windows
npm install -g @belt-sh/cli              # node.js (global install, pinned in package.json)
```

**manual install (full control — download, verify, then run):**

```bash
curl -fsSL https://cli.inference.sh -o /tmp/belt-install.sh
```

the installer is a short, readable shell script. it detects your os and architecture, downloads the matching binary from `dist.inference.sh`, verifies the binary's sha-256 checksum against the published manifest, and places it in your path. no elevated permissions required. the [installer source](https://cli.inference.sh) is publicly readable — review it before running:

```bash
cat /tmp/belt-install.sh   # review the script
sh /tmp/belt-install.sh    # run after review
```

### authenticate

```bash
belt login
belt me
```

### set up claude code integration

```bash
belt init claude-code
```

### quick start

```bash
belt skill list                   # your skills
belt skill store search "web"     # find skills in the store
belt knowledge list               # view your knowledge
belt app store                    # browse AI apps
```

[belt.sh](https://belt.sh) · [docs](https://belt.sh/docs)
