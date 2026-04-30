import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { McpError, ErrorCode } from "@modelcontextprotocol/sdk/types.js";
import { z } from "zod";
import { engine, EngineError } from "../client.js";

export function registerWorkspaces(server: McpServer): void {
  server.tool(
    "workspaces",
    "List available knowledge workspaces in the engine. Each workspace is its own isolated brain with its own wiki and signals.",
    {
      tenant: z
        .string()
        .optional()
        .describe(
          "Filter workspaces by tenant identifier. Omit to list all accessible workspaces.",
        ),
    },
    async ({ tenant }) => {
      try {
        const result = await engine.workspaces({ tenant });
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
        };
      } catch (err) {
        if (err instanceof EngineError) {
          throw new McpError(
            ErrorCode.InternalError,
            `Engine error ${err.status}: ${err.message}`,
          );
        }
        throw err;
      }
    },
  );
}
