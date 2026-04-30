import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { McpError, ErrorCode } from "@modelcontextprotocol/sdk/types.js";
import { z } from "zod";
import { engine, EngineError } from "../client.js";
import { config } from "../config.js";

export function registerAddMemory(server: McpServer): void {
  server.tool(
    "add_memory",
    "Add a fact, decision, or observation as a first-class memory. Required: content. Optional: is_static (true for permanent facts), audience, citation_uri (URI of source), source_chunk_id. Every memory is integrity-gated — must tie back to a source. Use when the user mentions information that should persist beyond this conversation.",
    {
      content: z
        .string()
        .min(1)
        .describe(
          "The memory content to store. Be precise and self-contained.",
        ),
      workspace: z
        .string()
        .optional()
        .describe(
          `Workspace to store into. Defaults to "${config.defaultWorkspace}".`,
        ),
      is_static: z
        .boolean()
        .optional()
        .describe(
          "Mark true for permanent ground-truth facts (e.g. company founding date, core product definition).",
        ),
      audience: z
        .string()
        .optional()
        .describe(
          "Audience lens this memory is relevant to (engineering, sales, legal, exec, all).",
        ),
      citation_uri: z
        .string()
        .optional()
        .describe(
          "URI of the source document or conversation this memory originates from.",
        ),
      source_chunk_id: z
        .string()
        .optional()
        .describe(
          "ID of the specific source chunk this memory is derived from.",
        ),
    },
    async ({
      content,
      workspace,
      is_static,
      audience,
      citation_uri,
      source_chunk_id,
    }) => {
      try {
        const result = await engine.addMemory({
          content,
          workspace: workspace ?? config.defaultWorkspace,
          is_static,
          audience,
          citation_uri,
          source_chunk_id,
        });
        return {
          content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
        };
      } catch (err) {
        if (err instanceof EngineError) {
          if (err.status === 422) {
            throw new McpError(
              ErrorCode.InvalidParams,
              `Memory integrity check failed: ${err.message}`,
            );
          }
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
