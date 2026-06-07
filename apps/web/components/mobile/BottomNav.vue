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
  match: string;
}

const tabs: TabItem[] = [
  {
    label: "Inicio",
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
    label: "Stock",
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
    Bottom navigation — native iOS/Android premium style.
    Active state: pill background + bold label + colored icon.
    Inactive: muted icon + muted label.
    safe-area-inset-bottom → clears iOS home indicator.
  -->
  <nav
    class="fixed bottom-0 inset-x-0 z-50 bg-white border-t border-border"
    style="padding-bottom: env(safe-area-inset-bottom)"
  >
    <div class="flex items-end h-[60px]">
      <NuxtLink
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        class="flex flex-1 flex-col items-center justify-center gap-[3px] py-2 relative"
      >
        <!-- Pill highlight behind active icon -->
        <span
          v-if="isTabActive(tab)"
          class="absolute top-2 w-12 h-7 rounded-full bg-primary/10 transition-all duration-200"
          aria-hidden="true"
        />

        <!-- Icon -->
        <component
          :is="tab.icon"
          class="size-[22px] relative z-10 transition-colors duration-150"
          :class="
            isTabActive(tab)
              ? 'text-primary stroke-[2.5]'
              : 'text-muted-foreground/70 stroke-[1.75]'
          "
        />

        <!-- Label -->
        <span
          class="text-[10px] relative z-10 leading-none transition-colors duration-150"
          :class="
            isTabActive(tab)
              ? 'text-primary font-bold'
              : 'text-muted-foreground/60 font-medium'
          "
        >
          {{ tab.label }}
        </span>
      </NuxtLink>
    </div>
  </nav>
</template>
