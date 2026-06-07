<script setup lang="ts">
import {
  LayoutDashboard,
  Package,
  ShoppingCart,
  Warehouse,
  MoreHorizontal,
} from "lucide-vue-next";

const route = useRoute();

interface TabItem {
  label: string;
  icon: Component;
  to: string;
  /** Match prefix — active when route starts with this */
  match: string;
}

const tabs: TabItem[] = [
  {
    label: "Dashboard",
    icon: LayoutDashboard,
    to: "/dashboard",
    match: "/dashboard",
  },
  {
    label: "Productos",
    icon: Package,
    to: "/products",
    match: "/products",
  },
  {
    label: "Ventas",
    icon: ShoppingCart,
    to: "/sales-orders",
    match: "/sales-orders",
  },
  {
    label: "Inventario",
    icon: Warehouse,
    to: "/inventory",
    match: "/inventory",
  },
  {
    label: "Más",
    icon: MoreHorizontal,
    to: "/settings",
    match: "/settings",
  },
];

function isTabActive(tab: TabItem) {
  return route.path === tab.match || route.path.startsWith(tab.match + "/");
}
</script>

<template>
  <!--
    Bottom navigation for native iOS/Android.
    Uses env(safe-area-inset-bottom) to push above the iOS home indicator.
  -->
  <nav
    class="fixed bottom-0 inset-x-0 z-50 flex border-t border-border bg-sidebar/95 backdrop-blur-sm"
    style="padding-bottom: env(safe-area-inset-bottom)"
  >
    <NuxtLink
      v-for="tab in tabs"
      :key="tab.to"
      :to="tab.to"
      class="flex flex-1 flex-col items-center justify-center gap-0.5 py-2 text-[10px] font-medium transition-colors"
      :class="
        isTabActive(tab)
          ? 'text-primary'
          : 'text-muted-foreground hover:text-foreground'
      "
    >
      <component
        :is="tab.icon"
        class="size-5 shrink-0"
        :class="isTabActive(tab) ? 'stroke-[2.5]' : 'stroke-[1.75]'"
      />
      <span>{{ tab.label }}</span>
    </NuxtLink>
  </nav>
</template>
