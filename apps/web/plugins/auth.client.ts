import { useAuthStore } from "~/stores/auth";

export default defineNuxtPlugin(() => {
  const store = useAuthStore();
  store.initFromStorage();
  if (store.isAuthenticated) {
    // Fire-and-forget: don't block app initialization.
    // fetchTenant handles 401 internally (clears stale tokens).
    store.fetchTenant();
  }
});
