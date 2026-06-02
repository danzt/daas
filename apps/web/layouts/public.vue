<script setup lang="ts">
import { Package, ShoppingCart, Check } from "lucide-vue-next";
import { useCartStore } from "~/stores/cart";

const route = useRoute();
const cartStore = useCartStore();

const tenantSlug = computed(() => (route.params.tenantSlug as string) ?? "");
// In S7+, replace with a usePublicTenant() composable that fetches tenant info.
const tenantDisplayName = computed(() => tenantSlug.value.replace(/-/g, " "));

// Hydrate the cart from localStorage on the client. The Pinia store can't
// read localStorage during SSR, so we trigger the restore here once the
// layout mounts.
onMounted(() => {
  cartStore.restore();
});

// Bounce animation on the cart icon whenever an item lands.
const bouncing = ref(false);
const showToast = ref(false);
const lastItemName = ref("");

watch(
  () => cartStore.lastAddedAt,
  (now, prev) => {
    if (!now || now === prev) return;
    const latest = cartStore.items[cartStore.items.length - 1];
    if (latest) lastItemName.value = latest.name;
    bouncing.value = true;
    showToast.value = true;
    setTimeout(() => (bouncing.value = false), 600);
    setTimeout(() => (showToast.value = false), 2200);
  },
);
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
          <NuxtLink
            :to="`/t/${tenantSlug}/cart`"
            :class="[
              'relative w-10 h-10 rounded-lg border bg-card hover:bg-muted transition-all flex items-center justify-center cursor-pointer',
              bouncing ? 'animate-bounce border-primary' : '',
            ]"
            :aria-label="`Carrito (${cartStore.totalItems} items)`"
          >
            <ShoppingCart class="w-4 h-4 text-foreground" />
            <span
              v-if="cartStore.totalItems > 0"
              class="absolute -top-1.5 -right-1.5 min-w-5 h-5 px-1 rounded-full bg-primary text-white text-[10px] font-bold flex items-center justify-center shadow"
            >
              {{ cartStore.totalItems }}
            </span>
          </NuxtLink>
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

    <!-- Toast: "Agregado al carrito" -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-300 ease-out"
        enter-from-class="translate-y-4 opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-active-class="transition duration-200 ease-in"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="translate-y-4 opacity-0"
      >
        <div
          v-if="showToast"
          class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 bg-foreground text-background px-4 py-3 rounded-xl shadow-2xl flex items-center gap-3 max-w-[90vw]"
        >
          <div
            class="w-8 h-8 rounded-full bg-emerald-500 flex items-center justify-center shrink-0"
          >
            <Check class="w-4 h-4 text-white" />
          </div>
          <div class="text-sm">
            <p class="font-semibold">Agregado al carrito</p>
            <p class="text-xs opacity-80 truncate max-w-[200px]">
              {{ lastItemName }}
            </p>
          </div>
          <NuxtLink
            :to="`/t/${tenantSlug}/cart`"
            class="ml-2 text-xs font-bold text-primary-foreground bg-primary hover:opacity-90 px-3 py-1.5 rounded-lg shrink-0"
          >
            Ver carrito
          </NuxtLink>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
