---
name: knowledge
description: "Save and search your knowledge base — discoveries, insights, references, preferences"
allowed-tools: Bash(belt knowledge *)
---

## Knowledge

```bash
belt knowledge search "query"       # search your knowledge base
belt knowledge list --limit 10      # list recent entries
belt knowledge get <id>             # view an entry
belt knowledge upload <file.md>     # save from file (supports frontmatter)
belt knowledge delete <id>          # remove an entry
```

### Saving from a file with frontmatter

```markdown
---
title: Redis connection pooling gotcha
type: observation
---
When using redis-go with connection pooling, the default MaxIdleConns is 0...
```

### Saving inline

```bash
echo "discovered X" | belt knowledge upload - --name "x-thing" --type observation
```

**Types:** observation, concept, preference, reference
