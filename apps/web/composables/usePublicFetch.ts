export function usePublicFetch<T>(
  url: string,
  options: Record<string, unknown> = {},
): Promise<T> {
  const config = useRuntimeConfig();
  const route = useRoute();
  const tenantSlug = route.params.tenantSlug as string;
  return $fetch<T>(url, {
    baseURL: `${config.public.apiBase}/t/${tenantSlug}/shop/v1`,
    ...options,
  });
}
