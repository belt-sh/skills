---
name: suggest
description: "Search skills, knowledge, and apps in one shot — find the best tool for any task"
allowed-tools: Bash(belt suggest *)
---

## Suggest

Search across skills, knowledge, and apps at once. Returns 3 apps, 3 knowledge entries, and 3 skills by default.

```bash
belt suggest "query"
```

### Filter by category

```bash
belt suggest app "query"
belt suggest skill "query"
belt suggest knowledge "query"
```

### Examples

```bash
belt suggest "generate a logo"
belt suggest app "text to speech"
belt suggest skill "web search API"
belt suggest knowledge "deployment"
```
