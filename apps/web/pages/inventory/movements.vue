<script setup lang="ts">
import {
  History,
  TrendingUp,
  TrendingDown,
  SlidersHorizontal,
  AlertTriangle,
  FolderOpen,
  ArrowLeft,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import type { Product } from "~/components/products/ProductFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

// ─── Types ────────────────────────────────────────────────────────────────────

interface Movement {
  id: string;
  tenant_id: string;
  product_id: string;
  type: "entry" | "exit" | "adjustment";
  quantity: number;
  unit_cost?: number;
  reference_type: string;
  reference_id?: string;
  notes: string;
  created_by: string;
  created_at: string;
}

// ─── State ────────────────────────────────────────────────────────────────────

const movements = ref<Movement[]>([]);
const products = ref<Product[]>([]);
const loading = ref(false);
const loadError = ref("");

const filterProduct = ref("");
const filterType = ref("");
const filterFrom = ref("");
const filterTo = ref("");

// ─── Derived ─────────────────────────────────────────────────────────────────

const productMap = computed<Record<string, Product>>(() => {
  const map: Record<string, Product> = {};
  for (const p of products.value) map[p.id] = p;
  return map;
});

// ─── Data ─────────────────────────────────────────────────────────────────────

async function fetchMovements() {
  loading.value = true;
  loadError.value = "";
  try {
    const params = new URLSearchParams();
    if (filterProduct.value) params.set("product_id", filterProduct.value);
    if (filterType.value) params.set("type", filterType.value);
    if (filterFrom.value)
      params.set("from", new Date(filterFrom.value).toISOString());
    if (filterTo.value)
      params.set("to", new Date(filterTo.value + "T23:59:59").toISOString());

    const url = `/api/v1/inventory/movements${params.toString() ? "?" + params.toString() : ""}`;
    const data = await useApiFetch<Movement[]>(url);
    movements.value = data ?? [];
  } catch {
    loadError.value = "Error al cargar los movimientos";
  } finally {
    loading.value = false;
  }
}

async function fetchProducts() {
  try {
    const data = await useApiFetch<Product[]>("/api/v1/products");
    products.value = data ?? [];
  } catch {
    // non-fatal — product names just show as IDs
  }
}

function applyFilters() {
  fetchMovements();
}

function clearFilters() {
  filterProduct.value = "";
  filterType.value = "";
  filterFrom.value = "";
  filterTo.value = "";
  fetchMovements();
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function typeLabel(type: Movement["type"]) {
  const map: Record<Movement["type"], string> = {
    entry: "Entrada",
    exit: "Salida",
    adjustment: "Ajuste",
  };
  return map[type] ?? type;
}

function typeBadgeClass(type: Movement["type"]) {
  if (type === "entry") return "bg-green-100 text-green-700";
  if (type === "exit") return "bg-red-100 text-red-700";
  return "bg-blue-100 text-blue-700";
}

function typeIcon(type: Movement["type"]) {
  if (type === "entry") return TrendingUp;
  if (type === "exit") return TrendingDown;
  return SlidersHorizontal;
}

function refTypeLabel(ref: string) {
  const map: Record<string, string> = {
    purchase_order: "Orden de compra",
    sale: "Venta",
    manual_adjustment: "Ajuste manual",
    opening: "Stock inicial",
  };
  return map[ref] ?? ref;
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString("es-VE", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

onMounted(() => {
  fetchProducts();
  fetchMovements();
});
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="mb-8 flex items-center gap-4">
      <NuxtLink
        to="/inventory"
        class="w-9 h-9 rounded-lg border border-gray-200 flex items-center justify-center text-gray-500 hover:bg-gray-50 transition-all duration-200 cursor-pointer"
      >
        <ArrowLeft class="w-4 h-4" />
      </NuxtLink>
      <div class="flex items-center gap-3">
        <div
          class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center"
        >
          <History class="w-5 h-5 text-primary" />
        </div>
        <div>
          <h1 class="text-2xl font-bold font-heading text-foreground">
            Historial de movimientos
          </h1>
          <p class="text-gray-500 text-sm">Entradas, salidas y ajustes</p>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div
      class="bg-white rounded-xl border border-gray-100 shadow-sm px-6 py-4 mb-4"
    >
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 mb-3">
        <!-- Product filter -->
        <select
          v-model="filterProduct"
          class="h-10 px-3 border border-gray-200 rounded-lg text-sm text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all duration-200 bg-white"
        >
          <option value="">Todos los productos</option>
          <option v-for="p in products" :key="p.id" :value="p.id">
            {{ p.name }}
          </option>
        </select>

        <!-- Type filter -->
        <select
          v-model="filterType"
          class="h-10 px-3 border border-gray-200 rounded-lg text-sm text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all duration-200 bg-white"
        >
          <option value="">Todos los tipos</option>
          <option value="entry">Entradas</option>
          <option value="exit">Salidas</option>
          <option value="adjustment">Ajustes</option>
        </select>

        <!-- From date -->
        <input
          v-model="filterFrom"
          type="date"
          class="h-10 px-3 border border-gray-200 rounded-lg text-sm text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all duration-200"
        />

        <!-- To date -->
        <input
          v-model="filterTo"
          type="date"
          class="h-10 px-3 border border-gray-200 rounded-lg text-sm text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all duration-200"
        />
      </div>

      <div class="flex items-center justify-end gap-2">
        <button
          type="button"
          class="h-9 px-4 border border-gray-200 text-sm font-semibold text-gray-600 rounded-lg hover:bg-gray-50 transition-all duration-200 cursor-pointer"
          @click="clearFilters"
        >
          Limpiar
        </button>
        <button
          type="button"
          class="h-9 px-4 bg-primary text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer"
          @click="applyFilters"
        >
          Filtrar
        </button>
      </div>
    </div>

    <!-- Content card -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm">
      <div
        class="px-6 py-4 border-b border-gray-100 flex items-center justify-between"
      >
        <h2 class="text-base font-bold font-heading text-foreground">
          {{ movements.length }} movimiento{{
            movements.length !== 1 ? "s" : ""
          }}
        </h2>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="px-6 py-5 space-y-3">
        <div
          v-for="n in 8"
          :key="n"
          class="h-16 animate-pulse bg-gray-100 rounded-lg"
        />
      </div>

      <!-- Error -->
      <div v-else-if="loadError" class="px-6 py-12 text-center">
        <AlertTriangle class="w-10 h-10 text-red-400 mx-auto mb-3" />
        <p class="text-red-600 font-medium">{{ loadError }}</p>
        <button
          type="button"
          class="mt-4 h-9 px-4 border border-gray-200 text-sm font-semibold text-gray-600 rounded-lg hover:bg-gray-50 cursor-pointer"
          @click="fetchMovements"
        >
          Reintentar
        </button>
      </div>

      <!-- Empty -->
      <div v-else-if="movements.length === 0" class="px-6 py-16 text-center">
        <FolderOpen class="w-12 h-12 text-gray-200 mx-auto mb-4" />
        <h3 class="text-base font-semibold text-gray-600 mb-1">
          Sin movimientos
        </h3>
        <p class="text-sm text-gray-400">
          No hay movimientos registrados con los filtros aplicados
        </p>
      </div>

      <!-- Timeline list -->
      <div v-else class="divide-y divide-gray-50">
        <div
          v-for="mov in movements"
          :key="mov.id"
          class="flex items-start gap-4 px-6 py-4 hover:bg-gray-50/50 transition-colors duration-150"
        >
          <!-- Type icon -->
          <div
            class="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0 mt-0.5"
            :class="
              mov.type === 'entry'
                ? 'bg-green-100'
                : mov.type === 'exit'
                  ? 'bg-red-100'
                  : 'bg-blue-100'
            "
          >
            <component
              :is="typeIcon(mov.type)"
              class="w-4 h-4"
              :class="
                mov.type === 'entry'
                  ? 'text-green-600'
                  : mov.type === 'exit'
                    ? 'text-red-600'
                    : 'text-blue-600'
              "
            />
          </div>

          <!-- Content -->
          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between gap-2">
              <div>
                <p class="text-sm font-semibold text-foreground">
                  {{
                    productMap[mov.product_id]?.name ??
                    mov.product_id.slice(0, 8) + "..."
                  }}
                </p>
                <p class="text-xs text-gray-500 mt-0.5">
                  {{ refTypeLabel(mov.reference_type) }}
                  <span v-if="mov.notes" class="text-gray-400">
                    · {{ mov.notes }}</span
                  >
                </p>
              </div>
              <div class="text-right flex-shrink-0">
                <span
                  class="inline-flex px-2 py-0.5 rounded-full text-xs font-bold"
                  :class="typeBadgeClass(mov.type)"
                >
                  {{ typeLabel(mov.type) }}
                  {{ mov.type === "exit" ? "-" : "+" }}{{ mov.quantity }}
                </span>
                <p class="text-xs text-gray-400 mt-1">
                  {{ formatDate(mov.created_at) }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
