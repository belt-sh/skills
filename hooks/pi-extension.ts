export default function (pi: any) {
  pi.on("before_agent_start", async (event: any) => {
    const { execSync } = require("child_process");
    try {
      const output = execSync("belt plugin hook user-prompt-submit", {
        input: JSON.stringify(event),
        timeout: 5000,
        encoding: "utf-8",
      });
      if (output.trim()) {
        const result = JSON.parse(output);
        if (result.systemPrompt) {
          return { systemPrompt: (event.systemPrompt || "") + "\n" + result.systemPrompt };
        }
      }
    } catch {}
    return {};
  });

  pi.on("agent_end", async () => {
    const { execSync } = require("child_process");
    try {
      execSync("belt plugin hook stop", { timeout: 5000 });
    } catch {}
  });
}
