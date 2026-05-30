import { useAuthStore } from "~/stores/auth";

export function useAuth() {
  const store = useAuthStore();

  return {
    user: store.user,
    tenant: store.tenant,
    isAuthenticated: store.isAuthenticated,
    isOwner: store.isOwner,
    tenantId: store.tenantId,
    login: store.login,
    logout: store.logout,
    register: store.register,
  };
}

export function useApiFetch<T>(
  url: string,
  options: Record<string, unknown> = {},
): Promise<T> {
  const store = useAuthStore();
  const config = useRuntimeConfig();
  const headers: Record<string, string> = store.accessToken
    ? { Authorization: `Bearer ${store.accessToken}` }
    : {};

  return $fetch<T>(url, {
    baseURL: config.public.apiBase,
    headers,
    ...options,
  });
}
