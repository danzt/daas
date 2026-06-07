import type { CapacitorConfig } from "@capacitor/cli";

const config: CapacitorConfig = {
  appId: "com.daas.app",
  appName: "DaaS",
  /**
   * webDir apunta al output de `nuxt generate` (SSG estático).
   * Para development con live reload, descomentar `server.url`.
   */
  webDir: ".output/public",
  server: {
    androidScheme: "https",
    // Live reload durante desarrollo:
    // url: 'http://192.168.x.x:3000',
    // cleartext: true,
  },
  ios: {
    contentInset: "automatic",
  },
  android: {
    allowMixedContent: false,
  },
  plugins: {
    StatusBar: {
      style: "dark",
      backgroundColor: "#FAF5FF", // bg-background del design system
    },
  },
};

export default config;
