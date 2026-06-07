<script setup lang="ts">
import {
  Warehouse,
  AlertTriangle,
  TrendingDown,
  FolderOpen,
  SlidersHorizontal,
  History,
  ArrowUpDown,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import { useAuthStore } from "~/stores/auth";
import AdjustmentModal from "~/components/inventory/AdjustmentModal.vue";
import type {
  Product,
  Category,
} from "~/components/products/ProductFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

const store = useAuthStore();

// ─── Types ────────────────────────────────────────────────────────────────────

interface StockItem {
  product_id: string;
  tenant_id: string;
  quantity_on_hand: number;
  last_updated_at: string;
}

interface StockRow extends StockItem {
  product: Product | null;
}

// ─── State ────────────────────────────────────────────────────────────────────

const stockItems = ref<StockItem[]>([]);
const products = ref<Product[]>([]);
const categories = ref<Category[]>([]);
const loading = ref(false);
const loadError = ref("");

const searchQuery = ref("");
const selectedCategory = ref("");
const lowStockOnly = ref(false);

const adjustModalOpen = ref(false);
const adjustTarget = ref<StockRow | null>(null);

const LOW_STOCK_THRESHOLD = 5;

// ─── Derived ─────────────────────────────────────────────────────────────────

const productMap = computed<Record<string, Product>>(() => {
  const map: Record<string, Product> = {};
  for (const p of products.value) map[p.id] = p;
  return map;
});

const rows = computed<StockRow[]>(() =>
  stockItems.value.map((st) => ({
    ...st,
    product: productMap.value[st.product_id] ?? null,
  })),
);

const filteredRows = computed(() => {
  return rows.value.filter((row) => {
    const p = row.product;
    if (!p) return false;

    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase();
      const match =
        p.name.toLowerCase().includes(q) ||
        (p.sku ?? "").toLowerCase().includes(q) ||
        (p.barcode ?? "").toLowerCase().includes(q);
      if (!match) return false;
    }

    if (selectedCategory.value && p.category_id !== selectedCategory.value)
      return false;
    if (lowStockOnly.value && row.quantity_on_hand >= LOW_STOCK_THRESHOLD)
      return false;

    return true;
  });
});

const stats = computed(() => {
  const total = rows.value.length;
  const low = rows.value.filter(
    (r) => r.quantity_on_hand < LOW_STOCK_THRESHOLD,
  ).length;
  const outOfStock = rows.value.filter((r) => r.quantity_on_hand === 0).length;
  const totalValue = rows.value.reduce((acc, r) => {
    const price = r.product?.is_fiscal
      ? (r.product?.fiscal_price ?? 0)
      : (r.product?.internal_price ?? 0);
    return acc + price * r.quantity_on_hand;
  }, 0);
  return { total, low, outOfStock, totalValue };
});

// ─── Data loading ─────────────────────────────────────────────────────────────

async function fetchAll() {
  loading.value = true;
  loadError.value = "";
  try {
    const [stockData, productData, categoryData] = await Promise.all([
      useApiFetch<StockItem[]>("/api/v1/inventory/stock"),
      useApiFetch<Product[]>("/api/v1/products"),
      useApiFetch<Category[]>("/api/v1/products/categories"),
    ]);
    stockItems.value = stockData ?? [];
    products.value = productData ?? [];
    categories.value = categoryData ?? [];
  } catch {
    loadError.value = "Error al cargar el inventario";
  } finally {
    loading.value = false;
  }
}

// ─── Adjustment ───────────────────────────────────────────────────────────────

function openAdjust(row: StockRow) {
  adjustTarget.value = row;
  adjustModalOpen.value = true;
}

async function onAdjusted() {
  await fetchAll();
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function stockBadgeClass(qty: number) {
  if (qty === 0) return "bg-red-100 text-red-700";
  if (qty < LOW_STOCK_THRESHOLD) return "bg-orange-100 text-orange-700";
  return "bg-green-100 text-green-700";
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

onMounted(fetchAll);
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="mb-8 flex items-center justify-between gap-4">
      <div class="flex items-center gap-3">
        <div
          class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center"
        >
          <Warehouse class="w-5 h-5 text-primary" />
        </div>
        <div>
          <h1 class="text-2xl font-bold font-heading text-foreground">
            Inventario
          </h1>
          <p class="text-muted-foreground text-sm">
            Stock actual y movimientos
          </p>
        </div>
      </div>
      <NuxtLink
        to="/inventory/movements"
        class="flex items-center gap-2 h-10 px-4 border border-input text-sm font-semibold text-muted-foreground rounded-lg hover:bg-gray-50 transition-all duration-200 cursor-pointer"
      >
        <History class="w-4 h-4" />
        Historial
      </NuxtLink>
    </div>

    <!-- Stats cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <div class="rounded-xl border bg-card border-border shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Total productos
        </p>
        <p class="text-2xl font-bold font-heading text-foreground">
          {{ stats.total }}
        </p>
      </div>
      <div class="rounded-xl border bg-card border-border shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Stock bajo
        </p>
        <p
          class="text-2xl font-bold font-heading"
          :class="stats.low > 0 ? 'text-orange-600' : 'text-foreground'"
        >
          {{ stats.low }}
        </p>
      </div>
      <div class="rounded-xl border bg-card border-border shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Sin stock
        </p>
        <p
          class="text-2xl font-bold font-heading"
          :class="stats.outOfStock > 0 ? 'text-red-600' : 'text-foreground'"
        >
          {{ stats.outOfStock }}
        </p>
      </div>
      <div class="rounded-xl border bg-card border-border shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Valor del inventario
        </p>
        <p class="text-2xl font-bold font-heading text-foreground">
          ${{ stats.totalValue.toFixed(2) }}
        </p>
      </div>
    </div>

    <!-- Filters -->
    <div
      class="rounded-xl border bg-card border-border shadow-sm px-6 py-4 mb-4 flex flex-col sm:flex-row gap-3"
    >
      <div class="flex-1">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Buscar por nombre, SKU o código..."
          class="w-full h-10 px-4 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary placeholder:text-muted-foreground transition-all duration-200"
        />
      </div>
      <select
        v-model="selectedCategory"
        class="h-10 px-3 border border-input rounded-lg text-sm text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all duration-200 bg-white"
      >
        <option value="">Todas las categorías</option>
        <option v-for="cat in categories" :key="cat.id" :value="cat.id">
          {{ cat.name }}
        </option>
      </select>
      <button
        type="button"
        :class="[
          'h-10 px-4 flex items-center gap-2 text-sm font-semibold rounded-lg border transition-all duration-200 cursor-pointer',
          lowStockOnly
            ? 'bg-orange-50 border-orange-300 text-orange-700'
            : 'border-input text-muted-foreground hover:bg-gray-50',
        ]"
        @click="lowStockOnly = !lowStockOnly"
      >
        <TrendingDown class="w-4 h-4" />
        Stock bajo
      </button>
    </div>

    <!-- Table card -->
    <div class="rounded-xl border bg-card border-border shadow-sm">
      <div
        class="px-6 py-4 border-b border-border flex items-center justify-between"
      >
        <h2 class="text-base font-bold font-heading text-foreground">
          {{ filteredRows.length }} producto{{
            filteredRows.length !== 1 ? "s" : ""
          }}
        </h2>
        <div class="flex items-center gap-2 text-xs text-muted-foreground">
          <SlidersHorizontal class="w-3.5 h-3.5" />
          <span>Stock bajo &lt; {{ LOW_STOCK_THRESHOLD }} unidades</span>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="px-6 py-5 space-y-3">
        <div
          v-for="n in 6"
          :key="n"
          class="h-12 animate-pulse bg-gray-100 rounded-lg"
        />
      </div>

      <!-- Error -->
      <div v-else-if="loadError" class="px-6 py-12 text-center">
        <AlertTriangle class="w-10 h-10 text-red-400 mx-auto mb-3" />
        <p class="text-red-600 font-medium">{{ loadError }}</p>
        <button
          type="button"
          class="mt-4 h-9 px-4 border border-input text-sm font-semibold text-muted-foreground rounded-lg hover:bg-gray-50 cursor-pointer"
          @click="fetchAll"
        >
          Reintentar
        </button>
      </div>

      <!-- Empty -->
      <div v-else-if="filteredRows.length === 0" class="px-6 py-16 text-center">
        <FolderOpen class="w-12 h-12 text-gray-200 mx-auto mb-4" />
        <h3 class="text-base font-semibold text-muted-foreground mb-1">
          Sin resultados
        </h3>
        <p class="text-sm text-muted-foreground">
          {{
            rows.length === 0
              ? "No hay productos en el inventario aún"
              : "Ningún producto coincide con los filtros aplicados"
          }}
        </p>
      </div>

      <!-- ── Mobile card list (xs / sm) ──────────────────────── -->
      <div v-else class="sm:hidden divide-y divide-border">
        <div
          v-for="row in filteredRows"
          :key="row.product_id"
          class="flex items-center gap-3 px-4 py-3"
        >
          <!-- Stock badge -->
          <span
            class="no-min-tap shrink-0 inline-flex items-center justify-center w-12 h-12 rounded-xl text-sm font-bold"
            :class="stockBadgeClass(row.quantity_on_hand)"
          >
            {{ row.quantity_on_hand.toFixed(0) }}
          </span>

          <!-- Product info -->
          <div class="flex-1 min-w-0">
            <p class="font-semibold text-sm text-foreground truncate">
              {{ row.product?.name ?? "—" }}
            </p>
            <p
              v-if="row.product?.sku"
              class="text-[11px] text-muted-foreground font-mono"
            >
              {{ row.product.sku }}
            </p>
            <p class="text-[11px] text-muted-foreground">
              ${{
                row.product?.is_fiscal
                  ? (row.product?.fiscal_price ?? 0).toFixed(2)
                  : (row.product?.internal_price ?? 0).toFixed(2)
              }}
            </p>
          </div>

          <!-- Adjust button -->
          <button
            v-if="store.isOwner"
            type="button"
            class="no-min-tap shrink-0 h-10 px-3 text-xs font-semibold border border-input rounded-lg text-muted-foreground hover:bg-gray-100 transition-all duration-200 cursor-pointer"
            @click="openAdjust(row)"
          >
            Ajustar
          </button>
        </div>
      </div>

      <!-- ── Desktop table (sm+) ────────────────────────────── -->
      <div
        v-if="filteredRows.length > 0"
        class="hidden sm:block overflow-x-auto"
      >
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border">
              <th
                class="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wide"
              >
                Producto
              </th>
              <th
                class="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wide hidden md:table-cell"
              >
                Categoría
              </th>
              <th
                class="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wide hidden lg:table-cell"
              >
                Precio
              </th>
              <th
                class="px-4 py-3 text-center text-xs font-semibold text-muted-foreground uppercase tracking-wide"
              >
                <span class="flex items-center justify-center gap-1">
                  <ArrowUpDown class="w-3 h-3" />
                  Stock
                </span>
              </th>
              <th
                class="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wide hidden lg:table-cell"
              >
                Actualizado
              </th>
              <th
                v-if="store.isOwner"
                class="px-4 py-3 text-right text-xs font-semibold text-muted-foreground uppercase tracking-wide"
              >
                Acciones
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr
              v-for="row in filteredRows"
              :key="row.product_id"
              class="hover:bg-gray-50/50 transition-colors duration-150"
            >
              <!-- Product -->
              <td class="px-6 py-4">
                <div>
                  <p class="font-semibold text-foreground">
                    {{ row.product?.name ?? "—" }}
                  </p>
                  <p
                    v-if="row.product?.sku"
                    class="text-xs text-muted-foreground font-mono mt-0.5"
                  >
                    {{ row.product.sku }}
                  </p>
                </div>
              </td>

              <!-- Category -->
              <td class="px-4 py-4 hidden md:table-cell">
                <span class="text-muted-foreground">
                  {{
                    categories.find((c) => c.id === row.product?.category_id)
                      ?.name ?? "—"
                  }}
                </span>
              </td>

              <!-- Price -->
              <td class="px-4 py-4 hidden lg:table-cell">
                <span class="font-medium text-foreground">
                  ${{
                    row.product?.is_fiscal
                      ? (row.product?.fiscal_price ?? 0).toFixed(2)
                      : (row.product?.internal_price ?? 0).toFixed(2)
                  }}
                </span>
                <span
                  v-if="row.product?.is_fiscal"
                  class="ml-1 text-xs text-muted-foreground"
                >
                  +{{ row.product?.tax_rate ?? 0 }}% IVA
                </span>
              </td>

              <!-- Stock badge -->
              <td class="px-4 py-4 text-center">
                <span
                  class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold"
                  :class="stockBadgeClass(row.quantity_on_hand)"
                >
                  {{ row.quantity_on_hand.toFixed(0) }}
                </span>
              </td>

              <!-- Updated at -->
              <td
                class="px-4 py-4 text-xs text-muted-foreground hidden lg:table-cell"
              >
                {{ formatDate(row.last_updated_at) }}
              </td>

              <!-- Actions -->
              <td v-if="store.isOwner" class="px-4 py-4 text-right">
                <button
                  type="button"
                  class="h-8 px-3 text-xs font-semibold border border-input rounded-lg text-muted-foreground hover:bg-gray-100 transition-all duration-200 cursor-pointer"
                  @click="openAdjust(row)"
                >
                  Ajustar
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Adjustment modal -->
    <AdjustmentModal
      v-if="adjustTarget"
      v-model:open="adjustModalOpen"
      :product-id="adjustTarget.product_id"
      :product-name="adjustTarget.product?.name ?? ''"
      :current-stock="adjustTarget.quantity_on_hand"
      @adjusted="onAdjusted"
    />
  </div>
</template>
