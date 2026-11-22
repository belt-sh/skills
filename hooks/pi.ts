export default function (pi: any) {
  pi.on("before_agent_start", async (event: any) => {
    const beltInput = JSON.stringify({ prompt: (event && event.prompt) || "" });
    const { execSync } = require("child_process");
    try {
      const out = execSync("belt plugin hook user-prompt-submit", { input: beltInput, timeout: 10000, stdio: ["pipe", "pipe", "ignore"] }).toString().trim();
      if (out) return { systemPrompt: (event.systemPrompt || "") + "\n" + out };
    } catch {}
  });
  pi.on("agent_end", async () => {
    const { execSync } = require("child_process");
    try { execSync("belt plugin hook stop", { timeout: 60000, stdio: ["ignore", "pipe", "ignore"] }); } catch {}
  });
}
