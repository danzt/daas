import { Capacitor } from "@capacitor/core";
import { StatusBar, Style } from "@capacitor/status-bar";

/**
 * Plugin client-only — inicializa Capacitor plugins en el arranque nativo.
 * En web no hace nada (los plugins no están disponibles).
 */
export default defineNuxtPlugin(() => {
  if (!Capacitor.isNativePlatform()) return;

  // Status bar: fondo claro, íconos oscuros (match del design system)
  StatusBar.setStyle({ style: Style.Light }).catch(() => {
    // Silencioso: en algunos simuladores/dispositivos puede no estar disponible
  });

  StatusBar.setBackgroundColor({ color: "#FAF5FF" }).catch(() => {});
});
