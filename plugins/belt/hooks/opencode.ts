export const BeltPlugin = async (_ctx: any) => {
  let beltPrompt = "";
  let beltSession = "";
  let beltPending = false;
  return {
    "chat.message": async (input: any, output: any) => {
      beltSession = (input && input.sessionID) || "";
      beltPrompt = ((output && output.parts) || [])
        .filter((p: any) => p && p.type === "text")
        .map((p: any) => p.text)
        .join("\n");
      beltPending = true;
    },
    "experimental.chat.system.transform": async (_input: any, output: any) => {
      if (!beltPending) return;
      beltPending = false;
      const beltInput = JSON.stringify({ prompt: beltPrompt, session_id: beltSession });
      const { execSync } = require("child_process");
      try {
        const out = execSync("belt plugin hook user-prompt-submit", { input: beltInput, timeout: 10000, stdio: ["pipe", "pipe", "ignore"], env: { ...process.env, AI_AGENT: "opencode" } }).toString().trim();
        if (out) output.system.push(out);
      } catch {}
    },
    "tool.execute.after": async () => {
      const { execSync } = require("child_process");
      try { execSync("belt plugin hook post-tool-use", { timeout: 10000, stdio: ["ignore", "pipe", "ignore"], env: { ...process.env, AI_AGENT: "opencode" } }); } catch {}
    },
    "event": async ({ event }: any) => {
      if (event && event.type === "session.idle") {
        const { execSync } = require("child_process");
        try { execSync("belt plugin hook stop", { timeout: 60000, stdio: ["ignore", "pipe", "ignore"], env: { ...process.env, AI_AGENT: "opencode" } }); } catch {}
      }
    },
  };
};
