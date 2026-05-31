<script setup lang="ts">
import {
  ShoppingBag,
  FileText,
  AlertTriangle,
  Package,
  TrendingUp,
  ClipboardList,
  ArrowRight,
  Loader2,
  RefreshCw,
} from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";
import { useApiFetch } from "~/composables/useAuth";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

const store = useAuthStore();

// ─── Types ────────────────────────────────────────────────────────────────────

interface SalesSummary {
  from: string;
  to: string;
  total_revenue: number;
  total_invoices: number;
  internal_count: number;
  fiscal_count: number;
}

interface InventorySnapshot {
  total_products: number;
  low_stock_count: number;
  out_of_stock_count: number;
}

interface SaleOrder {
  id: string;
  status: string;
  customer_name: string;
  total: number;
  created_at: string;
}

// ─── State ────────────────────────────────────────────────────────────────────

const loading = ref(false);
const loadError = ref("");

const todaySales = ref<SalesSummary | null>(null);
const monthSales = ref<SalesSummary | null>(null);
const inventory = ref<InventorySnapshot | null>(null);
const pendingOrders = ref<SaleOrder[]>([]);
const recentOrders = ref<SaleOrder[]>([]);

// ─── Helpers ─────────────────────────────────────────────────────────────────

function todayRange() {
  const now = new Date();
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const end = new Date(
    now.getFullYear(),
    now.getMonth(),
    now.getDate(),
    23,
    59,
    59,
  );
  return {
    from: start.toISOString().slice(0, 10),
    to: end.toISOString().slice(0, 10),
  };
}

function monthRange() {
  const now = new Date();
  const start = new Date(now.getFullYear(), now.getMonth(), 1);
  return {
    from: start.toISOString().slice(0, 10),
    to: new Date().toISOString().slice(0, 10),
  };
}

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
  });
}

const STATUS_CONFIG: Record<string, { label: string; class: string }> = {
  draft: { label: "Borrador", class: "bg-muted text-muted-foreground" },
  confirmed: {
    label: "Confirmada",
    class: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
  },
  invoiced: {
    label: "Facturada",
    class:
      "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400",
  },
  cancelled: {
    label: "Cancelada",
    class: "bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400",
  },
};

// ─── Load ─────────────────────────────────────────────────────────────────────

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    const { from: tFrom, to: tTo } = todayRange();
    const { from: mFrom, to: mTo } = monthRange();

    const [ts, ms, inv, orders] = await Promise.all([
      useApiFetch<SalesSummary>(
        `/api/v1/reports/sales?from=${tFrom}&to=${tTo}`,
      ),
      useApiFetch<SalesSummary>(
        `/api/v1/reports/sales?from=${mFrom}&to=${mTo}`,
      ),
      useApiFetch<InventorySnapshot>("/api/v1/reports/inventory"),
      useApiFetch<SaleOrder[]>("/api/v1/sales-orders"),
    ]);

    todaySales.value = ts;
    monthSales.value = ms;
    inventory.value = inv;
    pendingOrders.value = orders.filter((o) => o.status === "confirmed");
    recentOrders.value = orders.slice(0, 5);
  } catch {
    loadError.value = "No se pudieron cargar los datos del dashboard";
  } finally {
    loading.value = false;
  }
}

// ─── Computed stats ───────────────────────────────────────────────────────────

const stats = computed(() => [
  {
    label: "Ingresos hoy",
    value: fmtCurrency(todaySales.value?.total_revenue ?? 0),
    sub:
      (todaySales.value?.total_invoices ?? 0) === 0
        ? "Sin facturas hoy"
        : `${todaySales.value!.total_invoices} factura${todaySales.value!.total_invoices !== 1 ? "s" : ""}`,
    icon: TrendingUp,
    color: "text-cta",
    bg: "bg-cta/10",
  },
  {
    label: "Órdenes pendientes",
    value: String(pendingOrders.value.length),
    sub: "Confirmadas sin facturar",
    icon: ShoppingBag,
    color: "text-blue-600",
    bg: "bg-blue-100 dark:bg-blue-900/30",
    to: "/sales-orders?status=confirmed",
  },
  {
    label: "Stock bajo",
    value: String(
      (inventory.value?.low_stock_count ?? 0) +
        (inventory.value?.out_of_stock_count ?? 0),
    ),
    sub: `${inventory.value?.out_of_stock_count ?? 0} sin stock · ${inventory.value?.low_stock_count ?? 0} bajo`,
    icon: AlertTriangle,
    color:
      (inventory.value?.out_of_stock_count ?? 0) > 0
        ? "text-rose-600"
        : "text-yellow-600",
    bg:
      (inventory.value?.out_of_stock_count ?? 0) > 0
        ? "bg-rose-100 dark:bg-rose-900/30"
        : "bg-yellow-100 dark:bg-yellow-900/30",
    to: "/inventory",
  },
  {
    label: "Facturas este mes",
    value: String(monthSales.value?.total_invoices ?? 0),
    sub: `${monthSales.value?.internal_count ?? 0} internas · ${monthSales.value?.fiscal_count ?? 0} fiscales`,
    icon: FileText,
    color: "text-primary",
    bg: "bg-primary/10",
    to: "/invoices",
  },
]);

// ─── Month revenue summary ────────────────────────────────────────────────────

const monthRevenue = computed(() =>
  fmtCurrency(monthSales.value?.total_revenue ?? 0),
);

onMounted(load);
</script>

<template>
  <div class="p-6 space-y-8">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold font-heading text-foreground">
          Dashboard
        </h1>
        <p class="text-muted-foreground mt-0.5">
          Bienvenido,
          <span class="font-semibold text-foreground">{{
            store.tenant?.name || store.user?.email
          }}</span>
        </p>
      </div>
      <button
        :disabled="loading"
        class="p-2 rounded-lg hover:bg-muted transition-colors text-muted-foreground disabled:opacity-40"
        @click="load"
      >
        <RefreshCw :class="['w-4 h-4', loading ? 'animate-spin' : '']" />
      </button>
    </div>

    <!-- Error -->
    <div
      v-if="loadError"
      class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3"
    >
      {{ loadError }}
    </div>

    <!-- Loading skeleton -->
    <div
      v-if="loading && !todaySales"
      class="grid grid-cols-2 lg:grid-cols-4 gap-4"
    >
      <div
        v-for="i in 4"
        :key="i"
        class="bg-white rounded-xl border p-5 animate-pulse"
      >
        <div class="h-3 bg-muted rounded w-2/3 mb-4" />
        <div class="h-8 bg-muted rounded w-1/2 mb-2" />
        <div class="h-3 bg-muted rounded w-3/4" />
      </div>
    </div>

    <!-- Stats grid -->
    <div v-else class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <component
        :is="stat.to ? 'NuxtLink' : 'div'"
        v-for="stat in stats"
        :key="stat.label"
        :to="stat.to"
        :class="[
          'bg-white rounded-xl border border-border p-5 transition-shadow',
          stat.to ? 'hover:shadow-md cursor-pointer' : '',
        ]"
      >
        <div class="flex items-center justify-between mb-3">
          <span
            class="text-xs font-semibold text-muted-foreground uppercase tracking-wide"
          >
            {{ stat.label }}
          </span>
          <div
            :class="[
              'w-8 h-8 rounded-lg flex items-center justify-center',
              stat.bg,
            ]"
          >
            <component :is="stat.icon" :class="['w-4 h-4', stat.color]" />
          </div>
        </div>
        <p :class="['text-3xl font-bold font-heading', stat.color]">
          {{ stat.value }}
        </p>
        <p class="text-xs text-muted-foreground mt-1.5">{{ stat.sub }}</p>
      </component>
    </div>

    <!-- Month revenue banner -->
    <div
      v-if="!loading || monthSales"
      class="bg-primary rounded-xl p-6 flex items-center justify-between"
    >
      <div>
        <p class="text-white/70 text-sm font-medium mb-1">Ingresos del mes</p>
        <p class="text-4xl font-bold text-white font-heading">
          {{ monthRevenue }}
        </p>
        <p class="text-white/60 text-xs mt-1.5">
          {{ monthSales?.total_invoices ?? 0 }} facturas emitidas ·
          {{ monthSales?.internal_count ?? 0 }} internas ·
          {{ monthSales?.fiscal_count ?? 0 }} fiscales
        </p>
      </div>
      <div class="hidden sm:flex flex-col items-end gap-2">
        <Package class="w-12 h-12 text-white/20" />
        <p class="text-white/40 text-xs">
          {{ inventory?.total_products ?? 0 }} productos activos
        </p>
      </div>
    </div>

    <!-- Pending orders + Recent orders row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Pending sale orders -->
      <div class="bg-white border border-gray-100 rounded-xl overflow-hidden">
        <div
          class="flex items-center justify-between px-5 py-4 border-b border-border"
        >
          <h3 class="font-semibold text-foreground text-sm">
            Órdenes confirmadas sin facturar
          </h3>
          <NuxtLink
            to="/sales-orders"
            class="text-xs text-primary hover:underline flex items-center gap-1"
          >
            Ver todas
            <ArrowRight class="w-3 h-3" />
          </NuxtLink>
        </div>
        <div
          v-if="loading && !pendingOrders.length"
          class="px-5 py-8 flex justify-center"
        >
          <Loader2 class="w-5 h-5 animate-spin text-muted-foreground" />
        </div>
        <div
          v-else-if="pendingOrders.length === 0"
          class="px-5 py-8 text-center text-sm text-muted-foreground"
        >
          Sin órdenes pendientes
        </div>
        <ul v-else class="divide-y divide-border">
          <li
            v-for="order in pendingOrders.slice(0, 5)"
            :key="order.id"
            class="flex items-center justify-between px-5 py-3 hover:bg-muted/20 transition-colors"
          >
            <div>
              <NuxtLink
                :to="`/sales-orders/${order.id}`"
                class="text-sm font-medium text-foreground hover:text-primary transition-colors"
              >
                {{ order.customer_name || "Cliente anónimo" }}
              </NuxtLink>
              <p class="text-xs text-muted-foreground">
                {{ fmtDate(order.created_at) }}
              </p>
            </div>
            <div class="flex items-center gap-3">
              <span class="font-mono text-sm font-medium text-foreground">
                {{ fmtCurrency(order.total) }}
              </span>
              <NuxtLink
                :to="`/sales-orders/${order.id}`"
                class="text-xs text-primary hover:underline"
              >
                Facturar
              </NuxtLink>
            </div>
          </li>
        </ul>
      </div>

      <!-- Recent orders -->
      <div class="bg-white border border-gray-100 rounded-xl overflow-hidden">
        <div
          class="flex items-center justify-between px-5 py-4 border-b border-border"
        >
          <h3 class="font-semibold text-foreground text-sm">
            Órdenes recientes
          </h3>
          <NuxtLink
            to="/sales-orders"
            class="text-xs text-primary hover:underline flex items-center gap-1"
          >
            Ver todas
            <ArrowRight class="w-3 h-3" />
          </NuxtLink>
        </div>
        <div
          v-if="loading && !recentOrders.length"
          class="px-5 py-8 flex justify-center"
        >
          <Loader2 class="w-5 h-5 animate-spin text-muted-foreground" />
        </div>
        <div
          v-else-if="recentOrders.length === 0"
          class="px-5 py-8 text-center text-sm text-muted-foreground"
        >
          Sin órdenes aún
        </div>
        <ul v-else class="divide-y divide-border">
          <li
            v-for="order in recentOrders"
            :key="order.id"
            class="flex items-center justify-between px-5 py-3 hover:bg-muted/20 transition-colors"
          >
            <div>
              <NuxtLink
                :to="`/sales-orders/${order.id}`"
                class="text-sm font-medium text-foreground hover:text-primary transition-colors"
              >
                {{ order.customer_name || "Cliente anónimo" }}
              </NuxtLink>
              <p class="text-xs text-muted-foreground">
                {{ fmtDate(order.created_at) }}
              </p>
            </div>
            <div class="flex items-center gap-3">
              <span class="font-mono text-sm font-medium text-foreground">
                {{ fmtCurrency(order.total) }}
              </span>
              <span
                :class="[
                  'inline-flex items-center px-1.5 py-0.5 rounded-full text-xs font-medium',
                  STATUS_CONFIG[order.status]?.class,
                ]"
              >
                {{ STATUS_CONFIG[order.status]?.label }}
              </span>
            </div>
          </li>
        </ul>
      </div>
    </div>

    <!-- Quick actions -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <NuxtLink
        to="/sales-orders"
        class="bg-white border border-gray-100 rounded-xl p-4 flex flex-col items-center gap-2 hover:border-primary/40 hover:bg-primary/5 transition-colors group text-center"
      >
        <div
          class="p-2 bg-primary/10 rounded-lg group-hover:bg-primary/20 transition-colors"
        >
          <ShoppingBag class="w-5 h-5 text-primary" />
        </div>
        <span class="text-xs font-medium text-foreground">Nueva venta</span>
      </NuxtLink>
      <NuxtLink
        to="/invoices"
        class="bg-white border border-gray-100 rounded-xl p-4 flex flex-col items-center gap-2 hover:border-primary/40 hover:bg-primary/5 transition-colors group text-center"
      >
        <div
          class="p-2 bg-primary/10 rounded-lg group-hover:bg-primary/20 transition-colors"
        >
          <FileText class="w-5 h-5 text-primary" />
        </div>
        <span class="text-xs font-medium text-foreground">Facturas</span>
      </NuxtLink>
      <NuxtLink
        to="/inventory"
        class="bg-white border border-gray-100 rounded-xl p-4 flex flex-col items-center gap-2 hover:border-primary/40 hover:bg-primary/5 transition-colors group text-center"
      >
        <div
          class="p-2 bg-primary/10 rounded-lg group-hover:bg-primary/20 transition-colors"
        >
          <Package class="w-5 h-5 text-primary" />
        </div>
        <span class="text-xs font-medium text-foreground">Inventario</span>
      </NuxtLink>
      <NuxtLink
        to="/reports"
        class="bg-white border border-gray-100 rounded-xl p-4 flex flex-col items-center gap-2 hover:border-primary/40 hover:bg-primary/5 transition-colors group text-center"
      >
        <div
          class="p-2 bg-primary/10 rounded-lg group-hover:bg-primary/20 transition-colors"
        >
          <ClipboardList class="w-5 h-5 text-primary" />
        </div>
        <span class="text-xs font-medium text-foreground">Reportes</span>
      </NuxtLink>
    </div>
  </div>
</template>
