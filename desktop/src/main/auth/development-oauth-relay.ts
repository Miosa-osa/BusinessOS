import { createServer, type Server } from "node:http";

export const DEVELOPMENT_OAUTH_RELAY_PORT = 43821;

export interface DevelopmentOAuthRelay {
  url: string;
  close(): Promise<void>;
}

export async function startDevelopmentOAuthRelay(
  onToken: (token: string) => Promise<void>,
  port = DEVELOPMENT_OAUTH_RELAY_PORT,
): Promise<DevelopmentOAuthRelay> {
  const server = createServer(async (request, response) => {
    const requestUrl = new URL(
      request.url || "/",
      `http://127.0.0.1:${port}`,
    );

    if (request.method !== "GET" || requestUrl.pathname !== "/auth/callback") {
      response.writeHead(404, { "Cache-Control": "no-store" });
      response.end("Not found");
      return;
    }

    const token = requestUrl.searchParams.get("token")?.trim();
    if (!token) {
      response.writeHead(400, { "Cache-Control": "no-store" });
      response.end("Missing session token");
      return;
    }

    try {
      await onToken(token);
      response.writeHead(200, {
        "Cache-Control": "no-store",
        "Content-Type": "text/html; charset=utf-8",
      });
      response.end(
        "<!doctype html><meta charset=\"utf-8\"><title>Signed in - BusinessOS</title><p>Signed in. You can close this window.</p>",
      );
    } catch (error) {
      console.error("[auth] development OAuth relay failed", error);
      response.writeHead(500, { "Cache-Control": "no-store" });
      response.end("BusinessOS could not install the session");
    }
  });

  await listen(server, port);
  const address = server.address();
  if (!address || typeof address === "string") {
    await close(server);
    throw new Error("Development OAuth relay did not bind to a TCP port");
  }

  return {
    url: `http://127.0.0.1:${address.port}/auth/callback`,
    close: () => close(server),
  };
}

function listen(server: Server, port: number): Promise<void> {
  return new Promise((resolve, reject) => {
    const onError = (error: Error) => {
      server.off("listening", onListening);
      reject(error);
    };
    const onListening = () => {
      server.off("error", onError);
      resolve();
    };
    server.once("error", onError);
    server.once("listening", onListening);
    server.listen(port, "127.0.0.1");
  });
}

function close(server: Server): Promise<void> {
  if (!server.listening) return Promise.resolve();
  return new Promise((resolve, reject) => {
    server.close((error) => (error ? reject(error) : resolve()));
  });
}
