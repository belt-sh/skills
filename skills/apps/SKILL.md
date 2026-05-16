---
name: apps
description: "Search and run 250+ AI apps — image generation, video, TTS, web search, LLMs"
allowed-tools: Bash(belt app *)
---

## Apps

```bash
belt app store                      # browse the public app store
belt app store search "query"       # search the store
belt app list                       # list your apps
belt app get <namespace/name>       # view app details and input schema
belt app run <namespace/name> --input '{"prompt":"..."}' --save output.png
```

### Examples

```bash
belt app run inferencesh/flux-1-1-ultra --input '{"prompt":"a cat astronaut"}' --save cat.png
belt app run inferencesh/veo-3 --input '{"prompt":"ocean waves at sunset"}' --save waves.mp4
belt app run inferencesh/tavily-search --input '{"query":"latest AI news"}'
```
