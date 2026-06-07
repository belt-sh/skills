---
name: remember
description: "Save and recall knowledge — discoveries, insights, references, preferences"
allowed-tools: Bash(belt know *), Bash(belt knowledge *)
---

## Remember

Save knowledge that someone with zero context can understand and act on 6 months from now. Don't save generic rules — save the specific incident, decision, or reference with enough detail to be useful.

### Quality bar

Every entry MUST include: what specifically happened or was decided, where (repo/file/system/project), and why it matters (what breaks if ignored). If the content lacks a concrete noun (file, project, endpoint, command), add detail before saving.

### Structure by type

- **observation** — what happened + where + why it matters + when discovered
- **concept** — the mental model or architectural decision + what it implies for future work
- **preference** — the rule + the reason behind it (past incident or strong opinion)
- **reference** — what the resource is + where it lives + when to consult it

### Commands

```bash
belt know upload <file.md>           # save from file (supports frontmatter)
belt know search "query"             # search your knowledge base
belt know list --limit 10            # list recent entries
belt know get <id>                   # view an entry
belt know delete <id>                # remove an entry
```

```bash
echo "content" | belt know upload - --name "descriptive-name" --type observation --tags "area,topic"
```

**Types:** observation, concept, preference, reference
