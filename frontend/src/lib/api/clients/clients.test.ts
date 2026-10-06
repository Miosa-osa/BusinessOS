import { describe, expect, it } from "vitest";
import { normalizeClientListItem } from "./clients";

describe("normalizeClientListItem", () => {
  it("defaults missing aggregate values to zero", () => {
    const client = normalizeClientListItem({ id: "client-1", name: "Atlas Roofing" });

    expect(client.contacts_count).toBe(0);
    expect(client.interactions_count).toBe(0);
    expect(client.deals_count).toBe(0);
    expect(client.active_deals_value).toBe(0);
  });

  it("normalizes numeric aggregate strings from the API", () => {
    const client = normalizeClientListItem({
      id: "client-1",
      contacts_count: "2",
      interactions_count: 3,
      deals_count: "1",
      active_deals_value: "28000.00",
    });

    expect(client.contacts_count).toBe(2);
    expect(client.interactions_count).toBe(3);
    expect(client.deals_count).toBe(1);
    expect(client.active_deals_value).toBe(28000);
  });
});
