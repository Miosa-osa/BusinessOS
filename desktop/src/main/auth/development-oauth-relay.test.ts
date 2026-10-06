import assert from "node:assert/strict";
import test from "node:test";
import { startDevelopmentOAuthRelay } from "./development-oauth-relay";

test("development OAuth relay forwards a session token", async () => {
  let received = "";
  const relay = await startDevelopmentOAuthRelay(
    async (token) => {
      received = token;
    },
    0,
  );

  try {
    const response = await fetch(`${relay.url}?token=session-token`);

    assert.equal(response.status, 200);
    assert.equal(received, "session-token");
    assert.match(await response.text(), /signed in/i);
  } finally {
    await relay.close();
  }
});

test("development OAuth relay rejects missing tokens", async () => {
  const relay = await startDevelopmentOAuthRelay(async () => {}, 0);

  try {
    const response = await fetch(relay.url);

    assert.equal(response.status, 400);
  } finally {
    await relay.close();
  }
});
