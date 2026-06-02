<script setup lang="ts">
import {
  ShoppingCart,
  Trash2,
  Plus,
  Minus,
  ArrowLeft,
  ShieldCheck,
  Truck,
  RotateCcw,
  Sparkles,
} from "lucide-vue-next";
import { useCartStore } from "~/stores/cart";
import { useFormatPrice } from "~/composables/useFormatPrice";

definePageMeta({
  layout: "public",
});

const route = useRoute();
const tenantSlug = route.params.tenantSlug as string;
const cartStore = useCartStore();
const { format: fmt } = useFormatPrice("VE");

// Deterministic gradient (same as catalog/detail) so the thumb matches
function thumbGradient(id: string): string {
  const palettes = [
    "from-purple-200 via-violet-100 to-fuchsia-100",
    "from-blue-200 via-sky-100 to-cyan-100",
    "from-emerald-200 via-teal-100 to-cyan-100",
    "from-amber-200 via-orange-100 to-rose-100",
    "from-rose-200 via-pink-100 to-purple-100",
    "from-indigo-200 via-blue-100 to-cyan-100",
    "from-lime-200 via-emerald-100 to-teal-100",
    "from-fuchsia-200 via-rose-100 to-amber-100",
  ];
  const hash = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  return palettes[hash % palettes.length];
}

// Mock shipping/tax computations for visual UX. Real values land in S7-PR3.
const shippingCost = computed(() => (cartStore.subtotal > 100 ? 0 : 5));
const total = computed(() => cartStore.subtotal + shippingCost.value);

useSeoMeta({
  title: "Carrito · DaaS",
  description: "Revisá tu carrito de compras",
});
</script>

<template>
  <div class="bg-background min-h-[60vh]">
    <!-- Page header -->
    <div class="border-b bg-card/50">
      <div
        class="max-w-7xl mx-auto px-4 sm:px-6 py-5 flex items-center justify-between gap-4"
      >
        <div class="flex items-center gap-3">
          <NuxtLink
            :to="`/t/${tenantSlug}`"
            class="p-2 rounded-lg hover:bg-muted text-muted-foreground transition-colors"
            aria-label="Volver al catálogo"
          >
            <ArrowLeft class="w-4 h-4" />
          </NuxtLink>
          <div>
            <h1
              class="text-xl sm:text-2xl font-bold font-heading text-foreground flex items-center gap-2"
            >
              <ShoppingCart class="w-5 h-5 text-primary" />
              Tu carrito
            </h1>
            <p class="text-xs text-muted-foreground">
              {{ cartStore.totalItems }}
              {{ cartStore.totalItems === 1 ? "artículo" : "artículos" }}
            </p>
          </div>
        </div>

        <button
          v-if="!cartStore.isEmpty"
          type="button"
          class="text-xs text-muted-foreground hover:text-destructive transition-colors cursor-pointer hidden sm:inline"
          @click="cartStore.clearCart()"
        >
          Vaciar carrito
        </button>
      </div>
    </div>

    <!-- Empty state -->
    <div
      v-if="cartStore.isEmpty"
      class="max-w-7xl mx-auto px-4 sm:px-6 py-20 text-center"
    >
      <div
        class="inline-flex items-center justify-center w-20 h-20 rounded-full bg-primary/10 mb-5"
      >
        <ShoppingCart class="w-10 h-10 text-primary" />
      </div>
      <h2 class="text-xl font-bold text-foreground mb-2">
        Tu carrito está vacío
      </h2>
      <p class="text-sm text-muted-foreground mb-6 max-w-md mx-auto">
        Aún no agregaste productos. Explorá el catálogo y encontrá algo especial
        para vos.
      </p>
      <NuxtLink
        :to="`/t/${tenantSlug}`"
        class="inline-flex items-center gap-2 px-5 py-3 rounded-xl bg-primary text-white text-sm font-bold hover:opacity-90 transition-opacity"
      >
        Explorar productos
      </NuxtLink>
    </div>

    <!-- Cart layout -->
    <div
      v-else
      class="max-w-7xl mx-auto px-4 sm:px-6 py-6 grid grid-cols-1 lg:grid-cols-[1fr_360px] gap-6 lg:gap-8 pb-32 lg:pb-8"
    >
      <!-- Line items -->
      <section class="space-y-3">
        <div
          v-for="item in cartStore.items"
          :key="item.productId"
          class="border bg-card rounded-xl overflow-hidden p-3 sm:p-4 flex gap-3 sm:gap-4"
        >
          <NuxtLink
            :to="`/t/${tenantSlug}/product/${item.productId}`"
            :class="[
              'w-20 h-20 sm:w-28 sm:h-28 rounded-lg shrink-0 bg-gradient-to-br relative overflow-hidden',
              thumbGradient(item.productId),
            ]"
          >
            <div class="absolute inset-0 flex items-center justify-center">
              <ShoppingCart class="w-8 h-8 text-foreground/15" />
            </div>
          </NuxtLink>

          <div class="flex-1 min-w-0 flex flex-col gap-2">
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <p
                  v-if="item.category"
                  class="text-[10px] uppercase tracking-wider text-muted-foreground font-semibold mb-0.5"
                >
                  {{ item.category }}
                </p>
                <NuxtLink
                  :to="`/t/${tenantSlug}/product/${item.productId}`"
                  class="text-sm sm:text-base font-semibold text-foreground hover:text-primary transition-colors line-clamp-2"
                >
                  {{ item.name }}
                </NuxtLink>
                <p
                  v-if="item.is_fiscal"
                  class="text-[10px] text-muted-foreground mt-0.5"
                >
                  Fiscal · IVA incluido
                </p>
              </div>

              <button
                type="button"
                class="p-2 rounded-lg hover:bg-destructive/10 hover:text-destructive text-muted-foreground transition-colors cursor-pointer shrink-0"
                :aria-label="`Eliminar ${item.name}`"
                @click="cartStore.removeItem(item.productId)"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>

            <div class="flex items-end justify-between gap-3 mt-auto flex-wrap">
              <div class="inline-flex items-center border rounded-lg">
                <button
                  type="button"
                  :disabled="item.qty <= 1"
                  class="w-8 h-8 flex items-center justify-center hover:bg-muted disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
                  aria-label="Disminuir cantidad"
                  @click="cartStore.updateQty(item.productId, item.qty - 1)"
                >
                  <Minus class="w-3 h-3" />
                </button>
                <span class="w-9 text-center text-sm font-semibold">{{
                  item.qty
                }}</span>
                <button
                  type="button"
                  :disabled="!!item.stock_qty && item.qty >= item.stock_qty"
                  class="w-8 h-8 flex items-center justify-center hover:bg-muted disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
                  aria-label="Aumentar cantidad"
                  @click="cartStore.updateQty(item.productId, item.qty + 1)"
                >
                  <Plus class="w-3 h-3" />
                </button>
              </div>

              <div class="text-right">
                <p
                  class="text-base sm:text-lg font-bold font-mono text-foreground"
                >
                  {{ fmt(item.price * item.qty) }}
                </p>
                <p
                  v-if="item.qty > 1"
                  class="text-[10px] text-muted-foreground"
                >
                  {{ fmt(item.price) }} c/u
                </p>
              </div>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-between pt-2">
          <NuxtLink
            :to="`/t/${tenantSlug}`"
            class="inline-flex items-center gap-2 text-sm text-primary hover:underline font-semibold"
          >
            <ArrowLeft class="w-4 h-4" />
            Seguir comprando
          </NuxtLink>
          <button
            type="button"
            class="text-xs text-muted-foreground hover:text-destructive transition-colors cursor-pointer sm:hidden"
            @click="cartStore.clearCart()"
          >
            Vaciar carrito
          </button>
        </div>
      </section>

      <!-- Order summary (sticky on desktop) -->
      <aside>
        <div class="lg:sticky lg:top-20 space-y-4">
          <div class="border bg-card rounded-xl p-5 shadow-sm">
            <h2 class="text-base font-bold text-foreground mb-4">
              Resumen del pedido
            </h2>

            <dl class="space-y-2.5 text-sm">
              <div class="flex items-center justify-between">
                <dt class="text-muted-foreground">
                  Subtotal ({{ cartStore.totalItems }}
                  {{ cartStore.totalItems === 1 ? "artículo" : "artículos" }})
                </dt>
                <dd class="font-semibold text-foreground font-mono">
                  {{ fmt(cartStore.subtotal) }}
                </dd>
              </div>
              <div class="flex items-center justify-between">
                <dt class="text-muted-foreground">Envío</dt>
                <dd
                  :class="[
                    'font-semibold font-mono',
                    shippingCost === 0 ? 'text-emerald-600' : 'text-foreground',
                  ]"
                >
                  {{ shippingCost === 0 ? "Gratis" : fmt(shippingCost) }}
                </dd>
              </div>
              <div
                v-if="shippingCost > 0"
                class="flex items-center gap-1.5 text-[11px] text-muted-foreground bg-amber-50 border border-amber-100 rounded-md px-2.5 py-1.5 mt-1"
              >
                <Sparkles class="w-3 h-3 text-amber-500" />
                Agregá {{ fmt(100 - cartStore.subtotal) }} más y el envío es
                gratis
              </div>
              <div class="h-px bg-border my-3" />
              <div class="flex items-center justify-between text-base">
                <dt class="font-bold text-foreground">Total</dt>
                <dd class="font-bold text-foreground font-mono text-xl">
                  {{ fmt(total) }}
                </dd>
              </div>
            </dl>

            <NuxtLink
              :to="`/t/${tenantSlug}/checkout`"
              class="w-full mt-5 px-5 py-3.5 rounded-xl bg-primary text-white text-sm font-bold hover:opacity-90 transition-opacity flex items-center justify-center gap-2 cursor-pointer shadow-lg shadow-primary/20"
            >
              <ShoppingCart class="w-4 h-4" />
              Continuar con la compra
            </NuxtLink>

            <p class="text-[11px] text-center text-muted-foreground mt-2.5">
              Pago seguro · Sin compromiso hasta confirmar
            </p>
          </div>

          <!-- Trust footer -->
          <div class="border bg-card rounded-xl p-4 space-y-3">
            <div class="flex items-center gap-3 text-xs text-muted-foreground">
              <div
                class="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center shrink-0"
              >
                <Truck class="w-4 h-4 text-primary" />
              </div>
              <div>
                <p class="font-semibold text-foreground">Envío rápido</p>
                <p>1-3 días hábiles</p>
              </div>
            </div>
            <div class="flex items-center gap-3 text-xs text-muted-foreground">
              <div
                class="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center shrink-0"
              >
                <ShieldCheck class="w-4 h-4 text-primary" />
              </div>
              <div>
                <p class="font-semibold text-foreground">Compra protegida</p>
                <p>Tus datos están seguros</p>
              </div>
            </div>
            <div class="flex items-center gap-3 text-xs text-muted-foreground">
              <div
                class="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center shrink-0"
              >
                <RotateCcw class="w-4 h-4 text-primary" />
              </div>
              <div>
                <p class="font-semibold text-foreground">Devolución gratis</p>
                <p>Hasta 7 días después de recibir</p>
              </div>
            </div>
          </div>
        </div>
      </aside>
    </div>

    <!-- Sticky mobile checkout bar -->
    <div
      v-if="!cartStore.isEmpty"
      class="fixed bottom-0 inset-x-0 lg:hidden bg-card border-t shadow-lg z-30 px-4 py-3 safe-bottom"
    >
      <div class="flex items-center gap-3">
        <div class="flex-1">
          <p class="text-[11px] text-muted-foreground">Total</p>
          <p class="text-lg font-bold font-mono text-foreground">
            {{ fmt(total) }}
          </p>
        </div>
        <NuxtLink
          :to="`/t/${tenantSlug}/checkout`"
          class="flex-1 px-4 py-3 rounded-xl bg-primary text-white text-sm font-bold flex items-center justify-center gap-1.5 cursor-pointer"
        >
          Comprar
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
