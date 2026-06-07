import { Capacitor } from "@capacitor/core";

/**
 * Detecta la plataforma en tiempo de ejecución.
 * Safe para SSR: siempre retorna 'web' en el servidor.
 */
export const usePlatform = () => {
  const isNative = computed(() =>
    import.meta.server ? false : Capacitor.isNativePlatform(),
  );

  const platform = computed(() =>
    import.meta.server ? "web" : Capacitor.getPlatform(),
  );

  const isIos = computed(() => platform.value === "ios");
  const isAndroid = computed(() => platform.value === "android");
  const isWeb = computed(() => platform.value === "web");

  return { isNative, platform, isIos, isAndroid, isWeb };
};
