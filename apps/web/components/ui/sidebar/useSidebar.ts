import { inject, provide, ref, computed, onMounted, onUnmounted } from "vue";
import type { InjectionKey, Ref, ComputedRef } from "vue";

export interface SidebarContext {
  open: Ref<boolean>;
  openMobile: Ref<boolean>;
  isMobile: Ref<boolean>;
  state: ComputedRef<"expanded" | "collapsed">;
  setOpen: (value: boolean) => void;
  setOpenMobile: (value: boolean) => void;
  toggleSidebar: () => void;
}

export const SidebarKey: InjectionKey<SidebarContext> = Symbol("sidebar");

const COOKIE = "sidebar_state";
const ONE_WEEK = 60 * 60 * 24 * 7;

export function provideSidebar(defaultOpen = true): SidebarContext {
  const open = ref(defaultOpen);
  const openMobile = ref(false);
  const isMobile = ref(false);

  const state = computed<"expanded" | "collapsed">(() =>
    open.value ? "expanded" : "collapsed",
  );

  function setOpen(value: boolean) {
    open.value = value;
    if (typeof document !== "undefined") {
      document.cookie = `${COOKIE}=${value}; path=/; max-age=${ONE_WEEK}`;
    }
  }

  function setOpenMobile(value: boolean) {
    openMobile.value = value;
  }

  function toggleSidebar() {
    if (isMobile.value) {
      setOpenMobile(!openMobile.value);
    } else {
      setOpen(!open.value);
    }
  }

  function checkMobile() {
    if (typeof window === "undefined") return;
    isMobile.value = window.innerWidth < 768;
  }

  function onKeyDown(event: KeyboardEvent) {
    if (event.key === "b" && (event.metaKey || event.ctrlKey)) {
      event.preventDefault();
      toggleSidebar();
    }
  }

  onMounted(() => {
    checkMobile();
    window.addEventListener("resize", checkMobile);
    window.addEventListener("keydown", onKeyDown);
    // Restore from cookie
    const match = document.cookie.match(new RegExp(`${COOKIE}=([^;]+)`));
    if (match) open.value = match[1] === "true";
  });

  onUnmounted(() => {
    if (typeof window === "undefined") return;
    window.removeEventListener("resize", checkMobile);
    window.removeEventListener("keydown", onKeyDown);
  });

  const context: SidebarContext = {
    open,
    openMobile,
    isMobile,
    state,
    setOpen,
    setOpenMobile,
    toggleSidebar,
  };

  provide(SidebarKey, context);
  return context;
}

export function useSidebar(): SidebarContext {
  const ctx = inject(SidebarKey);
  if (!ctx) throw new Error("useSidebar must be used within a SidebarProvider");
  return ctx;
}
