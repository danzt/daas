import { Capacitor } from "@capacitor/core";

/**
 * useMobileMode — single source of truth for "should we render the
 * EXCLUSIVE mobile design?".
 *
 * Production trigger: the app is running inside Capacitor (iOS/Android).
 * Preview trigger (DEV ONLY): `?m=1` in the URL, persisted to localStorage,
 *   so the mobile design can be inspected in a desktop browser at phone
 *   width. Disable with `?m=0`.
 *
 * The preview override is guarded by `import.meta.dev`, so it is compiled
 * out of production builds entirely. A stray `?m=1` (or a stale localStorage
 * flag) can NEVER flip the real web app into mobile mode for users — in
 * production, isMobile === isNative.
 *
 * SSR-safe: always returns false on the server; the client re-evaluates on
 * mount.
 */

const PREVIEW_KEY = "daas:mobile-preview";

// Module-level singleton so every caller shares the same reactive flag.
const previewOverride = ref(false);
let initialized = false;

function initPreview() {
  // Development-only affordance. In production builds this whole block is
  // tree-shaken away (import.meta.dev === false), so previewOverride stays
  // false and the web app is never affected by the ?m flag.
  if (!import.meta.dev) return;
  if (initialized || import.meta.server) return;
  initialized = true;

  try {
    const url = new URL(window.location.href);
    const q = url.searchParams.get("m");
    if (q === "1") localStorage.setItem(PREVIEW_KEY, "1");
    else if (q === "0") localStorage.removeItem(PREVIEW_KEY);

    previewOverride.value = localStorage.getItem(PREVIEW_KEY) === "1";
  } catch {
    previewOverride.value = false;
  }
}

export const useMobileMode = () => {
  if (import.meta.client) initPreview();

  const isNative = computed(() =>
    import.meta.server ? false : Capacitor.isNativePlatform(),
  );

  /** True when we should render the exclusive mobile experience. */
  const isMobile = computed(() => isNative.value || previewOverride.value);

  return { isMobile, isNative };
};
