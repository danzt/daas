<script setup lang="ts">
import {
  Receipt,
  Plus,
  Search,
  Eye,
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle,
  RefreshCw,
  Loader2,
  FolderOpen,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import type { Product } from "~/components/products/ProductFormModal.vue";
import type { FiscalInvoice } from "~/components/invoices/FiscalInvoiceFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

// ─── Data ────────────────────────────────────────────────────────────────────
const invoices = ref<FiscalInvoice[]>([]);
const products = ref<Product[]>([]);
const loading = ref(false);
const loadError = ref("");

// ─── Filters ─────────────────────────────────────────────────────────────────
const searchQuery = ref("");
const statusFilter = ref("all");

// ─── Modal ───────────────────────────────────────────────────────────────────
const formModalOpen = ref(false);

// ─── Stats ───────────────────────────────────────────────────────────────────
const issuedCount = computed(
  () => invoices.value.filter((i) => i.status === "issued").length,
);
const pendingCount = computed(
  () =>
    invoices.value.filter(
      (i) => i.status === "draft" || i.status === "pending_fiscal",
    ).length,
);
const failedCount = computed(
  () => invoices.value.filter((i) => i.status === "failed").length,
);
const cancelledCount = computed(
  () => invoices.value.filter((i) => i.status === "cancelled").length,
);

// ─── Filtered ────────────────────────────────────────────────────────────────
const filteredInvoices = computed(() => {
  let list = invoices.value;
  if (statusFilter.value !== "all") {
    list = list.filter((i) => i.status === statusFilter.value);
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter(
      (i) =>
        i.fiscal_number?.toLowerCase().includes(q) ||
        i.customer_name?.toLowerCase().includes(q),
    );
  }
  return list;
});

const filterTabs = [
  { key: "all", label: "Todas" },
  { key: "draft", label: "Borradores" },
  { key: "issued", label: "Emitidas" },
  { key: "failed", label: "Fallidas" },
  { key: "cancelled", label: "Canceladas" },
];

// ─── API ─────────────────────────────────────────────────────────────────────
async function loadInvoices() {
  loading.value = true;
  loadError.value = "";
  try {
    invoices.value = await useApiFetch<FiscalInvoice[]>(
      "/api/v1/invoices/fiscal",
    );
  } catch {
    loadError.value = "No se pudieron cargar las facturas fiscales";
  } finally {
    loading.value = false;
  }
}

async function loadProducts() {
  try {
    products.value = await useApiFetch<Product[]>("/api/v1/products");
  } catch {
    // non-critical
  }
}

onMounted(() => {
  loadInvoices();
  loadProducts();
});

// ─── Helpers ─────────────────────────────────────────────────────────────────
const STATUS_CONFIG = {
  draft: {
    label: "Borrador",
    icon: Clock,
    class: "bg-muted text-muted-foreground",
  },
  pending_fiscal: {
    label: "Enviando...",
    icon: RefreshCw,
    class: "bg-blue-50 text-blue-600",
  },
  issued: {
    label: "Emitida",
    icon: CheckCircle,
    class: "bg-green-50 text-green-700",
  },
  failed: {
    label: "Fallida",
    icon: AlertCircle,
    class: "bg-orange-50 text-orange-600",
  },
  cancelled: {
    label: "Cancelada",
    icon: XCircle,
    class: "bg-red-50 text-red-600",
  },
} as const;

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

function formatCurrency(value: number) {
  return value.toFixed(2);
}
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold font-heading text-text-brand">
          Facturas Fiscales
        </h1>
        <p class="text-sm text-muted-foreground mt-0.5">
          Documentos SENIAT — productos con IVA
        </p>
      </div>
      <button
        type="button"
        class="flex items-center gap-2 h-10 px-5 bg-cta text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer"
        @click="formModalOpen = true"
      >
        <Plus class="w-4 h-4" />
        Nueva factura fiscal
      </button>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
      <div class="bg-card rounded-xl border shadow-sm px-5 py-4">
        <div class="flex items-center gap-3">
          <div
            class="w-9 h-9 rounded-lg bg-green-50 flex items-center justify-center flex-shrink-0"
          >
            <CheckCircle class="w-4 h-4 text-green-600" />
          </div>
          <div>
            <p
              class="text-2xl font-bold font-heading text-text-brand tabular-nums"
            >
              {{ issuedCount }}
            </p>
            <p class="text-xs text-muted-foreground">Emitidas</p>
          </div>
        </div>
      </div>

      <div class="bg-card rounded-xl border shadow-sm px-5 py-4">
        <div class="flex items-center gap-3">
          <div
            class="w-9 h-9 rounded-lg bg-muted flex items-center justify-center flex-shrink-0"
          >
            <Clock class="w-4 h-4 text-muted-foreground" />
          </div>
          <div>
            <p
              class="text-2xl font-bold font-heading text-text-brand tabular-nums"
            >
              {{ pendingCount }}
            </p>
            <p class="text-xs text-muted-foreground">Pendientes</p>
          </div>
        </div>
      </div>

      <div class="bg-card rounded-xl border shadow-sm px-5 py-4">
        <div class="flex items-center gap-3">
          <div
            class="w-9 h-9 rounded-lg bg-orange-50 flex items-center justify-center flex-shrink-0"
          >
            <AlertCircle class="w-4 h-4 text-orange-500" />
          </div>
          <div>
            <p
              class="text-2xl font-bold font-heading text-text-brand tabular-nums"
            >
              {{ failedCount }}
            </p>
            <p class="text-xs text-muted-foreground">Fallidas</p>
          </div>
        </div>
      </div>

      <div class="bg-card rounded-xl border shadow-sm px-5 py-4">
        <div class="flex items-center gap-3">
          <div
            class="w-9 h-9 rounded-lg bg-red-50 flex items-center justify-center flex-shrink-0"
          >
            <XCircle class="w-4 h-4 text-red-500" />
          </div>
          <div>
            <p
              class="text-2xl font-bold font-heading text-text-brand tabular-nums"
            >
              {{ cancelledCount }}
            </p>
            <p class="text-xs text-muted-foreground">Canceladas</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Filters + Table -->
    <div class="bg-card rounded-xl border shadow-sm">
      <div class="px-5 py-4 border-b flex flex-col sm:flex-row gap-3">
        <div class="flex items-center gap-1 bg-muted rounded-lg p-1">
          <button
            v-for="tab in filterTabs"
            :key="tab.key"
            type="button"
            :class="[
              'px-3 py-1.5 text-xs font-semibold rounded-md transition-all duration-200 cursor-pointer',
              statusFilter === tab.key
                ? 'bg-card text-text-brand shadow-sm'
                : 'text-muted-foreground hover:text-foreground',
            ]"
            @click="statusFilter = tab.key"
          >
            {{ tab.label }}
            <span
              v-if="tab.key === 'failed' && failedCount > 0"
              class="ml-1 px-1.5 py-0.5 bg-orange-500 text-white rounded-full text-[10px] font-bold"
            >
              {{ failedCount }}
            </span>
          </button>
        </div>

        <div class="relative flex-1 max-w-xs ml-auto">
          <Search
            class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none"
          />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Buscar por número fiscal o cliente..."
            class="w-full h-9 pl-9 pr-4 border rounded-lg text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-gray-400"
          />
        </div>
      </div>

      <div class="overflow-x-auto">
        <div v-if="loading" class="flex items-center justify-center py-16">
          <Loader2 class="w-6 h-6 text-primary animate-spin" />
        </div>

        <div
          v-else-if="loadError"
          class="flex flex-col items-center justify-center py-16 gap-3"
        >
          <p class="text-sm text-red-600">{{ loadError }}</p>
          <button
            type="button"
            class="text-xs text-primary font-semibold hover:underline cursor-pointer"
            @click="loadInvoices"
          >
            Reintentar
          </button>
        </div>

        <div
          v-else-if="filteredInvoices.length === 0"
          class="flex flex-col items-center justify-center py-16 gap-3"
        >
          <div
            class="w-14 h-14 rounded-full bg-muted flex items-center justify-center"
          >
            <FolderOpen class="w-7 h-7 text-muted-foreground/40" />
          </div>
          <div class="text-center">
            <p class="text-sm font-semibold text-text-brand">
              {{
                invoices.length === 0
                  ? "Aún no hay facturas fiscales"
                  : "Sin resultados"
              }}
            </p>
            <p class="text-xs text-muted-foreground mt-1">
              {{
                invoices.length === 0
                  ? "Creá tu primera factura SENIAT"
                  : "Probá cambiando los filtros"
              }}
            </p>
          </div>
          <button
            v-if="invoices.length === 0"
            type="button"
            class="flex items-center gap-2 h-9 px-4 bg-cta text-white text-xs font-semibold rounded-lg hover:opacity-90 transition-all cursor-pointer"
            @click="formModalOpen = true"
          >
            <Plus class="w-3.5 h-3.5" />
            Nueva factura fiscal
          </button>
        </div>

        <table v-else class="w-full">
          <thead>
            <tr class="border-b bg-muted/40">
              <th
                class="px-5 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider"
              >
                N° Fiscal
              </th>
              <th
                class="px-5 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider"
              >
                Cliente
              </th>
              <th
                class="px-4 py-3 text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider"
              >
                Base
              </th>
              <th
                class="px-4 py-3 text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider"
              >
                IVA
              </th>
              <th
                class="px-4 py-3 text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider"
              >
                Total
              </th>
              <th
                class="px-5 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider"
              >
                Estado
              </th>
              <th
                class="px-5 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider"
              >
                Fecha
              </th>
              <th
                class="px-5 py-3 text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider"
              >
                Acciones
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            <tr
              v-for="inv in filteredInvoices"
              :key="inv.id"
              class="hover:bg-muted/30 transition-colors duration-150"
            >
              <td class="px-5 py-4">
                <span class="text-sm font-semibold font-mono text-text-brand">
                  {{ inv.fiscal_number || "—" }}
                </span>
              </td>
              <td class="px-5 py-4">
                <p class="text-sm text-text-brand">
                  {{ inv.customer_name || "Consumidor Final" }}
                </p>
                <p
                  v-if="inv.customer_id_number"
                  class="text-xs text-muted-foreground"
                >
                  {{ inv.customer_id_type.toUpperCase() }}:
                  {{ inv.customer_id_number }}
                </p>
              </td>
              <td class="px-4 py-4 text-right">
                <span class="text-sm tabular-nums text-muted-foreground">
                  {{ formatCurrency(inv.subtotal_base) }}
                </span>
              </td>
              <td class="px-4 py-4 text-right">
                <span class="text-sm tabular-nums text-muted-foreground">
                  {{ formatCurrency(inv.tax_amount) }}
                </span>
              </td>
              <td class="px-4 py-4 text-right">
                <span
                  class="text-sm font-semibold tabular-nums text-text-brand"
                >
                  {{ formatCurrency(inv.total) }}
                </span>
              </td>
              <td class="px-5 py-4">
                <span
                  :class="[
                    'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold',
                    STATUS_CONFIG[inv.status].class,
                  ]"
                >
                  <component
                    :is="STATUS_CONFIG[inv.status].icon"
                    class="w-3 h-3"
                  />
                  {{ STATUS_CONFIG[inv.status].label }}
                </span>
              </td>
              <td class="px-5 py-4">
                <span class="text-sm text-muted-foreground">
                  {{ formatDate(inv.created_at) }}
                </span>
              </td>
              <td class="px-5 py-4 text-right">
                <NuxtLink
                  :to="`/invoices/fiscal/${inv.id}`"
                  class="inline-flex items-center gap-1.5 h-8 px-3 border text-xs font-semibold text-muted-foreground rounded-lg hover:bg-muted hover:text-text-brand transition-all duration-200"
                >
                  <Eye class="w-3.5 h-3.5" />
                  Ver
                </NuxtLink>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <FiscalInvoiceFormModal
      v-model="formModalOpen"
      :products="products"
      @saved="loadInvoices"
    />
  </div>
</template>
