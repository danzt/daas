<script setup lang="ts">
import { Package, Plus, AlertTriangle, Search } from "lucide-vue-next";
import type {
  Product,
  Category,
} from "~/components/products/ProductFormModal.vue";

/**
 * MobileScreensProducts — presentational mobile product catalog.
 * Owns its own search + type filter (mobile UX, independent of web filters).
 * Emits row/create actions back to the page which holds the modal + handlers.
 */
const props = defineProps<{
  products: Product[];
  categories: Category[];
  loading: boolean;
  loadError: string;
  isOwner: boolean;
}>();

const emit = defineEmits<{
  create: [];
  edit: [product: Product];
  retry: [];
}>();

const search = ref("");
const typeFilter = ref("all");

const typeOptions = computed(() => [
  { key: "all", label: "Todos", count: props.products.length },
  {
    key: "fiscal",
    label: "Fiscal",
    count: props.products.filter((p) => p.is_fiscal).length,
  },
  {
    key: "interno",
    label: "Interno",
    count: props.products.filter((p) => !p.is_fiscal).length,
  },
]);

const filtered = computed(() => {
  let list = props.products;
  const q = search.value.trim().toLowerCase();
  if (q) {
    list = list.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        (p.sku ?? "").toLowerCase().includes(q),
    );
  }
  if (typeFilter.value === "fiscal") list = list.filter((p) => p.is_fiscal);
  else if (typeFilter.value === "interno")
    list = list.filter((p) => !p.is_fiscal);
  return list;
});

function categoryName(id?: string) {
  if (!id) return "Sin categoría";
  return props.categories.find((c) => c.id === id)?.name ?? "Sin categoría";
}

function priceOf(p: Product) {
  const price = p.is_fiscal ? p.fiscal_price : p.internal_price;
  if (price == null) return "—";
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(price);
}
</script>

<template>
  <MobileScreen title="Productos" :subtitle="`${products.length} en catálogo`">
    <template #action>
      <button
        v-if="isOwner"
        type="button"
        aria-label="Nuevo producto"
        class="no-min-tap flex size-10 items-center justify-center rounded-full text-primary active:bg-accent"
        @click="emit('create')"
      >
        <Plus class="size-6" />
      </button>
    </template>

    <!-- Search + filter -->
    <MobileSearchBar v-model="search" placeholder="Buscar por nombre o SKU…" />
    <MobileSegment v-model="typeFilter" :options="typeOptions" />

    <!-- Error -->
    <div v-if="loadError" class="px-4 py-12 text-center">
      <AlertTriangle class="mx-auto mb-3 size-10 text-destructive" />
      <p class="font-medium text-destructive">{{ loadError }}</p>
      <button
        type="button"
        class="mt-4 h-10 rounded-xl border border-border px-5 text-sm font-semibold text-muted-foreground active:bg-accent"
        @click="emit('retry')"
      >
        Reintentar
      </button>
    </div>

    <!-- Loading -->
    <div v-else-if="loading && !products.length" class="space-y-2 px-4 pt-2">
      <div
        v-for="n in 7"
        :key="n"
        class="h-[72px] animate-pulse rounded-2xl bg-muted"
      />
    </div>

    <!-- Empty -->
    <div v-else-if="!products.length" class="px-6 py-16 text-center">
      <Package class="mx-auto mb-4 size-12 text-muted-foreground/20" />
      <h3 class="mb-1 text-base font-semibold text-muted-foreground">
        Sin productos aún
      </h3>
      <p class="text-sm text-muted-foreground">
        Agregá tu primer producto para comenzar
      </p>
    </div>

    <!-- Filtered empty -->
    <div v-else-if="!filtered.length" class="px-6 py-12 text-center">
      <Search class="mx-auto mb-3 size-10 text-muted-foreground/20" />
      <p class="font-medium text-muted-foreground">Sin resultados</p>
    </div>

    <!-- Product list -->
    <div v-else class="divide-y divide-border border-t border-border bg-white">
      <MobileListItem
        v-for="product in filtered"
        :key="product.id"
        :chevron="false"
        :class="!product.active ? 'opacity-55' : ''"
        @click="emit('edit', product)"
      >
        <template #leading>
          <div class="relative">
            <div
              class="size-14 overflow-hidden rounded-xl border border-border bg-muted"
            >
              <img
                v-if="product.image_url"
                :src="product.image_url"
                :alt="product.name"
                class="size-full object-cover"
              />
              <div
                v-else
                class="flex size-full items-center justify-center bg-gradient-to-br from-primary/5 to-primary/15"
              >
                <Package class="size-6 text-primary/40" />
              </div>
            </div>
            <span
              v-if="product.is_fiscal"
              class="absolute -left-1 -top-1 rounded bg-primary px-1 py-px text-[8px] font-bold leading-none text-white"
              >F</span
            >
          </div>
        </template>

        <p class="truncate text-sm font-semibold leading-tight text-foreground">
          {{ product.name }}
        </p>
        <div class="mt-1 flex items-center gap-1.5">
          <span
            v-if="product.sku"
            class="rounded bg-muted px-1.5 py-px font-mono text-[10px] text-muted-foreground"
          >
            {{ product.sku }}
          </span>
          <span class="truncate text-[11px] text-muted-foreground">
            {{ categoryName(product.category_id) }}
          </span>
        </div>

        <template #trailing>
          <div class="flex flex-col items-end">
            <span
              class="font-mono text-sm font-bold tabular-nums text-foreground"
            >
              {{ priceOf(product) }}
            </span>
            <span
              v-if="product.is_fiscal && product.tax_rate != null"
              class="text-[10px] text-muted-foreground"
              >+{{ product.tax_rate }}% IVA</span
            >
            <span
              v-if="!product.active"
              class="mt-0.5 rounded bg-muted px-1.5 py-px text-[9px] font-bold uppercase text-muted-foreground"
              >Inactivo</span
            >
          </div>
        </template>
      </MobileListItem>
    </div>

    <!-- FAB -->
    <MobileFab v-if="isOwner" label="Nuevo" @click="emit('create')" />
  </MobileScreen>
</template>
