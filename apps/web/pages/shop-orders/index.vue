<script setup lang="ts">
import {
  ShoppingCart,
  Search,
  Loader2,
  FolderOpen,
  ChevronRight,
  Clock,
  CreditCard,
  Package,
  CheckCircle2,
  XCircle,
  Mail,
  Phone,
  MapPin,
  TrendingUp,
  Download,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

// ─── Types ────────────────────────────────────────────────────────────────────
interface OrderLine {
  id: string;
  product_id: string;
  name: string;
  unit_price: number;
  quantity: number;
  subtotal: number;
  is_fiscal: boolean;
}

export interface ShopOrder {
  id: string;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  shipping_address: string;
  shipping_city: string;
  subtotal: number;
  shipping_cost: number;
  total: number;
  status: "pending" | "paid" | "fulfilled" | "delivered" | "cancelled";
  notes: string;
  paid_at: string | null;
  fulfilled_at: string | null;
  delivered_at: string | null;
  cancelled_at: string | null;
  created_at: string;
  lines?: OrderLine[];
}

// ─── State ────────────────────────────────────────────────────────────────────
const { isMobile } = useMobileMode();

const orders = ref<ShopOrder[]>([]);
const loading = ref(false);
const loadError = ref("");
const searchQuery = ref("");
const statusFilter = ref<"all" | ShopOrder["status"]>("all");

// ─── Computed ─────────────────────────────────────────────────────────────────
const filteredOrders = computed(() => {
  let list = orders.value;
  if (statusFilter.value !== "all") {
    list = list.filter((o) => o.status === statusFilter.value);
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter(
      (o) =>
        o.customer_name?.toLowerCase().includes(q) ||
        o.customer_email?.toLowerCase().includes(q) ||
        o.id.toLowerCase().includes(q),
    );
  }
  return list;
});

const counts = computed(() => ({
  all: orders.value.length,
  pending: orders.value.filter((o) => o.status === "pending").length,
  paid: orders.value.filter((o) => o.status === "paid").length,
  fulfilled: orders.value.filter((o) => o.status === "fulfilled").length,
  delivered: orders.value.filter((o) => o.status === "delivered").length,
  cancelled: orders.value.filter((o) => o.status === "cancelled").length,
}));

const revenue = computed(() =>
  orders.value
    .filter((o) => o.status !== "cancelled")
    .reduce((sum, o) => sum + o.total, 0),
);

const pendingRevenue = computed(() =>
  orders.value
    .filter((o) => o.status === "pending")
    .reduce((sum, o) => sum + o.total, 0),
);

const STATUS_CONFIG: Record<
  string,
  { label: string; class: string; icon: Component }
> = {
  pending: {
    label: "Pendiente",
    class: "bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300",
    icon: Clock,
  },
  paid: {
    label: "Pagada",
    class: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
    icon: CreditCard,
  },
  fulfilled: {
    label: "Despachada",
    class:
      "bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400",
    icon: Package,
  },
  delivered: {
    label: "Entregada",
    class:
      "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400",
    icon: CheckCircle2,
  },
  cancelled: {
    label: "Cancelada",
    class: "bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400",
    icon: XCircle,
  },
};

const filterTabs = [
  { key: "all", label: "Todas" },
  { key: "pending", label: "Pendientes" },
  { key: "paid", label: "Pagadas" },
  { key: "fulfilled", label: "Despachadas" },
  { key: "delivered", label: "Entregadas" },
  { key: "cancelled", label: "Canceladas" },
];

// ─── Helpers ──────────────────────────────────────────────────────────────────
function fmtCurrency(n: number) {
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
}

function fmtDate(d: string) {
  return new Date(d).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function shortId(id: string) {
  return id.slice(0, 8).toUpperCase();
}

interface ListResponse {
  data: ShopOrder[];
  total: number;
}

// ─── API ─────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    const resp = await useApiFetch<ListResponse>("/api/v1/shop-orders");
    orders.value = resp.data ?? [];
  } catch {
    loadError.value = "No se pudieron cargar las órdenes online";
  } finally {
    loading.value = false;
  }
}

onMounted(load);

// ─── CSV export ───────────────────────────────────────────────────────────────

function exportCSV() {
  const rows = filteredOrders.value;
  if (!rows.length) return;

  const headers = [
    "Pedido",
    "Fecha",
    "Cliente",
    "Email",
    "Teléfono",
    "Ciudad",
    "Estado",
    "Subtotal",
    "Costo envío",
    "Total",
  ];

  const statusLabel: Record<string, string> = {
    pending: "Pendiente",
    paid: "Pagada",
    fulfilled: "Despachada",
    delivered: "Entregada",
    cancelled: "Cancelada",
  };

  const escape = (v: string | number | null | undefined) => {
    const s = String(v ?? "").replace(/"/g, '""');
    return `"${s}"`;
  };

  const lines = [
    headers.map(escape).join(","),
    ...rows.map((o) =>
      [
        shortId(o.id),
        fmtDate(o.created_at),
        o.customer_name,
        o.customer_email,
        o.customer_phone ?? "",
        o.shipping_city ?? "",
        statusLabel[o.status] ?? o.status,
        o.subtotal.toFixed(2),
        o.shipping_cost.toFixed(2),
        o.total.toFixed(2),
      ]
        .map(escape)
        .join(","),
    ),
  ];

  const blob = new Blob(["﻿" + lines.join("\r\n")], {
    type: "text/csv;charset=utf-8;",
  });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  const today = new Date().toISOString().slice(0, 10);
  const suffix = statusFilter.value !== "all" ? `-${statusFilter.value}` : "";
  a.download = `pedidos${suffix}-${today}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}
</script>

<template>
  <!-- MOBILE -->
  <MobileScreensShopOrders
    v-if="isMobile"
    :orders="orders"
    :loading="loading"
    :load-error="loadError"
  />

  <!-- WEB -->
  <div v-else class="p-4 sm:p-6 space-y-5">
    <!-- Header -->
    <div class="flex items-center justify-between gap-4 flex-wrap">
      <div>
        <h1 class="text-2xl font-bold text-foreground flex items-center gap-2">
          <ShoppingCart class="w-6 h-6 text-primary" />
          Pedidos online
        </h1>
        <p class="text-sm text-muted-foreground mt-0.5">
          Pedidos recibidos desde tu tienda pública
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button
          type="button"
          :disabled="filteredOrders.length === 0"
          class="px-3 py-2 text-sm rounded-lg border hover:bg-muted transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
          @click="exportCSV"
        >
          <Download class="w-3.5 h-3.5" />
          Exportar CSV
          <span
            v-if="filteredOrders.length > 0"
            class="text-xs bg-muted text-muted-foreground rounded px-1"
            >{{ filteredOrders.length }}</span
          >
        </button>
        <button
          type="button"
          class="px-3 py-2 text-sm rounded-lg border hover:bg-muted transition-colors flex items-center gap-2 cursor-pointer"
          @click="load"
        >
          <Loader2 v-if="loading" class="w-3.5 h-3.5 animate-spin" />
          Actualizar
        </button>
      </div>
    </div>

    <!-- KPIs -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div class="border bg-card rounded-xl p-4">
        <p class="text-xs text-muted-foreground uppercase tracking-wide">
          Pendientes
        </p>
        <p class="text-2xl font-bold font-mono text-amber-600 mt-1">
          {{ counts.pending }}
        </p>
        <p class="text-xs text-muted-foreground mt-0.5">
          Bs.S {{ fmtCurrency(pendingRevenue) }} esperando
        </p>
      </div>
      <div class="border bg-card rounded-xl p-4">
        <p class="text-xs text-muted-foreground uppercase tracking-wide">
          Por despachar
        </p>
        <p class="text-2xl font-bold font-mono text-blue-600 mt-1">
          {{ counts.paid }}
        </p>
        <p class="text-xs text-muted-foreground mt-0.5">Pagadas, sin enviar</p>
      </div>
      <div class="border bg-card rounded-xl p-4">
        <p class="text-xs text-muted-foreground uppercase tracking-wide">
          En camino
        </p>
        <p class="text-2xl font-bold font-mono text-purple-600 mt-1">
          {{ counts.fulfilled }}
        </p>
        <p class="text-xs text-muted-foreground mt-0.5">Esperando entrega</p>
      </div>
      <div class="border bg-card rounded-xl p-4">
        <p class="text-xs text-muted-foreground uppercase tracking-wide">
          Ingresos totales
        </p>
        <p class="text-2xl font-bold font-mono text-emerald-600 mt-1">
          {{ fmtCurrency(revenue) }}
        </p>
        <p class="text-xs text-muted-foreground mt-0.5">
          <TrendingUp class="w-3 h-3 inline" />
          {{ orders.length }} pedidos en total
        </p>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex items-center gap-3 flex-wrap">
      <div class="relative flex-1 min-w-48">
        <Search
          class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground"
        />
        <input
          v-model="searchQuery"
          class="w-full h-10 pl-9 pr-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
          placeholder="Buscar por cliente, email o # de pedido..."
        />
      </div>
      <div class="flex gap-1 bg-muted/50 rounded-lg p-1 overflow-x-auto">
        <button
          v-for="tab in filterTabs"
          :key="tab.key"
          :class="[
            'px-3 py-1.5 rounded-md text-sm font-medium transition-colors whitespace-nowrap cursor-pointer flex items-center gap-1.5',
            statusFilter === tab.key
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground',
          ]"
          @click="statusFilter = tab.key as typeof statusFilter"
        >
          {{ tab.label }}
          <span
            v-if="counts[tab.key as keyof typeof counts] > 0"
            :class="[
              'text-[10px] px-1.5 py-0.5 rounded-full font-bold',
              statusFilter === tab.key
                ? 'bg-primary/10 text-primary'
                : 'bg-muted text-muted-foreground',
            ]"
          >
            {{ counts[tab.key as keyof typeof counts] }}
          </span>
        </button>
      </div>
    </div>

    <!-- Error -->
    <div
      v-if="loadError"
      class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3"
    >
      {{ loadError }}
    </div>

    <!-- Loading -->
    <div
      v-if="loading && orders.length === 0"
      class="flex items-center justify-center py-20 text-muted-foreground"
    >
      <Loader2 class="w-6 h-6 animate-spin mr-2" />
      Cargando pedidos...
    </div>

    <!-- Empty -->
    <div
      v-else-if="filteredOrders.length === 0"
      class="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3 border-2 border-dashed border-border rounded-xl"
    >
      <FolderOpen class="w-12 h-12 opacity-30" />
      <p class="text-sm font-semibold text-foreground">
        {{
          searchQuery || statusFilter !== "all"
            ? "No hay pedidos que coincidan con el filtro"
            : "Aún no tenés pedidos online"
        }}
      </p>
      <p
        v-if="!searchQuery && statusFilter === 'all'"
        class="text-xs max-w-md text-center"
      >
        Cuando un cliente compre en tu tienda pública, su pedido aparecerá acá.
      </p>
    </div>

    <!-- Orders table -->
    <div v-else class="border bg-card rounded-xl overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border bg-muted/30">
              <th class="px-4 py-3 text-left font-medium text-muted-foreground">
                Pedido
              </th>
              <th class="px-4 py-3 text-left font-medium text-muted-foreground">
                Cliente
              </th>
              <th class="px-4 py-3 text-left font-medium text-muted-foreground">
                Estado
              </th>
              <th
                class="px-4 py-3 text-left font-medium text-muted-foreground hidden md:table-cell"
              >
                Fecha
              </th>
              <th
                class="px-4 py-3 text-right font-medium text-muted-foreground"
              >
                Total
              </th>
              <th class="px-4 py-3" />
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            <tr
              v-for="o in filteredOrders"
              :key="o.id"
              class="hover:bg-muted/20 transition-colors group"
            >
              <td class="px-4 py-3">
                <p class="font-mono font-semibold text-foreground text-xs">
                  #{{ shortId(o.id) }}
                </p>
                <p v-if="o.shipping_city" class="text-xs text-muted-foreground">
                  <MapPin class="w-3 h-3 inline" />
                  {{ o.shipping_city }}
                </p>
              </td>
              <td class="px-4 py-3">
                <p class="font-medium text-foreground">
                  {{ o.customer_name }}
                </p>
                <p
                  class="text-xs text-muted-foreground flex items-center gap-1"
                >
                  <Mail class="w-3 h-3" />
                  {{ o.customer_email }}
                </p>
              </td>
              <td class="px-4 py-3">
                <span
                  :class="[
                    'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium',
                    STATUS_CONFIG[o.status]?.class,
                  ]"
                >
                  <component
                    :is="STATUS_CONFIG[o.status]?.icon"
                    class="w-3 h-3"
                  />
                  {{ STATUS_CONFIG[o.status]?.label }}
                </span>
              </td>
              <td
                class="px-4 py-3 text-muted-foreground hidden md:table-cell text-xs"
              >
                {{ fmtDate(o.created_at) }}
              </td>
              <td
                class="px-4 py-3 text-right font-mono font-medium text-foreground"
              >
                Bs.S {{ fmtCurrency(o.total) }}
              </td>
              <td class="px-4 py-3 text-right">
                <NuxtLink
                  :to="`/shop-orders/${o.id}`"
                  class="opacity-0 group-hover:opacity-100 p-1.5 rounded-lg hover:bg-muted transition-all text-muted-foreground inline-flex"
                  :aria-label="`Ver pedido ${shortId(o.id)}`"
                >
                  <ChevronRight class="w-4 h-4" />
                </NuxtLink>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
