// NUXT_CAPACITOR=true → build estático para empaquetar con Capacitor (iOS/Android)
// Sin esa variable → SSR normal para el deploy web
const isCapacitorBuild = process.env.NUXT_CAPACITOR === "true";

export default defineNuxtConfig({
  devtools: { enabled: !isCapacitorBuild },
  ssr: !isCapacitorBuild, // SPA puro para Capacitor, SSR para web
  modules: [
    "@pinia/nuxt",
    "@nuxtjs/tailwindcss",
    "shadcn-nuxt",
    "@nuxt/eslint",
  ],
  shadcn: {
    prefix: "",
    componentDir: "./components/ui",
  },
  css: ["~/assets/css/main.css"],
  typescript: {
    strict: true,
    typeCheck: false,
  },
  eslint: {
    config: {
      standalone: false,
    },
  },
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || "http://localhost:8080",
    },
  },
  // Rutas a pre-renderizar cuando NUXT_CAPACITOR=true
  ...(isCapacitorBuild && {
    nitro: {
      prerender: {
        routes: ["/"],
        crawlLinks: true,
      },
    },
  }),
});
