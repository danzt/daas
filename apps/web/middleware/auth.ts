import { useAuthStore } from "~/stores/auth";

export default defineNuxtRouteMiddleware((to) => {
  const store = useAuthStore();
  // Hydrate from localStorage on first navigation (client-side)
  store.initFromStorage();

  const publicPaths = ["/auth/login", "/auth/register", "/t/"];
  const isPublic = publicPaths.some((p) => to.path.startsWith(p));

  if (!store.isAuthenticated && !isPublic) {
    return navigateTo("/auth/login");
  }

  if (store.isAuthenticated && to.path.startsWith("/auth")) {
    return navigateTo("/dashboard");
  }
});
