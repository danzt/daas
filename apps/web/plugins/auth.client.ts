import { useAuthStore } from "~/stores/auth";

export default defineNuxtPlugin(() => {
  const store = useAuthStore();
  store.initFromStorage();
});
