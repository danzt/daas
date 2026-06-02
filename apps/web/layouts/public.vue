<script setup lang="ts">
import { Package, ShoppingCart } from "lucide-vue-next";
import { useCartStore } from "~/stores/cart";

const route = useRoute();
const cartStore = useCartStore();

const tenantSlug = computed(() => (route.params.tenantSlug as string) ?? "");
// In S7, replace with a usePublicTenant() composable that fetches tenant info.
const tenantDisplayName = computed(() => tenantSlug.value.replace(/-/g, " "));
</script>

<template>
  <div class="min-h-screen flex flex-col bg-background">
    <!-- Sticky header -->
    <header
      class="sticky top-0 z-30 border-b bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/80"
    >
      <div
        class="max-w-7xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between gap-4"
      >
        <NuxtLink
          :to="`/t/${tenantSlug}`"
          class="flex items-center gap-3 group"
        >
          <div
            class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors"
          >
            <Package class="w-5 h-5 text-primary" />
          </div>
          <div>
            <h1
              class="text-base font-bold font-heading text-foreground capitalize"
            >
              {{ tenantDisplayName }}
            </h1>
            <p class="text-[11px] text-muted-foreground -mt-0.5">
              Tienda online
            </p>
          </div>
        </NuxtLink>

        <div class="flex items-center gap-3">
          <button
            type="button"
            class="relative w-10 h-10 rounded-lg border bg-card hover:bg-muted transition-colors flex items-center justify-center"
            :aria-label="`Cart (${cartStore.totalItems} items)`"
          >
            <ShoppingCart class="w-4 h-4 text-foreground" />
            <span
              v-if="cartStore.totalItems > 0"
              class="absolute -top-1 -right-1 w-5 h-5 rounded-full bg-primary text-white text-[10px] font-semibold flex items-center justify-center"
            >
              {{ cartStore.totalItems }}
            </span>
          </button>
        </div>
      </div>
    </header>

    <!-- Main content -->
    <main class="flex-1">
      <slot />
    </main>

    <!-- Footer -->
    <footer class="border-t bg-card mt-auto">
      <div
        class="max-w-7xl mx-auto px-4 sm:px-6 py-6 text-center text-sm text-muted-foreground"
      >
        Powered by <span class="font-semibold text-primary">DaaS</span>
      </div>
    </footer>
  </div>
</template>
