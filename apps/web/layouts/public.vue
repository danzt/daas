<script setup lang="ts">
import { Package, ShoppingCart, Check, ChevronRight } from "lucide-vue-next";
import { useCartStore } from "~/stores/cart";
import { useBranding } from "~/composables/useBranding";
import { useFormatPrice } from "~/composables/useFormatPrice";

const route = useRoute();
const cartStore = useCartStore();

const tenantSlug = computed(() => (route.params.tenantSlug as string) ?? "");

// Branding — fetch once when the layout mounts.
const { storeName, logoURL, primaryColor, fetchBranding } =
  useBranding(tenantSlug);

// Inject CSS variable override for primary color when branding provides one.
// useHead is reactive — it re-runs whenever primaryColor changes.
useHead(
  computed(() => ({
    style: [
      {
        innerHTML: `:root { --primary: ${primaryColor.value}; }`,
        id: "tenant-primary-color",
      },
    ],
  })),
);

const { format: fmt } = useFormatPrice("VE");

// Hydrate cart + fetch branding on client mount.
onMounted(() => {
  cartStore.restore();
  fetchBranding();
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
          <!-- Logo: show image if set, else fallback to icon -->
          <div
            class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors overflow-hidden"
          >
            <img
              v-if="logoURL"
              :src="logoURL"
              :alt="storeName"
              class="w-full h-full object-contain"
            />
            <Package v-else class="w-5 h-5 text-primary" />
          </div>
          <div>
            <h1
              class="text-base font-bold font-heading text-foreground capitalize"
            >
              {{ storeName }}
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

    <!-- Mobile sticky cart bar — appears when cart has items -->
    <Transition
      enter-active-class="transition-transform duration-300 ease-out"
      enter-from-class="translate-y-full"
      enter-to-class="translate-y-0"
      leave-active-class="transition-transform duration-200 ease-in"
      leave-from-class="translate-y-0"
      leave-to-class="translate-y-full"
    >
      <div
        v-if="cartStore.totalItems > 0"
        class="fixed bottom-0 inset-x-0 z-40 sm:hidden border-t border-border bg-card/95 backdrop-blur px-4 pt-3"
        style="padding-bottom: calc(0.75rem + env(safe-area-inset-bottom))"
      >
        <NuxtLink
          :to="`/t/${tenantSlug}/cart`"
          class="flex h-14 w-full items-center justify-between rounded-2xl bg-primary px-4 text-primary-foreground shadow-lg shadow-primary/25"
        >
          <div class="flex items-center gap-2.5">
            <span
              class="flex size-7 items-center justify-center rounded-full bg-white/20 text-xs font-bold"
            >
              {{ cartStore.totalItems }}
            </span>
            <span class="text-sm font-semibold">Ver carrito</span>
          </div>
          <div class="flex items-center gap-1">
            <span class="font-mono font-bold">{{
              fmt(cartStore.subtotal)
            }}</span>
            <ChevronRight class="size-4 opacity-70" />
          </div>
        </NuxtLink>
      </div>
    </Transition>

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
