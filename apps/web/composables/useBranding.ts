/**
 * useBranding — fetches and caches storefront branding for a given tenant slug.
 *
 * Returns reactive refs for all branding fields. Falls back to DaaS defaults
 * when fields are empty so callers never need null-checks on display values.
 */

export interface BrandingData {
  store_name: string;
  tagline: string;
  logo_url: string;
  banner_url: string;
  primary_color: string;
}

const DEFAULT_PRIMARY_COLOR = "#7C3AED";

export function useBranding(tenantSlug: Ref<string> | string) {
  const config = useRuntimeConfig();
  const slug = isRef(tenantSlug) ? tenantSlug : ref(tenantSlug);

  const branding = ref<BrandingData>({
    store_name: "",
    tagline: "",
    logo_url: "",
    banner_url: "",
    primary_color: "",
  });
  const brandingLoaded = ref(false);

  // Resolved values with graceful fallbacks.
  const storeName = computed(
    () => branding.value.store_name || slug.value.replace(/-/g, " "),
  );
  const tagline = computed(() => branding.value.tagline || "");
  const logoURL = computed(() => branding.value.logo_url || "");
  const bannerURL = computed(() => branding.value.banner_url || "");
  const primaryColor = computed(
    () => branding.value.primary_color || DEFAULT_PRIMARY_COLOR,
  );

  async function fetchBranding() {
    if (!slug.value) return;
    try {
      const data = await $fetch<BrandingData>(
        `/t/${slug.value}/shop/v1/branding`,
        { baseURL: config.public.apiBase },
      );
      branding.value = data;
    } catch {
      // non-fatal — defaults are used
    } finally {
      brandingLoaded.value = true;
    }
  }

  return {
    branding,
    brandingLoaded,
    storeName,
    tagline,
    logoURL,
    bannerURL,
    primaryColor,
    fetchBranding,
  };
}
