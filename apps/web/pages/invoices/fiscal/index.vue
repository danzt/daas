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
  RotateCcw,
  AlertTriangle,
  Wifi,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import type { Product } from "~/components/products/ProductFormModal.vue";
import FiscalInvoiceFormModal from "~/components/invoices/FiscalInvoiceFormModal.vue";
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

// ─── Retry state ─────────────────────────────────────────────────────────────
const retrying = ref<Set<string>>(new Set());

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

// Invoices that need attention (pending_fiscal or failed)
const attentionCount = computed(
  () =>
    invoices.value.filter(
      (i) => i.status === "failed" || i.status === "pending_fiscal",
    ).length,
);
const hasPendingFiscal = computed(() =>
  invoices.value.some((i) => i.status === "pending_fiscal"),
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
  { key: "pending_fiscal", label: "Enviando" },
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

async function handleRetry(inv: FiscalInvoice) {
  if (retrying.value.has(inv.id)) return;
  retrying.value = new Set([...retrying.value, inv.id]);
  try {
    const updated = await useApiFetch<FiscalInvoice>(
      `/api/v1/invoices/fiscal/${inv.id}/retry`,
      { method: "POST" },
    );
    const idx = invoices.value.findIndex((i) => i.id === inv.id);
    if (idx !== -1) invoices.value[idx] = updated;
  } catch {
    // error shown in row — reload to get fresh state
    await loadInvoices();
  } finally {
    const next = new Set(retrying.value);
    next.delete(inv.id);
    retrying.value = next;
  }
}

// ─── Auto-refresh while pending_fiscal invoices exist ────────────────────────
let refreshTimer: ReturnType<typeof setInterval> | null = null;

function startAutoRefresh() {
  if (refreshTimer) return;
  refreshTimer = setInterval(async () => {
    if (!hasPendingFiscal.value) {
      stopAutoRefresh();
      return;
    }
    try {
      invoices.value = await useApiFetch<FiscalInvoice[]>(
        "/api/v1/invoices/fiscal",
      );
    } catch {
      // silent
    }
  }, 30_000);
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
}

watch(hasPendingFiscal, (has) => {
  if (has) startAutoRefresh();
  else stopAutoRefresh();
});

onMounted(() => {
  loadInvoices();
  loadProducts();
});

onUnmounted(() => stopAutoRefresh());

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
  <div class="p-4 sm:p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold font-heading text-foreground">
          Facturas Fiscales
        </h1>
        <p class="text-sm text-muted-foreground mt-0.5">
          Documentos SENIAT — productos con IVA
        </p>
      </div>
      <button
        type="button"
        class="flex items-center gap-2 h-10 px-5 bg-primary text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer"
        @click="formModalOpen = true"
      >
        <Plus class="w-4 h-4" />
        Nueva factura fiscal
      </button>
    </div>

    <!-- ─── Attention banner ─────────────────────────────────────────────────── -->
    <Transition
      enter-active-class="transition-all duration-300"
      enter-from-class="opacity-0 -translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition-all duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="attentionCount > 0 && !loading"
        :class="[
          'rounded-xl border px-5 py-4 flex items-start gap-4',
          failedCount > 0
            ? 'bg-orange-50 border-orange-200'
            : 'bg-blue-50 border-blue-200',
        ]"
      >
        <div
          :class="[
            'w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0 mt-0.5',
            failedCount > 0 ? 'bg-orange-100' : 'bg-blue-100',
          ]"
        >
          <component
            :is="failedCount > 0 ? AlertTriangle : Wifi"
            :class="[
              'w-4 h-4',
              failedCount > 0 ? 'text-orange-600' : 'text-blue-600',
            ]"
          />
        </div>
        <div class="flex-1">
          <p
            :class="[
              'text-sm font-semibold',
              failedCount > 0 ? 'text-orange-800' : 'text-blue-800',
            ]"
          >
            <template v-if="failedCount > 0">
              {{ failedCount }} factura{{
                failedCount !== 1 ? "s" : ""
              }}
              fallida{{ failedCount !== 1 ? "s" : "" }} — requieren atención
            </template>
            <template v-else>
              {{ hasPendingFiscal ? "Enviando al SENIAT..." : "" }}
            </template>
          </p>
          <p
            :class="[
              'text-xs mt-1',
              failedCount > 0 ? 'text-orange-600' : 'text-blue-600',
            ]"
          >
            <template v-if="failedCount > 0">
              El worker reintenta automáticamente con backoff exponencial. Podés
              forzar un reintento manual desde la tabla.
            </template>
            <template v-else>
              Actualizando estado automáticamente cada 30 segundos.
            </template>
          </p>
        </div>
        <button
          v-if="failedCount > 0"
          type="button"
          class="text-xs font-semibold text-orange-700 hover:underline cursor-pointer shrink-0"
          @click="statusFilter = 'failed'"
        >
          Ver fallidas
        </button>
      </div>
    </Transition>

    <!-- Stats -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
      <div class="rounded-xl border bg-card shadow-sm px-5 py-4">
        <div class="flex items-center gap-3">
          <div
            class="w-9 h-9 rounded-lg bg-green-50 flex items-center justify-center flex-shrink-0"
          >
            <CheckCircle class="w-4 h-4 text-green-600" />
          </div>
          <div>
            <p
              class="text-2xl font-bold font-heading text-foreground tabular-nums"
            >
              {{ issuedCount }}
            </p>
            <p class="text-xs text-muted-foreground">Emitidas</p>
          </div>
        </div>
      </div>

      <div class="rounded-xl border bg-card shadow-sm px-5 py-4">
        <div class="flex items-center gap-3">
          <div
            class="w-9 h-9 rounded-lg bg-muted flex items-center justify-center flex-shrink-0"
          >
            <Clock class="w-4 h-4 text-muted-foreground" />
          </div>
          <div>
            <p
              class="text-2xl font-bold font-heading text-foreground tabular-nums"
            >
              {{ pendingCount }}
            </p>
            <p class="text-xs text-muted-foreground">Pendientes</p>
          </div>
        </div>
      </div>

      <div
        class="rounded-xl border shadow-sm px-5 py-4 transition-colors"
        :class="failedCount > 0 ? 'bg-orange-50 border-orange-200' : 'bg-card'"
      >
        <div class="flex items-center gap-3">
          <div
            class="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0"
            :class="failedCount > 0 ? 'bg-orange-100' : 'bg-orange-50'"
          >
            <AlertCircle
              class="w-4 h-4"
              :class="failedCount > 0 ? 'text-orange-600' : 'text-orange-400'"
            />
          </div>
          <div>
            <p
              class="text-2xl font-bold font-heading tabular-nums"
              :class="failedCount > 0 ? 'text-orange-700' : 'text-foreground'"
            >
              {{ failedCount }}
            </p>
            <p class="text-xs text-muted-foreground">Fallidas</p>
          </div>
        </div>
      </div>

      <div class="rounded-xl border bg-card shadow-sm px-5 py-4">
        <div class="flex items-center gap-3">
          <div
            class="w-9 h-9 rounded-lg bg-red-50 flex items-center justify-center flex-shrink-0"
          >
            <XCircle class="w-4 h-4 text-red-500" />
          </div>
          <div>
            <p
              class="text-2xl font-bold font-heading text-foreground tabular-nums"
            >
              {{ cancelledCount }}
            </p>
            <p class="text-xs text-muted-foreground">Canceladas</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Filters + Table -->
    <div class="rounded-xl border bg-card shadow-sm">
      <div class="px-5 py-4 border-b flex flex-col sm:flex-row gap-3">
        <div class="flex items-center gap-1 bg-muted rounded-lg p-1 flex-wrap">
          <button
            v-for="tab in filterTabs"
            :key="tab.key"
            type="button"
            :class="[
              'px-3 py-1.5 text-xs font-semibold rounded-md transition-all duration-200 cursor-pointer',
              statusFilter === tab.key
                ? 'bg-white text-foreground shadow-sm'
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
            <span
              v-if="tab.key === 'pending_fiscal' && hasPendingFiscal"
              class="ml-1 inline-block w-1.5 h-1.5 bg-blue-500 rounded-full animate-pulse"
            />
          </button>
        </div>

        <div class="flex items-center gap-2 ml-auto">
          <!-- Manual refresh -->
          <button
            type="button"
            :disabled="loading"
            class="h-9 w-9 flex items-center justify-center border rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-all disabled:opacity-40 cursor-pointer"
            title="Actualizar"
            @click="loadInvoices"
          >
            <RefreshCw class="w-3.5 h-3.5" :class="loading && 'animate-spin'" />
          </button>

          <div class="relative max-w-xs w-full">
            <Search
              class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none"
            />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Buscar por número fiscal o cliente..."
              class="w-full h-9 pl-9 pr-4 border rounded-lg text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-muted-foreground"
            />
          </div>
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
            <p class="text-sm font-semibold text-foreground">
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
            class="flex items-center gap-2 h-9 px-4 bg-primary text-white text-xs font-semibold rounded-lg hover:opacity-90 transition-all cursor-pointer"
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
              :class="[
                'transition-colors duration-150',
                inv.status === 'failed'
                  ? 'bg-orange-50/40 hover:bg-orange-50/70'
                  : inv.status === 'pending_fiscal'
                    ? 'bg-blue-50/30 hover:bg-blue-50/50'
                    : 'hover:bg-muted/30',
              ]"
            >
              <td class="px-5 py-4">
                <span class="text-sm font-semibold font-mono text-foreground">
                  {{ inv.fiscal_number || "—" }}
                </span>
              </td>
              <td class="px-5 py-4">
                <p class="text-sm text-foreground">
                  {{ inv.customer_name || "Consumidor Final" }}
                </p>
                <p
                  v-if="inv.customer_id_number"
                  class="text-xs text-muted-foreground"
                >
                  {{ inv.customer_id_type.toUpperCase() }}:
                  {{ inv.customer_id_number }}
                </p>
                <!-- fail_reason inline -->
                <p
                  v-if="inv.status === 'failed' && inv.fail_reason"
                  class="text-xs text-orange-600 mt-1 max-w-xs truncate"
                  :title="inv.fail_reason"
                >
                  ⚠ {{ inv.fail_reason }}
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
                  class="text-sm font-semibold tabular-nums text-foreground"
                >
                  {{ formatCurrency(inv.total) }}
                </span>
              </td>
              <td class="px-5 py-4">
                <div class="flex flex-col gap-1">
                  <span
                    :class="[
                      'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold w-fit',
                      STATUS_CONFIG[inv.status].class,
                    ]"
                  >
                    <component
                      :is="STATUS_CONFIG[inv.status].icon"
                      class="w-3 h-3"
                      :class="inv.status === 'pending_fiscal' && 'animate-spin'"
                    />
                    {{ STATUS_CONFIG[inv.status].label }}
                  </span>
                  <!-- retry count badge -->
                  <span
                    v-if="inv.retry_count > 0"
                    class="text-xs text-muted-foreground"
                  >
                    {{ inv.retry_count }} intento{{
                      inv.retry_count !== 1 ? "s" : ""
                    }}
                  </span>
                </div>
              </td>
              <td class="px-5 py-4">
                <span class="text-sm text-muted-foreground">
                  {{ formatDate(inv.created_at) }}
                </span>
              </td>
              <td class="px-5 py-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <!-- Retry button — only for failed invoices -->
                  <button
                    v-if="inv.status === 'failed'"
                    type="button"
                    :disabled="retrying.has(inv.id)"
                    class="inline-flex items-center gap-1.5 h-8 px-3 border border-orange-300 bg-orange-50 text-orange-700 text-xs font-semibold rounded-lg hover:bg-orange-100 transition-all duration-200 disabled:opacity-50 cursor-pointer"
                    @click="handleRetry(inv)"
                  >
                    <Loader2
                      v-if="retrying.has(inv.id)"
                      class="w-3 h-3 animate-spin"
                    />
                    <RotateCcw v-else class="w-3 h-3" />
                    {{ retrying.has(inv.id) ? "Enviando..." : "Reintentar" }}
                  </button>

                  <NuxtLink
                    :to="`/invoices/fiscal/${inv.id}`"
                    class="inline-flex items-center gap-1.5 h-8 px-3 border text-xs font-semibold text-muted-foreground rounded-lg hover:bg-muted hover:text-foreground transition-all duration-200"
                  >
                    <Eye class="w-3.5 h-3.5" />
                    Ver
                  </NuxtLink>
                </div>
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
