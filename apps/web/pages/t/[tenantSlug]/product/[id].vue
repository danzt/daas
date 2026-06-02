<script setup lang="ts">
import { ArrowLeft, ShoppingCart, Loader2 } from "lucide-vue-next";
import { usePublicFetch } from "~/composables/usePublicFetch";
import { useFormatPrice } from "~/composables/useFormatPrice";

definePageMeta({
  layout: "public",
});

const route = useRoute();
const productId = route.params.id as string;
const tenantSlug = route.params.tenantSlug as string;

interface ShopProduct {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  price: number;
  fiscal_price?: number;
  category?: string;
  stock_qty: number;
  is_fiscal: boolean;
}

const product = ref<ShopProduct | null>(null);
const loading = ref(false);
const notFound = ref(false);

const { format: fmt } = useFormatPrice("VE");

async function load() {
  loading.value = true;
  try {
    product.value = await usePublicFetch<ShopProduct>(`/products/${productId}`);
  } catch (err: unknown) {
    const e = err as { status?: number; statusCode?: number };
    if (e?.status === 404 || e?.statusCode === 404) {
      notFound.value = true;
    }
  } finally {
    loading.value = false;
  }
}

await load();

// SEO / OG meta (only when product loaded)
useSeoMeta({
  title: () => (product.value ? product.value.name : "Producto no encontrado"),
  description: () => product.value?.description?.slice(0, 160) ?? "",
  ogTitle: () => product.value?.name ?? "Producto",
  ogDescription: () => product.value?.description?.slice(0, 160) ?? "",
  ogImage: "/og-placeholder.png",
  ogType: "website",
});
</script>

<template>
  <div class="max-w-5xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
    <!-- Back link -->
    <NuxtLink
      :to="`/t/${tenantSlug}`"
      class="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-primary transition-colors mb-6"
    >
      <ArrowLeft class="w-4 h-4" />
      Volver al catálogo
    </NuxtLink>

    <!-- Loading -->
    <div
      v-if="loading"
      class="flex items-center justify-center py-20 text-muted-foreground"
    >
      <Loader2 class="w-6 h-6 animate-spin mr-2" />
      Cargando...
    </div>

    <!-- Not found -->
    <div v-else-if="notFound" class="text-center py-20">
      <h1 class="text-2xl font-bold text-foreground mb-2">
        Producto no encontrado
      </h1>
      <p class="text-muted-foreground mb-6">
        El producto que buscás no está disponible.
      </p>
      <NuxtLink
        :to="`/t/${tenantSlug}`"
        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-white text-sm font-semibold hover:opacity-90"
      >
        Volver al catálogo
      </NuxtLink>
    </div>

    <!-- Product detail -->
    <div
      v-else-if="product"
      class="grid grid-cols-1 md:grid-cols-2 gap-6 lg:gap-10"
    >
      <!-- Image -->
      <div class="border bg-card rounded-xl overflow-hidden">
        <div
          class="aspect-square bg-gradient-to-br from-primary/5 to-primary/10 flex items-center justify-center relative"
        >
          <ShoppingCart class="w-24 h-24 text-primary/20" />
          <span
            v-if="product.is_fiscal"
            class="absolute top-3 left-3 px-2.5 py-1 rounded-full bg-primary text-white text-xs font-semibold"
          >
            FISCAL
          </span>
        </div>
      </div>

      <!-- Info -->
      <div class="space-y-4">
        <div>
          <p
            v-if="product.category"
            class="text-xs uppercase tracking-wider text-muted-foreground mb-1"
          >
            {{ product.category }}
          </p>
          <h1
            class="text-2xl sm:text-3xl font-bold font-heading text-foreground"
          >
            {{ product.name }}
          </h1>
        </div>

        <div class="flex items-baseline gap-3">
          <p class="text-3xl sm:text-4xl font-bold font-mono text-foreground">
            {{
              fmt(
                product.is_fiscal && product.fiscal_price
                  ? product.fiscal_price
                  : product.price,
              )
            }}
          </p>
          <span v-if="product.is_fiscal" class="text-xs text-muted-foreground">
            + IVA
          </span>
        </div>

        <p
          v-if="product.stock_qty === 0"
          class="text-sm font-semibold text-rose-600"
        >
          Agotado
        </p>
        <p v-else-if="product.stock_qty <= 5" class="text-sm text-amber-600">
          Quedan solo {{ product.stock_qty }} unidades
        </p>
        <p v-else class="text-sm text-emerald-600">
          En stock — listo para enviar
        </p>

        <div v-if="product.description" class="border-t pt-4 mt-4">
          <h2 class="text-sm font-semibold text-foreground mb-2">
            Descripción
          </h2>
          <p class="text-sm text-muted-foreground whitespace-pre-line">
            {{ product.description }}
          </p>
        </div>

        <button
          type="button"
          disabled
          class="w-full px-5 py-3 rounded-lg bg-primary/20 text-primary font-semibold cursor-not-allowed flex items-center justify-center gap-2 mt-6"
          title="Carrito disponible próximamente"
        >
          <ShoppingCart class="w-4 h-4" />
          Agregar al carrito
        </button>
        <p class="text-xs text-center text-muted-foreground">
          Carrito disponible próximamente
        </p>
      </div>
    </div>
  </div>
</template>
