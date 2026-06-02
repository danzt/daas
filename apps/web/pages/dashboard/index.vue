<script setup lang="ts">
import {
  ShoppingBag,
  FileText,
  AlertTriangle,
  Package,
  DollarSign,
  ClipboardList,
  ArrowRight,
  Loader2,
  RefreshCw,
  TrendingUp,
  TrendingDown,
} from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";
import { useApiFetch } from "~/composables/useAuth";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "~/components/ui/card";
import { Badge } from "~/components/ui/badge";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

const store = useAuthStore();

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

const loading = ref(false);
const loadError = ref("");

const todaySales = ref<SalesSummary | null>(null);
const monthSales = ref<SalesSummary | null>(null);
const inventory = ref<InventorySnapshot | null>(null);
const pendingOrders = ref<SaleOrder[]>([]);
const recentOrders = ref<SaleOrder[]>([]);

function todayRange() {
  const now = new Date();
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  return {
    from: start.toISOString().slice(0, 10),
    to: new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)
      .toISOString()
      .slice(0, 10),
  };
}

function monthRange() {
  const now = new Date();
  return {
    from: new Date(now.getFullYear(), now.getMonth(), 1)
      .toISOString()
      .slice(0, 10),
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
  draft: { label: "Borrador", class: "bg-secondary text-secondary-foreground" },
  confirmed: {
    label: "Confirmada",
    class: "bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300",
  },
  invoiced: {
    label: "Facturada",
    class:
      "bg-emerald-50 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300",
  },
  cancelled: {
    label: "Cancelada",
    class: "bg-rose-50 text-rose-700 dark:bg-rose-950 dark:text-rose-300",
  },
};

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

const stats = computed(() => [
  {
    icon: DollarSign,
    description: "Ingresos hoy",
    value: fmtCurrency(todaySales.value?.total_revenue ?? 0),
    badge: {
      label: `${todaySales.value?.total_invoices ?? 0} ${(todaySales.value?.total_invoices ?? 0) === 1 ? "factura" : "facturas"}`,
      trend: "up" as const,
    },
    sub:
      (todaySales.value?.total_invoices ?? 0) === 0
        ? "Sin actividad de ventas hoy"
        : "Ventas registradas en el día",
  },
  {
    icon: ShoppingBag,
    description: "Órdenes pendientes",
    value: String(pendingOrders.value.length),
    badge: pendingOrders.value.length
      ? { label: "Confirmadas", trend: "up" as const }
      : null,
    sub: "Esperando facturación",
    to: "/sales-orders?status=confirmed",
  },
  {
    icon: AlertTriangle,
    description: "Stock bajo",
    value: String(
      (inventory.value?.low_stock_count ?? 0) +
        (inventory.value?.out_of_stock_count ?? 0),
    ),
    badge:
      (inventory.value?.out_of_stock_count ?? 0) > 0
        ? { label: "Crítico", trend: "down" as const }
        : null,
    sub: `${inventory.value?.out_of_stock_count ?? 0} sin stock · ${inventory.value?.low_stock_count ?? 0} bajo`,
    to: "/inventory",
  },
  {
    icon: FileText,
    description: "Facturas del mes",
    value: String(monthSales.value?.total_invoices ?? 0),
    badge: monthSales.value?.total_invoices
      ? { label: "Acumulado", trend: "up" as const }
      : null,
    sub: `${monthSales.value?.internal_count ?? 0} internas · ${monthSales.value?.fiscal_count ?? 0} fiscales`,
    to: "/invoices",
  },
]);

const monthRevenue = computed(() =>
  fmtCurrency(monthSales.value?.total_revenue ?? 0),
);

onMounted(load);
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Dashboard</h1>
        <p class="text-sm text-muted-foreground mt-1">
          Bienvenido,
          <span class="font-medium text-foreground">{{
            store.tenant?.name || store.user?.email
          }}</span>
        </p>
      </div>
      <button
        type="button"
        :disabled="loading"
        class="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors disabled:opacity-40"
        @click="load"
      >
        <RefreshCw :class="['w-4 h-4', loading ? 'animate-spin' : '']" />
      </button>
    </div>

    <!-- Error -->
    <div
      v-if="loadError"
      class="rounded-lg border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
    >
      {{ loadError }}
    </div>

    <!-- Stats grid -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <component
        :is="stat.to ? 'NuxtLink' : 'div'"
        v-for="stat in stats"
        :key="stat.description"
        :to="stat.to"
        :class="stat.to ? 'block transition-shadow hover:shadow-md' : ''"
      >
        <Card>
          <CardHeader>
            <CardTitle>
              <div
                class="flex size-7 items-center justify-center rounded-lg border bg-muted text-muted-foreground"
              >
                <component :is="stat.icon" class="size-4" />
              </div>
            </CardTitle>
            <CardDescription>{{ stat.description }}</CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-1">
            <div class="flex flex-wrap items-center gap-2">
              <div
                class="text-3xl font-medium tabular-nums leading-none tracking-tight"
              >
                {{ stat.value }}
              </div>
              <Badge
                v-if="stat.badge"
                :variant="stat.badge.trend === 'down' ? 'danger' : 'secondary'"
              >
                <TrendingUp
                  v-if="stat.badge.trend === 'up'"
                  class="size-3 mr-1"
                />
                <TrendingDown v-else class="size-3 mr-1" />
                {{ stat.badge.label }}
              </Badge>
            </div>
            <p class="text-sm text-muted-foreground">{{ stat.sub }}</p>
          </CardContent>
        </Card>
      </component>
    </div>

    <!-- Month revenue + Inventory snapshot -->
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <Card class="lg:col-span-2">
        <CardHeader>
          <CardDescription>Ingresos del mes</CardDescription>
          <CardTitle class="text-4xl font-medium tabular-nums tracking-tight">
            {{ monthRevenue }}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p class="text-sm text-muted-foreground">
            {{ monthSales?.total_invoices ?? 0 }} facturas emitidas ·
            {{ monthSales?.internal_count ?? 0 }} internas ·
            {{ monthSales?.fiscal_count ?? 0 }} fiscales
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>
            <div
              class="flex size-7 items-center justify-center rounded-lg border bg-muted text-muted-foreground"
            >
              <Package class="size-4" />
            </div>
          </CardTitle>
          <CardDescription>Productos activos</CardDescription>
        </CardHeader>
        <CardContent class="flex flex-col gap-1">
          <div
            class="text-3xl font-medium tabular-nums leading-none tracking-tight"
          >
            {{ inventory?.total_products ?? 0 }}
          </div>
          <p class="text-sm text-muted-foreground">Catálogo total del tenant</p>
        </CardContent>
      </Card>
    </div>

    <!-- Pending + Recent rows -->
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Confirmadas sin facturar</CardTitle>
            <CardDescription>Listas para emitir factura</CardDescription>
          </div>
          <NuxtLink
            to="/sales-orders"
            class="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
          >
            Ver todas
            <ArrowRight class="size-3" />
          </NuxtLink>
        </CardHeader>
        <CardContent class="px-0">
          <div
            v-if="loading && !pendingOrders.length"
            class="px-4 py-8 flex justify-center"
          >
            <Loader2 class="w-5 h-5 animate-spin text-muted-foreground" />
          </div>
          <div
            v-else-if="pendingOrders.length === 0"
            class="px-4 py-8 text-center text-sm text-muted-foreground"
          >
            Sin órdenes pendientes
          </div>
          <ul v-else class="divide-y">
            <li
              v-for="order in pendingOrders.slice(0, 5)"
              :key="order.id"
              class="flex items-center justify-between px-4 py-3 hover:bg-accent/40 transition-colors"
            >
              <div>
                <NuxtLink
                  :to="`/sales-orders/${order.id}`"
                  class="text-sm font-medium hover:underline"
                >
                  {{ order.customer_name || "Cliente anónimo" }}
                </NuxtLink>
                <p class="text-xs text-muted-foreground">
                  {{ fmtDate(order.created_at) }}
                </p>
              </div>
              <div class="flex items-center gap-3">
                <span class="font-mono text-sm tabular-nums">
                  {{ fmtCurrency(order.total) }}
                </span>
                <NuxtLink
                  :to="`/sales-orders/${order.id}`"
                  class="text-xs text-muted-foreground hover:text-foreground hover:underline"
                >
                  Facturar
                </NuxtLink>
              </div>
            </li>
          </ul>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Órdenes recientes</CardTitle>
            <CardDescription>Últimas 5 órdenes registradas</CardDescription>
          </div>
          <NuxtLink
            to="/sales-orders"
            class="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
          >
            Ver todas
            <ArrowRight class="size-3" />
          </NuxtLink>
        </CardHeader>
        <CardContent class="px-0">
          <div
            v-if="loading && !recentOrders.length"
            class="px-4 py-8 flex justify-center"
          >
            <Loader2 class="w-5 h-5 animate-spin text-muted-foreground" />
          </div>
          <div
            v-else-if="recentOrders.length === 0"
            class="px-4 py-8 text-center text-sm text-muted-foreground"
          >
            Sin órdenes aún
          </div>
          <ul v-else class="divide-y">
            <li
              v-for="order in recentOrders"
              :key="order.id"
              class="flex items-center justify-between px-4 py-3 hover:bg-accent/40 transition-colors"
            >
              <div>
                <NuxtLink
                  :to="`/sales-orders/${order.id}`"
                  class="text-sm font-medium hover:underline"
                >
                  {{ order.customer_name || "Cliente anónimo" }}
                </NuxtLink>
                <p class="text-xs text-muted-foreground">
                  {{ fmtDate(order.created_at) }}
                </p>
              </div>
              <div class="flex items-center gap-3">
                <span class="font-mono text-sm tabular-nums">
                  {{ fmtCurrency(order.total) }}
                </span>
                <span
                  :class="[
                    'inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium',
                    STATUS_CONFIG[order.status]?.class,
                  ]"
                >
                  {{ STATUS_CONFIG[order.status]?.label }}
                </span>
              </div>
            </li>
          </ul>
        </CardContent>
      </Card>
    </div>

    <!-- Quick actions -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <NuxtLink
        v-for="action in [
          { to: '/sales-orders', icon: ShoppingBag, label: 'Nueva venta' },
          { to: '/invoices', icon: FileText, label: 'Facturas' },
          { to: '/inventory', icon: Package, label: 'Inventario' },
          { to: '/reports', icon: ClipboardList, label: 'Reportes' },
        ]"
        :key="action.to"
        :to="action.to"
        class="rounded-xl border bg-card p-4 flex flex-col items-center gap-2 hover:bg-accent transition-colors text-center"
      >
        <div
          class="flex size-9 items-center justify-center rounded-lg border bg-muted text-muted-foreground"
        >
          <component :is="action.icon" class="size-4" />
        </div>
        <span class="text-xs font-medium">{{ action.label }}</span>
      </NuxtLink>
    </div>
  </div>
</template>
