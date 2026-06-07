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
  Store,
  ShoppingCart,
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
const { isMobile } = useMobileMode();

// ─── Types ────────────────────────────────────────────────────────────────────

interface PeriodStats {
  b2b_revenue: number;
  b2c_revenue: number;
  invoice_count: number;
}

interface PendingStats {
  shop_orders: number;
  sales_orders: number;
}

interface InventoryStats {
  low_stock: number;
  out_of_stock: number;
}

interface ShopOrderSummary {
  id: string;
  customer_name: string;
  customer_email: string;
  total: number;
  status: string;
  created_at: string;
}

interface DashboardResponse {
  today: PeriodStats;
  month: PeriodStats;
  pending: PendingStats;
  inventory: InventoryStats;
  recent_shop_orders: ShopOrderSummary[];
}

// ─── State ────────────────────────────────────────────────────────────────────

const loading = ref(false);
const loadError = ref("");
const data = ref<DashboardResponse | null>(null);

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
  });
}

// ─── Status badge config ──────────────────────────────────────────────────────

const B2B_STATUS: Record<string, { label: string; class: string }> = {
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

const SHOP_STATUS: Record<string, { label: string; class: string }> = {
  pending: {
    label: "Pendiente",
    class: "bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300",
  },
  paid: {
    label: "Pagada",
    class: "bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300",
  },
  fulfilled: {
    label: "Preparada",
    class:
      "bg-purple-50 text-purple-700 dark:bg-purple-950 dark:text-purple-300",
  },
  delivered: {
    label: "Entregada",
    class:
      "bg-emerald-50 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300",
  },
  cancelled: {
    label: "Cancelada",
    class: "bg-rose-50 text-rose-700 dark:bg-rose-950 dark:text-rose-300",
  },
};

// ─── Load ─────────────────────────────────────────────────────────────────────

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    data.value = await useApiFetch<DashboardResponse>("/api/v1/dashboard");
  } catch {
    loadError.value = "No se pudieron cargar los datos del dashboard";
  } finally {
    loading.value = false;
  }
}

// ─── Computed stats cards ─────────────────────────────────────────────────────

const stats = computed(() => [
  {
    icon: DollarSign,
    description: "Ingresos hoy (B2B)",
    value: fmtCurrency(data.value?.today.b2b_revenue ?? 0),
    badge: {
      label: `${data.value?.today.invoice_count ?? 0} ${(data.value?.today.invoice_count ?? 0) === 1 ? "factura" : "facturas"}`,
      trend: "up" as const,
    },
    sub:
      (data.value?.today.invoice_count ?? 0) === 0
        ? "Sin actividad de ventas hoy"
        : "Ventas registradas en el día",
  },
  {
    icon: Store,
    description: "Ingresos tienda hoy",
    value: fmtCurrency(data.value?.today.b2c_revenue ?? 0),
    badge:
      (data.value?.today.b2c_revenue ?? 0) > 0
        ? { label: "Tienda online", trend: "up" as const }
        : null,
    sub: "Órdenes pagadas / entregadas hoy",
  },
  {
    icon: ShoppingBag,
    description: "Órdenes B2B pendientes",
    value: String(data.value?.pending.sales_orders ?? 0),
    badge:
      (data.value?.pending.sales_orders ?? 0) > 0
        ? { label: "Confirmadas", trend: "up" as const }
        : null,
    sub: "Esperando facturación",
    to: "/sales-orders?status=confirmed",
  },
  {
    icon: ShoppingCart,
    description: "Pedidos tienda",
    value: String(data.value?.pending.shop_orders ?? 0),
    badge:
      (data.value?.pending.shop_orders ?? 0) > 0
        ? { label: "Sin pagar", trend: "down" as const }
        : null,
    sub: "Pedidos pendientes de pago",
    to: "/shop-orders?status=pending",
  },
  {
    icon: AlertTriangle,
    description: "Stock bajo",
    value: String(
      (data.value?.inventory.low_stock ?? 0) +
        (data.value?.inventory.out_of_stock ?? 0),
    ),
    badge:
      (data.value?.inventory.out_of_stock ?? 0) > 0
        ? { label: "Crítico", trend: "down" as const }
        : null,
    sub: `${data.value?.inventory.out_of_stock ?? 0} sin stock · ${data.value?.inventory.low_stock ?? 0} bajo`,
    to: "/inventory",
  },
  {
    icon: FileText,
    description: "Facturas del mes",
    value: String(data.value?.month.invoice_count ?? 0),
    badge:
      (data.value?.month.invoice_count ?? 0) > 0
        ? { label: "Acumulado", trend: "up" as const }
        : null,
    sub: `Ingresos B2B: ${fmtCurrency(data.value?.month.b2b_revenue ?? 0)}`,
    to: "/invoices",
  },
]);

const monthRevenue = computed(() =>
  fmtCurrency(
    (data.value?.month.b2b_revenue ?? 0) + (data.value?.month.b2c_revenue ?? 0),
  ),
);

const monthB2BRevenue = computed(() =>
  fmtCurrency(data.value?.month.b2b_revenue ?? 0),
);
const monthB2CRevenue = computed(() =>
  fmtCurrency(data.value?.month.b2c_revenue ?? 0),
);

onMounted(load);
</script>

<template>
  <!-- ════════════════════════════════════════════════════════════════
       MOBILE — exclusive native screen (presentational component)
  ═════════════════════════════════════════════════════════════════ -->
  <MobileScreensDashboard
    v-if="isMobile"
    :data="data"
    :loading="loading"
    :load-error="loadError"
    :tenant-name="store.tenant?.name || store.user?.email || ''"
    @reload="load"
  />

  <!-- ════════════════════════════════════════════════════════════════
       WEB — unchanged desktop dashboard
  ═════════════════════════════════════════════════════════════════ -->
  <div v-else class="p-4 sm:p-6 space-y-6 max-w-7xl mx-auto">
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

    <!-- Stats grid — 6 cards: 2 cols mobile, 3 cols md, 6 cols xl -->
    <div
      class="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-6"
    >
      <component
        :is="stat.to ? 'NuxtLink' : 'div'"
        v-for="stat in stats"
        :key="stat.description"
        :to="stat.to"
        :class="stat.to ? 'block transition-shadow hover:shadow-md' : ''"
      >
        <Card class="h-full">
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

    <!-- Month revenue summary -->
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <Card class="lg:col-span-2">
        <CardHeader>
          <CardDescription>Ingresos del mes (total)</CardDescription>
          <CardTitle class="text-4xl font-medium tabular-nums tracking-tight">
            {{ monthRevenue }}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p class="text-sm text-muted-foreground">
            B2B {{ monthB2BRevenue }} · Tienda {{ monthB2CRevenue }} ·
            {{ data?.month.invoice_count ?? 0 }} facturas emitidas
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
          <CardDescription>Inventario</CardDescription>
        </CardHeader>
        <CardContent class="flex flex-col gap-1">
          <div
            class="text-3xl font-medium tabular-nums leading-none tracking-tight"
          >
            {{
              (data?.inventory.low_stock ?? 0) +
              (data?.inventory.out_of_stock ?? 0)
            }}
          </div>
          <p class="text-sm text-muted-foreground">
            {{ data?.inventory.out_of_stock ?? 0 }} sin stock ·
            {{ data?.inventory.low_stock ?? 0 }} bajo mínimo
          </p>
        </CardContent>
      </Card>
    </div>

    <!-- B2B pending + Recent shop orders -->
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <!-- B2B: confirmed, not yet invoiced -->
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
          <div v-if="loading && !data" class="px-4 py-8 flex justify-center">
            <Loader2 class="w-5 h-5 animate-spin text-muted-foreground" />
          </div>
          <div
            v-else-if="(data?.pending.sales_orders ?? 0) === 0"
            class="px-4 py-8 text-center text-sm text-muted-foreground"
          >
            Sin órdenes pendientes
          </div>
          <div v-else class="px-4 py-6 text-center">
            <p class="text-3xl font-semibold tabular-nums">
              {{ data?.pending.sales_orders }}
            </p>
            <p class="text-sm text-muted-foreground mt-1">
              órdenes confirmadas sin facturar
            </p>
            <NuxtLink
              to="/sales-orders?status=confirmed"
              class="mt-3 inline-flex items-center gap-1 text-xs text-primary hover:underline"
            >
              Gestionar <ArrowRight class="size-3" />
            </NuxtLink>
          </div>
        </CardContent>
      </Card>

      <!-- B2C: recent shop orders -->
      <Card>
        <CardHeader class="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Pedidos recientes (tienda)</CardTitle>
            <CardDescription>Últimos 5 pedidos online</CardDescription>
          </div>
          <NuxtLink
            to="/shop-orders"
            class="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
          >
            Ver todos
            <ArrowRight class="size-3" />
          </NuxtLink>
        </CardHeader>
        <CardContent class="px-0">
          <div v-if="loading && !data" class="px-4 py-8 flex justify-center">
            <Loader2 class="w-5 h-5 animate-spin text-muted-foreground" />
          </div>
          <div
            v-else-if="!data?.recent_shop_orders.length"
            class="px-4 py-8 text-center text-sm text-muted-foreground"
          >
            Sin pedidos aún
          </div>
          <ul v-else class="divide-y">
            <li
              v-for="order in data.recent_shop_orders"
              :key="order.id"
              class="flex items-center justify-between px-4 py-3 hover:bg-accent/40 transition-colors"
            >
              <div class="min-w-0">
                <NuxtLink
                  :to="`/shop-orders/${order.id}`"
                  class="text-sm font-medium hover:underline truncate block"
                >
                  {{
                    order.customer_name ||
                    order.customer_email ||
                    "Cliente anónimo"
                  }}
                </NuxtLink>
                <p class="text-xs text-muted-foreground">
                  {{ fmtDate(order.created_at) }}
                </p>
              </div>
              <div class="flex items-center gap-3 shrink-0 ml-2">
                <span class="font-mono text-sm tabular-nums">
                  {{ fmtCurrency(order.total) }}
                </span>
                <span
                  :class="[
                    'inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium',
                    SHOP_STATUS[order.status]?.class,
                  ]"
                >
                  {{ SHOP_STATUS[order.status]?.label ?? order.status }}
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
          { to: '/shop-orders', icon: ShoppingCart, label: 'Tienda online' },
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
