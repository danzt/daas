<script setup lang="ts">
import {
  Search,
  ShoppingCart,
  Star,
  Filter,
  X,
  ChevronDown,
  ChevronRight,
  SlidersHorizontal,
  TrendingUp,
  Sparkles,
  Tag,
  FolderOpen,
  Eye,
  Package,
  Truck,
  ShieldCheck,
} from "lucide-vue-next";
import { usePublicFetch } from "~/composables/usePublicFetch";
import { useFormatPrice } from "~/composables/useFormatPrice";
import { useCartStore } from "~/stores/cart";

const cartStore = useCartStore();

function addToCart(product: ShopProduct) {
  if (product.stock_qty === 0) return;
  cartStore.addItem({
    productId: product.id,
    name: product.name,
    price: effectivePrice(product),
    qty: 1,
    category: product.category,
    is_fiscal: product.is_fiscal,
    stock_qty: product.stock_qty,
  });
}

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

const allProducts = ref<ShopProduct[]>([]);
const loading = ref(false);
const loadError = ref("");
const page = ref(1);
const perPage = 24;
const totalCount = ref(0);
const totalPages = ref(1);

const searchQuery = ref("");
const selectedCategory = ref<string | null>(null);
const selectedPriceRange = ref<string | null>(null);
const onlyInStock = ref(false);
const selectedType = ref<"all" | "fiscal" | "internal">("all");
const sortBy = ref<"featured" | "price-asc" | "price-desc" | "name">(
  "featured",
);
const mobileFiltersOpen = ref(false);

const { format: fmt } = useFormatPrice("VE");

const categories = computed(() => {
  const map = new Map<string, number>();
  allProducts.value.forEach((p) => {
    if (p.category) map.set(p.category, (map.get(p.category) ?? 0) + 1);
  });
  return Array.from(map.entries()).map(([name, count]) => ({ name, count }));
});

const priceRanges = [
  { label: "Menos de Bs.S 10", value: "0-10", min: 0, max: 10 },
  { label: "Bs.S 10 a 100", value: "10-100", min: 10, max: 100 },
  { label: "Bs.S 100 a 500", value: "100-500", min: 100, max: 500 },
  { label: "Más de Bs.S 500", value: "500+", min: 500, max: Infinity },
];

function effectivePrice(p: ShopProduct): number {
  return p.is_fiscal && p.fiscal_price ? p.fiscal_price : p.price;
}

const filteredProducts = computed(() => {
  let list = [...allProducts.value];

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.description?.toLowerCase().includes(q) ||
        p.category?.toLowerCase().includes(q),
    );
  }

  if (selectedCategory.value) {
    list = list.filter((p) => p.category === selectedCategory.value);
  }

  if (selectedPriceRange.value) {
    const range = priceRanges.find((r) => r.value === selectedPriceRange.value);
    if (range) {
      list = list.filter((p) => {
        const price = effectivePrice(p);
        return price >= range.min && price < range.max;
      });
    }
  }

  if (onlyInStock.value) {
    list = list.filter((p) => p.stock_qty > 0);
  }

  if (selectedType.value === "fiscal") {
    list = list.filter((p) => p.is_fiscal);
  } else if (selectedType.value === "internal") {
    list = list.filter((p) => !p.is_fiscal);
  }

  if (sortBy.value === "price-asc") {
    list.sort((a, b) => effectivePrice(a) - effectivePrice(b));
  } else if (sortBy.value === "price-desc") {
    list.sort((a, b) => effectivePrice(b) - effectivePrice(a));
  } else if (sortBy.value === "name") {
    list.sort((a, b) => a.name.localeCompare(b.name));
  }

  return list;
});

function activeFiltersCount(): number {
  let n = 0;
  if (selectedCategory.value) n++;
  if (selectedPriceRange.value) n++;
  if (onlyInStock.value) n++;
  if (selectedType.value !== "all") n++;
  return n;
}

function clearFilters() {
  selectedCategory.value = null;
  selectedPriceRange.value = null;
  onlyInStock.value = false;
  selectedType.value = "all";
  searchQuery.value = "";
}

// Deterministic gradient per product (visual identity for placeholder cards)
function productGradient(id: string): string {
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

function productRating(id: string): { stars: number; count: number } {
  const hash = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  const stars = 3.8 + (hash % 12) / 10;
  const count = 47 + (hash % 400);
  return { stars: Math.round(stars * 10) / 10, count };
}

function productSoldCount(id: string): number {
  const hash = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  return 12 + (hash % 80);
}

function isBestSeller(p: ShopProduct, index: number): boolean {
  return index === 0 && filteredProducts.value.length > 0 && p.stock_qty > 0;
}

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    const resp = await usePublicFetch<ListResponse>(
      `/products?page=${page.value}&per_page=${perPage}`,
    );
    allProducts.value = resp.data ?? [];
    totalCount.value = resp.meta?.total_count ?? 0;
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
  <div class="bg-background">
    <!-- ── Hero Search ──────────────────────────────────────────────────── -->
    <div
      class="border-b bg-gradient-to-br from-primary/5 via-purple-50 to-fuchsia-50/50"
    >
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 sm:py-12">
        <div class="text-center mb-6">
          <h1
            class="text-2xl sm:text-4xl font-bold font-heading text-foreground mb-2"
          >
            Descubrí productos para vos
          </h1>
          <p class="text-sm sm:text-base text-muted-foreground">
            Envío rápido · Pago seguro · Garantía de satisfacción
          </p>
        </div>

        <div class="max-w-3xl mx-auto relative">
          <Search
            class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground pointer-events-none"
          />
          <input
            v-model="searchQuery"
            type="search"
            class="w-full h-12 sm:h-14 pl-12 pr-4 rounded-xl border-2 border-primary/20 bg-white text-base shadow-sm focus:outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all"
            placeholder="Buscar productos, marcas y más..."
          />
          <button
            v-if="searchQuery"
            type="button"
            aria-label="Limpiar búsqueda"
            class="absolute right-3 top-1/2 -translate-y-1/2 p-1.5 rounded-lg hover:bg-muted text-muted-foreground cursor-pointer transition-colors"
            @click="searchQuery = ''"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <div
          class="flex items-center justify-center gap-4 sm:gap-6 mt-6 text-xs sm:text-sm text-muted-foreground flex-wrap"
        >
          <div class="flex items-center gap-1.5">
            <Truck class="w-4 h-4 text-primary" /> Envío rápido
          </div>
          <div class="flex items-center gap-1.5">
            <ShieldCheck class="w-4 h-4 text-primary" /> Compra protegida
          </div>
          <div class="flex items-center gap-1.5">
            <Package class="w-4 h-4 text-primary" /> Devoluciones gratis
          </div>
        </div>
      </div>
    </div>

    <!-- ── Main: Sidebar + Grid ─────────────────────────────────────────── -->
    <div class="max-w-7xl mx-auto px-4 sm:px-6 py-6">
      <!-- Toolbar -->
      <div class="flex items-center justify-between gap-3 mb-5 flex-wrap">
        <div class="flex items-center gap-2 text-sm text-muted-foreground">
          <NuxtLink
            :to="`/t/${$route.params.tenantSlug}`"
            class="hover:text-primary transition-colors"
          >
            Inicio
          </NuxtLink>
          <ChevronRight class="w-3.5 h-3.5" />
          <span class="text-foreground font-medium">
            {{ selectedCategory ?? "Todos los productos" }}
          </span>
          <span v-if="!loading" class="text-muted-foreground">
            · {{ filteredProducts.length }} resultados
          </span>
        </div>

        <div class="flex items-center gap-2">
          <button
            type="button"
            class="lg:hidden inline-flex items-center gap-2 px-3 py-2 rounded-lg border bg-card hover:bg-muted transition-colors text-sm font-medium cursor-pointer"
            @click="mobileFiltersOpen = true"
          >
            <SlidersHorizontal class="w-4 h-4" />
            Filtros
            <span
              v-if="activeFiltersCount() > 0"
              class="inline-flex items-center justify-center min-w-5 h-5 px-1.5 rounded-full bg-primary text-white text-[10px] font-bold"
            >
              {{ activeFiltersCount() }}
            </span>
          </button>

          <div class="relative">
            <select
              v-model="sortBy"
              class="appearance-none pl-3 pr-9 py-2 rounded-lg border bg-card text-sm font-medium hover:bg-muted transition-colors focus:outline-none focus:border-primary cursor-pointer"
            >
              <option value="featured">Destacados</option>
              <option value="price-asc">Precio: menor a mayor</option>
              <option value="price-desc">Precio: mayor a menor</option>
              <option value="name">Nombre A-Z</option>
            </select>
            <ChevronDown
              class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 pointer-events-none text-muted-foreground"
            />
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-[260px_1fr] gap-6">
        <!-- Sidebar (desktop) -->
        <aside class="hidden lg:block">
          <div class="sticky top-20 space-y-5">
            <div
              v-if="activeFiltersCount() > 0"
              class="flex items-center justify-between p-3 rounded-xl bg-primary/5 border border-primary/20"
            >
              <div class="flex items-center gap-2">
                <Filter class="w-4 h-4 text-primary" />
                <span class="text-sm font-medium"
                  >{{ activeFiltersCount() }} filtros</span
                >
              </div>
              <button
                type="button"
                class="text-xs text-primary hover:underline font-medium cursor-pointer"
                @click="clearFilters"
              >
                Limpiar
              </button>
            </div>

            <div class="rounded-xl border bg-card p-4">
              <h3
                class="text-sm font-bold text-foreground mb-3 flex items-center gap-2"
              >
                <Tag class="w-4 h-4 text-primary" />
                Categorías
              </h3>
              <div class="space-y-1.5">
                <button
                  type="button"
                  :class="[
                    'w-full text-left px-2 py-1.5 rounded-md text-sm transition-colors flex items-center justify-between cursor-pointer',
                    selectedCategory === null
                      ? 'bg-primary/10 text-primary font-medium'
                      : 'hover:bg-muted text-muted-foreground',
                  ]"
                  @click="selectedCategory = null"
                >
                  <span>Todas</span>
                  <span class="text-xs">{{ allProducts.length }}</span>
                </button>
                <button
                  v-for="cat in categories"
                  :key="cat.name"
                  type="button"
                  :class="[
                    'w-full text-left px-2 py-1.5 rounded-md text-sm transition-colors flex items-center justify-between cursor-pointer',
                    selectedCategory === cat.name
                      ? 'bg-primary/10 text-primary font-medium'
                      : 'hover:bg-muted text-muted-foreground',
                  ]"
                  @click="selectedCategory = cat.name"
                >
                  <span>{{ cat.name }}</span>
                  <span class="text-xs">{{ cat.count }}</span>
                </button>
              </div>
            </div>

            <div class="rounded-xl border bg-card p-4">
              <h3 class="text-sm font-bold text-foreground mb-3">Precio</h3>
              <div class="space-y-2">
                <label
                  v-for="range in priceRanges"
                  :key="range.value"
                  class="flex items-center gap-2 cursor-pointer text-sm text-muted-foreground hover:text-foreground transition-colors"
                >
                  <input
                    v-model="selectedPriceRange"
                    type="radio"
                    :value="range.value"
                    class="accent-primary cursor-pointer"
                  />
                  {{ range.label }}
                </label>
                <button
                  v-if="selectedPriceRange"
                  type="button"
                  class="text-xs text-primary hover:underline mt-1 cursor-pointer"
                  @click="selectedPriceRange = null"
                >
                  Limpiar precio
                </button>
              </div>
            </div>

            <div class="rounded-xl border bg-card p-4">
              <h3 class="text-sm font-bold text-foreground mb-3">
                Tipo de factura
              </h3>
              <div class="space-y-2">
                <label
                  class="flex items-center gap-2 cursor-pointer text-sm text-muted-foreground hover:text-foreground transition-colors"
                >
                  <input
                    v-model="selectedType"
                    type="radio"
                    value="all"
                    class="accent-primary cursor-pointer"
                  />
                  Todos
                </label>
                <label
                  class="flex items-center gap-2 cursor-pointer text-sm text-muted-foreground hover:text-foreground transition-colors"
                >
                  <input
                    v-model="selectedType"
                    type="radio"
                    value="fiscal"
                    class="accent-primary cursor-pointer"
                  />
                  Fiscal (con IVA)
                </label>
                <label
                  class="flex items-center gap-2 cursor-pointer text-sm text-muted-foreground hover:text-foreground transition-colors"
                >
                  <input
                    v-model="selectedType"
                    type="radio"
                    value="internal"
                    class="accent-primary cursor-pointer"
                  />
                  Interno
                </label>
              </div>
            </div>

            <div class="rounded-xl border bg-card p-4">
              <label class="flex items-center justify-between cursor-pointer">
                <span class="text-sm font-medium text-foreground"
                  >Solo en stock</span
                >
                <span
                  :class="[
                    'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
                    onlyInStock ? 'bg-primary' : 'bg-muted',
                  ]"
                >
                  <input
                    v-model="onlyInStock"
                    type="checkbox"
                    class="sr-only"
                  />
                  <span
                    :class="[
                      'inline-block h-4 w-4 rounded-full bg-white shadow transition-transform',
                      onlyInStock ? 'translate-x-6' : 'translate-x-1',
                    ]"
                  />
                </span>
              </label>
            </div>
          </div>
        </aside>

        <!-- Grid -->
        <div>
          <div
            v-if="loading && allProducts.length === 0"
            class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 gap-3 sm:gap-4"
          >
            <div
              v-for="n in 8"
              :key="n"
              class="border bg-card rounded-xl overflow-hidden"
            >
              <div class="aspect-square bg-muted animate-pulse" />
              <div class="p-3 space-y-2">
                <div class="h-3 bg-muted rounded animate-pulse" />
                <div class="h-3 w-2/3 bg-muted rounded animate-pulse" />
                <div class="h-5 bg-muted rounded animate-pulse mt-2" />
              </div>
            </div>
          </div>

          <div
            v-else-if="loadError"
            class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-xl px-4 py-6 text-center"
          >
            {{ loadError }}
          </div>

          <div
            v-else-if="filteredProducts.length === 0"
            class="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3 border-2 border-dashed border-border rounded-xl"
          >
            <FolderOpen class="w-16 h-16 opacity-20" />
            <p class="text-base font-semibold text-foreground">
              No encontramos productos
            </p>
            <p class="text-sm">
              Probá con otros filtros o términos de búsqueda
            </p>
            <button
              v-if="activeFiltersCount() > 0 || searchQuery"
              type="button"
              class="mt-2 px-4 py-2 rounded-lg bg-primary text-white text-sm font-semibold hover:opacity-90 transition-opacity cursor-pointer"
              @click="clearFilters"
            >
              Limpiar todos los filtros
            </button>
          </div>

          <div
            v-else
            class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 gap-3 sm:gap-4"
          >
            <NuxtLink
              v-for="(product, index) in filteredProducts"
              :key="product.id"
              :to="`/t/${$route.params.tenantSlug}/product/${product.id}`"
              class="group relative border bg-card rounded-xl overflow-hidden hover:shadow-xl hover:shadow-primary/10 hover:border-primary/30 hover:-translate-y-0.5 transition-all duration-300 cursor-pointer flex flex-col"
            >
              <div
                :class="[
                  'aspect-square bg-gradient-to-br relative overflow-hidden',
                  productGradient(product.id),
                ]"
              >
                <div class="absolute inset-0 flex items-center justify-center">
                  <ShoppingCart
                    class="w-16 h-16 text-foreground/15 group-hover:scale-110 transition-transform duration-300"
                  />
                </div>

                <div class="absolute top-2 left-2 flex flex-col gap-1.5">
                  <span
                    v-if="isBestSeller(product, index)"
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-amber-500 text-white text-[10px] font-bold shadow-sm"
                  >
                    <TrendingUp class="w-2.5 h-2.5" />
                    BEST SELLER
                  </span>
                  <span
                    v-if="product.is_fiscal"
                    class="inline-flex items-center px-2 py-0.5 rounded-full bg-primary text-white text-[10px] font-bold shadow-sm"
                  >
                    FISCAL
                  </span>
                </div>

                <span
                  v-if="product.stock_qty > 0 && product.stock_qty <= 5"
                  class="absolute top-2 right-2 inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-amber-100 text-amber-700 text-[10px] font-semibold shadow-sm"
                >
                  <Sparkles class="w-2.5 h-2.5" />
                  ¡Pocos!
                </span>

                <div
                  v-if="product.stock_qty === 0"
                  class="absolute inset-0 bg-foreground/40 backdrop-blur-sm flex items-center justify-center"
                >
                  <span
                    class="px-3 py-1 rounded-full bg-card text-foreground text-xs font-bold shadow"
                    >AGOTADO</span
                  >
                </div>

                <div
                  class="absolute inset-x-0 bottom-0 p-2 opacity-0 group-hover:opacity-100 translate-y-2 group-hover:translate-y-0 transition-all duration-300"
                >
                  <div
                    class="bg-card/95 backdrop-blur rounded-lg px-3 py-1.5 text-xs font-semibold text-foreground text-center flex items-center justify-center gap-1.5"
                  >
                    <Eye class="w-3.5 h-3.5" />
                    Ver detalle
                  </div>
                </div>
              </div>

              <div class="p-3 sm:p-4 flex flex-col flex-1">
                <p
                  v-if="product.category"
                  class="text-[10px] uppercase tracking-wider text-muted-foreground font-semibold mb-1"
                >
                  {{ product.category }}
                </p>
                <h3
                  class="text-sm font-semibold text-foreground line-clamp-2 group-hover:text-primary transition-colors min-h-[2.5rem]"
                >
                  {{ product.name }}
                </h3>

                <div class="flex items-center gap-1 mt-1.5">
                  <div class="flex items-center">
                    <Star
                      v-for="i in 5"
                      :key="i"
                      :class="[
                        'w-3 h-3',
                        i <= Math.round(productRating(product.id).stars)
                          ? 'fill-amber-400 text-amber-400'
                          : 'text-muted-foreground/30',
                      ]"
                    />
                  </div>
                  <span class="text-[11px] text-muted-foreground">
                    {{ productRating(product.id).stars }} ({{
                      productRating(product.id).count
                    }})
                  </span>
                </div>

                <p
                  class="text-[10px] text-muted-foreground mt-1 hidden sm:block"
                >
                  {{ productSoldCount(product.id) }} vendidos este mes
                </p>

                <div class="flex-1" />

                <div class="flex items-baseline gap-2 mt-3">
                  <span
                    class="text-lg sm:text-xl font-bold font-mono text-foreground"
                  >
                    {{ fmt(effectivePrice(product)) }}
                  </span>
                </div>
                <p
                  v-if="product.is_fiscal"
                  class="text-[10px] text-muted-foreground mb-2"
                >
                  + IVA incluido
                </p>

                <button
                  type="button"
                  :disabled="product.stock_qty === 0"
                  class="w-full mt-2 px-3 py-2 rounded-lg bg-primary/10 hover:bg-primary hover:text-white disabled:opacity-50 disabled:cursor-not-allowed text-primary text-xs font-bold transition-all flex items-center justify-center gap-1.5 cursor-pointer disabled:hover:bg-primary/10 disabled:hover:text-primary"
                  :title="
                    product.stock_qty === 0
                      ? 'Producto agotado'
                      : 'Agregar al carrito'
                  "
                  @click.prevent.stop="addToCart(product)"
                >
                  <ShoppingCart class="w-3.5 h-3.5" />
                  Agregar
                </button>
              </div>
            </NuxtLink>
          </div>

          <div
            v-if="!loading && totalPages > 1"
            class="flex items-center justify-center gap-2 mt-10"
          >
            <button
              type="button"
              :disabled="page <= 1"
              class="px-4 py-2 text-sm rounded-lg border hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
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
              class="px-4 py-2 text-sm rounded-lg border hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              @click="
                page++;
                load();
              "
            >
              Siguiente
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Mobile filters drawer -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-200"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-200"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div v-if="mobileFiltersOpen" class="fixed inset-0 z-50 lg:hidden">
          <div
            class="absolute inset-0 bg-foreground/40 backdrop-blur-sm"
            @click="mobileFiltersOpen = false"
          />
          <div
            class="absolute inset-y-0 right-0 w-[90%] max-w-sm bg-card shadow-2xl flex flex-col"
          >
            <header
              class="flex items-center justify-between px-4 py-3 border-b"
            >
              <h2
                class="text-base font-bold text-foreground flex items-center gap-2"
              >
                <SlidersHorizontal class="w-4 h-4 text-primary" />
                Filtros
              </h2>
              <button
                type="button"
                aria-label="Cerrar filtros"
                class="p-2 rounded-lg hover:bg-muted text-muted-foreground cursor-pointer"
                @click="mobileFiltersOpen = false"
              >
                <X class="w-4 h-4" />
              </button>
            </header>

            <div class="flex-1 overflow-y-auto p-4 space-y-5">
              <div>
                <h3 class="text-sm font-bold text-foreground mb-2">
                  Categorías
                </h3>
                <div class="space-y-1">
                  <button
                    type="button"
                    :class="[
                      'w-full text-left px-3 py-2 rounded-md text-sm transition-colors flex items-center justify-between cursor-pointer',
                      selectedCategory === null
                        ? 'bg-primary/10 text-primary font-medium'
                        : 'hover:bg-muted',
                    ]"
                    @click="selectedCategory = null"
                  >
                    <span>Todas</span
                    ><span class="text-xs">{{ allProducts.length }}</span>
                  </button>
                  <button
                    v-for="cat in categories"
                    :key="cat.name"
                    type="button"
                    :class="[
                      'w-full text-left px-3 py-2 rounded-md text-sm transition-colors flex items-center justify-between cursor-pointer',
                      selectedCategory === cat.name
                        ? 'bg-primary/10 text-primary font-medium'
                        : 'hover:bg-muted',
                    ]"
                    @click="selectedCategory = cat.name"
                  >
                    <span>{{ cat.name }}</span
                    ><span class="text-xs">{{ cat.count }}</span>
                  </button>
                </div>
              </div>

              <div>
                <h3 class="text-sm font-bold text-foreground mb-2">Precio</h3>
                <div class="space-y-2">
                  <label
                    v-for="range in priceRanges"
                    :key="range.value"
                    class="flex items-center gap-2 cursor-pointer text-sm"
                  >
                    <input
                      v-model="selectedPriceRange"
                      type="radio"
                      :value="range.value"
                      class="accent-primary cursor-pointer"
                    />
                    {{ range.label }}
                  </label>
                </div>
              </div>

              <div>
                <h3 class="text-sm font-bold text-foreground mb-2">Tipo</h3>
                <div class="space-y-2">
                  <label class="flex items-center gap-2 cursor-pointer text-sm">
                    <input
                      v-model="selectedType"
                      type="radio"
                      value="all"
                      class="accent-primary cursor-pointer"
                    />
                    Todos
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer text-sm">
                    <input
                      v-model="selectedType"
                      type="radio"
                      value="fiscal"
                      class="accent-primary cursor-pointer"
                    />
                    Fiscal
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer text-sm">
                    <input
                      v-model="selectedType"
                      type="radio"
                      value="internal"
                      class="accent-primary cursor-pointer"
                    />
                    Interno
                  </label>
                </div>
              </div>

              <label
                class="flex items-center justify-between cursor-pointer pt-2 border-t"
              >
                <span class="text-sm font-medium">Solo en stock</span>
                <span
                  :class="[
                    'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
                    onlyInStock ? 'bg-primary' : 'bg-muted',
                  ]"
                >
                  <input
                    v-model="onlyInStock"
                    type="checkbox"
                    class="sr-only"
                  />
                  <span
                    :class="[
                      'inline-block h-4 w-4 rounded-full bg-white shadow transition-transform',
                      onlyInStock ? 'translate-x-6' : 'translate-x-1',
                    ]"
                  />
                </span>
              </label>
            </div>

            <footer class="border-t p-4 flex items-center gap-3">
              <button
                type="button"
                class="flex-1 px-4 py-2.5 rounded-lg border hover:bg-muted text-sm font-semibold cursor-pointer"
                @click="clearFilters"
              >
                Limpiar
              </button>
              <button
                type="button"
                class="flex-1 px-4 py-2.5 rounded-lg bg-primary text-white text-sm font-semibold hover:opacity-90 cursor-pointer"
                @click="mobileFiltersOpen = false"
              >
                Aplicar ({{ filteredProducts.length }})
              </button>
            </footer>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
