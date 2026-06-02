<script setup lang="ts">
import { FolderOpen, ShoppingCart } from "lucide-vue-next";
import { usePublicFetch } from "~/composables/usePublicFetch";
import { useFormatPrice } from "~/composables/useFormatPrice";

definePageMeta({
  layout: "public",
});

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

interface ListResponse {
  data: ShopProduct[];
  meta: {
    total_count: number;
    page: number;
    per_page: number;
    total_pages: number;
  };
}

const products = ref<ShopProduct[]>([]);
const loading = ref(false);
const loadError = ref("");
const page = ref(1);
const perPage = 20;
const totalPages = ref(1);

const { format: fmt } = useFormatPrice("VE");

function stockBadgeClass(qty: number): string {
  if (qty === 0)
    return "bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400";
  if (qty <= 5)
    return "bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400";
  return "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400";
}
function stockBadgeLabel(qty: number): string {
  if (qty === 0) return "Agotado";
  if (qty <= 5) return `${qty} en stock`;
  return "En stock";
}

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    const resp = await usePublicFetch<ListResponse>(
      `/products?page=${page.value}&per_page=${perPage}`,
    );
    products.value = resp.data ?? [];
    totalPages.value = resp.meta?.total_pages ?? 1;
  } catch {
    loadError.value = "No se pudieron cargar los productos";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
    <!-- Page heading -->
    <div class="mb-6">
      <h2 class="text-2xl sm:text-3xl font-bold font-heading text-foreground">
        Catálogo
      </h2>
      <p class="text-sm text-muted-foreground mt-1">
        Explorá nuestros productos
      </p>
    </div>

    <!-- Loading skeleton (8 cards) -->
    <div
      v-if="loading && products.length === 0"
      class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3 sm:gap-4"
    >
      <div
        v-for="n in 8"
        :key="n"
        class="border bg-card rounded-xl overflow-hidden"
      >
        <div class="aspect-square bg-muted animate-pulse" />
        <div class="p-3 sm:p-4 space-y-2">
          <div class="h-4 bg-muted rounded animate-pulse" />
          <div class="h-3 w-2/3 bg-muted rounded animate-pulse" />
          <div class="h-6 bg-muted rounded animate-pulse mt-3" />
        </div>
      </div>
    </div>

    <!-- Error -->
    <div
      v-else-if="loadError"
      class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3"
    >
      {{ loadError }}
    </div>

    <!-- Empty -->
    <div
      v-else-if="!loading && products.length === 0"
      class="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3"
    >
      <FolderOpen class="w-12 h-12 opacity-30" />
      <p class="text-sm">Aún no hay productos en esta tienda</p>
    </div>

    <!-- Product grid -->
    <div
      v-else
      class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3 sm:gap-4"
    >
      <NuxtLink
        v-for="product in products"
        :key="product.id"
        :to="`/t/${$route.params.tenantSlug}/product/${product.id}`"
        class="border bg-card rounded-xl overflow-hidden group hover:shadow-lg hover:border-primary/30 transition-all"
      >
        <!-- Image placeholder -->
        <div
          class="aspect-square bg-gradient-to-br from-primary/5 to-primary/10 flex items-center justify-center relative"
        >
          <ShoppingCart class="w-12 h-12 text-primary/30" />
          <span
            v-if="product.is_fiscal"
            class="absolute top-2 left-2 px-2 py-0.5 rounded-full bg-primary text-white text-[10px] font-semibold"
          >
            FISCAL
          </span>
        </div>

        <!-- Body -->
        <div class="p-3 sm:p-4 space-y-2">
          <h3
            class="text-sm font-semibold text-foreground line-clamp-2 group-hover:text-primary transition-colors"
          >
            {{ product.name }}
          </h3>
          <p v-if="product.category" class="text-xs text-muted-foreground">
            {{ product.category }}
          </p>
          <div class="flex items-end justify-between gap-2 pt-1">
            <p class="text-base sm:text-lg font-bold font-mono text-foreground">
              {{
                fmt(
                  product.is_fiscal && product.fiscal_price
                    ? product.fiscal_price
                    : product.price,
                )
              }}
            </p>
            <span
              :class="[
                'inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium whitespace-nowrap',
                stockBadgeClass(product.stock_qty),
              ]"
            >
              {{ stockBadgeLabel(product.stock_qty) }}
            </span>
          </div>
          <button
            type="button"
            disabled
            class="w-full mt-2 px-3 py-2 rounded-lg bg-primary/20 text-primary text-xs font-semibold cursor-not-allowed flex items-center justify-center gap-1.5"
            title="Carrito disponible próximamente"
          >
            <ShoppingCart class="w-3.5 h-3.5" />
            Agregar
          </button>
        </div>
      </NuxtLink>
    </div>

    <!-- Pagination (only if multiple pages) -->
    <div
      v-if="!loading && totalPages > 1"
      class="flex items-center justify-center gap-2 mt-8"
    >
      <button
        type="button"
        :disabled="page <= 1"
        class="px-3 py-1.5 text-sm rounded-lg border hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
        @click="
          page--;
          load();
        "
      >
        Anterior
      </button>
      <span class="text-sm text-muted-foreground px-3">
        Página {{ page }} de {{ totalPages }}
      </span>
      <button
        type="button"
        :disabled="page >= totalPages"
        class="px-3 py-1.5 text-sm rounded-lg border hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
        @click="
          page++;
          load();
        "
      >
        Siguiente
      </button>
    </div>
  </div>
</template>
