export const BeltPlugin = async (_ctx: any) => {
  return {
    "chat.send": async () => {
      const { execSync } = require("child_process");
      try { execSync("belt plugin hook user-prompt-submit", { timeout: 5000 }); } catch {}
    },
    "tool.pre": async () => {
      const { execSync } = require("child_process");
      try { execSync("belt plugin hook pre-tool-use", { timeout: 5000 }); } catch {}
    },
    "tool.post": async () => {
      const { execSync } = require("child_process");
      try { execSync("belt plugin hook post-tool-use", { timeout: 5000 }); } catch {}
    },
    "event": async ({ event }: any) => {
      if (event && event.type === "session.idle") {
        const { execSync } = require("child_process");
        try { execSync("belt plugin hook stop", { timeout: 5000 }); } catch {}
      }
    },
  };
};

export default { id: "belt", server: BeltPlugin };
