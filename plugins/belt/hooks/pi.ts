export default function (pi: any) {
  pi.on("before_agent_start", async () => {
    const { execSync } = require("child_process");
    try { execSync("belt plugin hook user-prompt-submit", { timeout: 5000 }); } catch {}
  });
  pi.on("agent_end", async () => {
    const { execSync } = require("child_process");
    try { execSync("belt plugin hook stop", { timeout: 5000 }); } catch {}
  });
}
