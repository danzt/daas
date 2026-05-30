import { describe, it, expect, vi, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useAuthStore } from "../../stores/auth";

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

/**
 * Build a minimal JWT with the given payload (NOT cryptographically valid —
 * only used to test client-side decoding logic).
 */
function buildFakeJwt(payload: object): string {
  const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const body = btoa(JSON.stringify(payload));
  return `${header}.${body}.fakesig`;
}

const fakeToken = buildFakeJwt({
  sub: "user-123",
  email: "owner@acme.com",
  exp: Math.floor(Date.now() / 1000) + 3600,
  app_metadata: { tenant_id: "tenant-abc", role: "owner" },
});

const mockLoginResponse = {
  access_token: fakeToken,
  token_type: "Bearer",
  expires_in: 3600,
  refresh_token: "refresh-xyz",
  user: { id: "user-123", email: "owner@acme.com" },
};

// ──────────────────────────────────────────────────────────────────────────────
// Nuxt/runtime stubs
// ──────────────────────────────────────────────────────────────────────────────

vi.stubGlobal("useRuntimeConfig", () => ({
  public: { apiBase: "http://localhost:8080" },
}));

// Mock $fetch globally
const mockFetch = vi.fn();
vi.stubGlobal("$fetch", mockFetch);

// Stub localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, val: string) => {
      store[key] = val;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();
Object.defineProperty(globalThis, "localStorage", { value: localStorageMock });

// ──────────────────────────────────────────────────────────────────────────────
// Tests
// ──────────────────────────────────────────────────────────────────────────────

describe("useAuthStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    localStorageMock.clear();
    mockFetch.mockReset();
  });

  describe("login()", () => {
    it("sets isAuthenticated to true after successful login", async () => {
      mockFetch.mockResolvedValueOnce(mockLoginResponse);

      const store = useAuthStore();
      expect(store.isAuthenticated).toBe(false);

      await store.login("owner@acme.com", "password123");

      expect(store.isAuthenticated).toBe(true);
    });

    it("stores the access token in localStorage", async () => {
      mockFetch.mockResolvedValueOnce(mockLoginResponse);

      const store = useAuthStore();
      await store.login("owner@acme.com", "password123");

      expect(localStorageMock.getItem("daas_token")).toBe(fakeToken);
    });

    it("decodes JWT and sets user with tenant_id and role", async () => {
      mockFetch.mockResolvedValueOnce(mockLoginResponse);

      const store = useAuthStore();
      await store.login("owner@acme.com", "password123");

      expect(store.user).toMatchObject({
        id: "user-123",
        email: "owner@acme.com",
        tenantId: "tenant-abc",
        role: "owner",
      });
      expect(store.isOwner).toBe(true);
      expect(store.tenantId).toBe("tenant-abc");
    });

    it("throws when $fetch rejects (bad credentials)", async () => {
      mockFetch.mockRejectedValueOnce({
        data: { detail: "Invalid credentials" },
      });

      const store = useAuthStore();
      await expect(
        store.login("bad@email.com", "wrongpass"),
      ).rejects.toBeTruthy();
      expect(store.isAuthenticated).toBe(false);
    });
  });

  describe("logout()", () => {
    it("clears token and user state", async () => {
      // Login first
      mockFetch.mockResolvedValueOnce(mockLoginResponse);
      const store = useAuthStore();
      await store.login("owner@acme.com", "password123");
      expect(store.isAuthenticated).toBe(true);

      // Logout (POST /auth/logout may succeed or fail — store handles it)
      mockFetch.mockResolvedValueOnce({});
      await store.logout();

      expect(store.isAuthenticated).toBe(false);
      expect(store.user).toBeNull();
      expect(store.accessToken).toBeNull();
      expect(localStorageMock.getItem("daas_token")).toBeNull();
    });

    it("clears state even when logout API call fails", async () => {
      mockFetch.mockResolvedValueOnce(mockLoginResponse);
      const store = useAuthStore();
      await store.login("owner@acme.com", "password123");

      // Simulate API error on logout
      mockFetch.mockRejectedValueOnce(new Error("Network error"));
      await store.logout();

      expect(store.isAuthenticated).toBe(false);
      expect(store.user).toBeNull();
    });
  });

  describe("initFromStorage()", () => {
    it("restores session from a valid stored token", () => {
      localStorageMock.setItem("daas_token", fakeToken);

      const store = useAuthStore();
      store.initFromStorage();

      expect(store.isAuthenticated).toBe(true);
      expect(store.user?.email).toBe("owner@acme.com");
    });

    it("clears storage when stored token is expired", () => {
      const expiredToken = buildFakeJwt({
        sub: "user-old",
        email: "old@acme.com",
        exp: Math.floor(Date.now() / 1000) - 3600, // expired 1h ago
        app_metadata: { tenant_id: "tenant-xyz", role: "employee" },
      });
      localStorageMock.setItem("daas_token", expiredToken);

      const store = useAuthStore();
      store.initFromStorage();

      expect(store.isAuthenticated).toBe(false);
      expect(localStorageMock.getItem("daas_token")).toBeNull();
    });
  });
});
