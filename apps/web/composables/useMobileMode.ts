import { Capacitor } from "@capacitor/core";

/**
 * useMobileMode — single source of truth for "should we render the
 * EXCLUSIVE mobile design?".
 *
 * Production trigger: the app is running inside Capacitor (iOS/Android).
 * Preview trigger (dev only): `?m=1` in the URL, persisted to localStorage,
 *   so the mobile design can be inspected in a desktop browser at phone
 *   width without compiling to a device. Disable with `?m=0`.
 *
 * SSR-safe: always returns false on the server (no window/Capacitor),
 * the client re-evaluates on mount. Web at any width is NEVER affected
 * unless the preview flag is explicitly set — so the web app is safe by
 * construction.
 */

const PREVIEW_KEY = "daas:mobile-preview";

// Module-level singleton so every caller shares the same reactive flag.
const previewOverride = ref(false);
let initialized = false;

function initPreview() {
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
