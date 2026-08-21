const { execSync } = require("child_process");

function runHook(event: string, input?: any): any {
  try {
    const output = execSync(`belt plugin hook ${event}`, {
      input: input ? JSON.stringify(input) : undefined,
      timeout: 5000,
      encoding: "utf-8",
    });
    if (output.trim()) return JSON.parse(output);
  } catch {}
  return {};
}

const BeltPlugin = async (_ctx: any) => {
  return {
    "experimental.chat.system.transform": async (_input: any, output: any) => {
      const result = runHook("user-prompt-submit");
      if (result.systemPrompt) {
        output.system.push(result.systemPrompt);
      }
    },
    "tool.execute.before": async () => {
      runHook("pre-tool-use");
    },
    "tool.execute.after": async () => {
      runHook("post-tool-use");
    },
    "event": async ({ event }: any) => {
      if (event.type === "session.idle") {
        runHook("stop");
      }
    },
  };
};

export default { id: "belt", server: BeltPlugin };
