---
name: remember
description: "Save and recall knowledge — discoveries, insights, references, preferences"
allowed-tools: Bash(belt know *), Bash(belt knowledge *)
---

## Remember

```bash
belt know upload <file.md>           # save from file (supports frontmatter)
belt know search "query"             # search your knowledge base
belt know list --limit 10            # list recent entries
belt know get <id>                   # view an entry
belt know delete <id>                # remove an entry
```

### Inline

```bash
echo "discovered X" | belt know upload - --name "x-thing" --type observation
```

### With tags

```bash
echo "insight" | belt know upload - --name "name" --type observation --tags "redis,perf"
```

**Types:** observation, concept, preference, reference
