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
    onResponseError({ response }: { response: { status: number } }) {
      if (response.status === 401 && !import.meta.server) {
        // Token rejected — clear auth state and redirect to login
        store.accessToken = null;
        store.user = null;
        store.tenant = null;
        localStorage.removeItem("daas_token");
        navigateTo("/auth/login");
      }
    },
  } as Record<string, unknown>);
}
