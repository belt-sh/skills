---
name: skill
description: "Search, use, install, and publish skills from the belt registry"
allowed-tools: Bash(belt skill *)
---

## Skills

```bash
belt skill list                     # list your skills
belt skill search "query"           # search your skills
belt skill store                    # browse the public skill store
belt skill store search "query"     # search the store
belt skill use <namespace/name>     # fetch and print a skill (on-demand)
belt skill install <namespace/name> # install locally for your agent
belt skill get <namespace/name>     # view skill details
belt skill upload <path>            # publish a skill (dir with SKILL.md or single file)
```

Upload respects visibility — defaults to private. Change visibility on the platform.
Same name = new version. Identical content is skipped (dedup).
