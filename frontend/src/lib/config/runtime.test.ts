import { beforeEach, describe, expect, it } from "vitest";
import {
  LOCAL_BACKEND_URL,
  LOCAL_OSA_URL,
  getApiBaseUrl,
  getBackendUrl,
  getLegacyApiBaseUrl,
} from "./runtime";

describe("runtime config", () => {
  beforeEach(() => {
    localStorage.clear();
    Reflect.deleteProperty(window, "electron");
  });

  it("uses relative versioned API URLs on localhost web", () => {
    expect(getApiBaseUrl()).toBe("/api/v1");
  });

  it("uses the local backend as the web backend on localhost", () => {
    expect(getBackendUrl()).toBe(LOCAL_BACKEND_URL);
  });

  it("uses the local OSA API in Electron local mode", () => {
    Object.defineProperty(window, "electron", {
      configurable: true,
      value: {},
    });
    localStorage.setItem("businessos_mode", "local");

    expect(getApiBaseUrl()).toBe(`${LOCAL_OSA_URL}/api/v1`);
    expect(getLegacyApiBaseUrl()).toBe(`${LOCAL_OSA_URL}/api`);
  });

  it("preserves legacy cloud URL storage for unversioned clients", () => {
    Object.defineProperty(window, "electron", {
      configurable: true,
      value: {},
    });
    localStorage.setItem("businessos_mode", "cloud");
    localStorage.setItem("businessos_cloud_url", "https://cloud.example");

    expect(getLegacyApiBaseUrl()).toBe("https://cloud.example/api");
  });
});
