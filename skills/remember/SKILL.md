---
name: remember
description: "Save and recall knowledge — discoveries, insights, references, preferences"
allowed-tools: Bash(belt knowledge *)
---

## Remember

```bash
belt knowledge upload <file.md>     # save from file (supports frontmatter)
belt knowledge search "query"       # search your knowledge base
belt knowledge list --limit 10      # list recent entries
belt knowledge get <id>             # view an entry
belt knowledge delete <id>          # remove an entry
```

### Inline

```bash
echo "discovered X" | belt knowledge upload - --name "x-thing" --type observation
```

**Types:** observation, concept, preference, reference
