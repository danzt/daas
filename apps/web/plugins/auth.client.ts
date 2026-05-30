import { useAuthStore } from "~/stores/auth";

export default defineNuxtPlugin(async () => {
  const store = useAuthStore();
  store.initFromStorage();
  if (store.isAuthenticated) {
    await store.fetchTenant();
  }
});
