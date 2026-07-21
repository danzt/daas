<script setup lang="ts">
import {
  RefreshCw,
  DollarSign,
  Store,
  ShoppingBag,
  ShoppingCart,
  AlertTriangle,
  Package,
  ClipboardList,
  ChevronRight,
  Loader2,
} from "lucide-vue-next";

/**
 * MobileScreensDashboard — presentational mobile dashboard.
 * Pure: receives data + flags as props, emits `reload`. No fetching here,
 * so it can be driven by the real page OR a mock preview harness.
 */

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
export interface DashboardData {
  today: PeriodStats;
  month: PeriodStats;
  pending: PendingStats;
  inventory: InventoryStats;
  recent_shop_orders: ShopOrderSummary[];
}

const props = defineProps<{
  data: DashboardData | null;
  loading: boolean;
  loadError: string;
  tenantName: string;
}>();

const emit = defineEmits<{ reload: [] }>();

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

const SHOP_STATUS: Record<string, { label: string; class: string }> = {
  pending: { label: "Pendiente", class: "bg-amber-50 text-amber-700" },
  paid: { label: "Pagada", class: "bg-blue-100 text-blue-700" },
  fulfilled: { label: "Preparada", class: "bg-purple-100 text-purple-700" },
  delivered: { label: "Entregada", class: "bg-emerald-100 text-emerald-700" },
  cancelled: { label: "Cancelada", class: "bg-rose-100 text-rose-600" },
};

const monthRevenue = computed(() =>
  fmtCurrency(
    (props.data?.month.b2b_revenue ?? 0) + (props.data?.month.b2c_revenue ?? 0),
  ),
);
const monthB2B = computed(() =>
  fmtCurrency(props.data?.month.b2b_revenue ?? 0),
);
const monthB2C = computed(() =>
  fmtCurrency(props.data?.month.b2c_revenue ?? 0),
);
const stockAttention = computed(
  () =>
    (props.data?.inventory.low_stock ?? 0) +
    (props.data?.inventory.out_of_stock ?? 0),
);

const quickActions = [
  { to: "/sales-orders/new", icon: ShoppingBag, label: "Vender" },
  { to: "/shop-orders", icon: ShoppingCart, label: "Tienda" },
  { to: "/inventory", icon: Package, label: "Stock" },
  { to: "/reports", icon: ClipboardList, label: "Reportes" },
];
</script>

<template>
  <MobileScreen>
    <template #header>
      <header
        class="sticky top-0 z-30 bg-primary"
        style="padding-top: env(safe-area-inset-top)"
      >
        <div class="flex h-14 items-center justify-between px-5">
          <div class="min-w-0">
            <p class="text-xs leading-none text-white/60">Hola,</p>
            <p class="truncate text-base font-bold leading-tight text-white">
              {{ tenantName }}
            </p>
          </div>
          <button
            type="button"
            aria-label="Actualizar"
            :disabled="loading"
            class="no-min-tap flex size-10 items-center justify-center rounded-full text-white/80 active:bg-white/15 disabled:opacity-50"
            @click="emit('reload')"
          >
            <RefreshCw :class="['size-5', loading ? 'animate-spin' : '']" />
          </button>
        </div>
      </header>
    </template>

    <!-- Purple revenue hero -->
    <div class="bg-primary px-5 pb-9 pt-1 text-white">
      <p class="text-sm text-white/70">Ingresos del mes</p>
      <p
        class="mt-1 text-[34px] font-black leading-none tracking-tight tabular-nums"
      >
        {{ monthRevenue }}
      </p>
      <div class="mt-3 flex items-center gap-2">
        <span
          class="rounded-full bg-white/15 px-2.5 py-1 text-xs font-semibold text-white"
        >
          B2B {{ monthB2B }}
        </span>
        <span
          class="rounded-full bg-white/15 px-2.5 py-1 text-xs font-semibold text-white"
        >
          Tienda {{ monthB2C }}
        </span>
      </div>
    </div>

    <!-- White sheet sliding over the hero -->
    <div class="-mt-5 min-h-full rounded-t-3xl bg-background pt-4">
      <div v-if="loadError" class="px-4 pb-2">
        <div
          class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
        >
          {{ loadError }}
        </div>
      </div>

      <!-- Today summary -->
      <MobileSectionHeader title="Resumen de hoy" />
      <div class="grid grid-cols-2 gap-3 px-4">
        <MobileStatTile
          :icon="DollarSign"
          tone="primary"
          label="Ingresos hoy (B2B)"
          :value="fmtCurrency(data?.today.b2b_revenue ?? 0)"
          :sub="`${data?.today.invoice_count ?? 0} facturas`"
        />
        <MobileStatTile
          :icon="Store"
          tone="success"
          label="Tienda hoy"
          :value="fmtCurrency(data?.today.b2c_revenue ?? 0)"
          sub="Órdenes pagadas"
        />
        <MobileStatTile
          :icon="ShoppingBag"
          tone="primary"
          label="Órdenes B2B"
          :value="String(data?.pending.sales_orders ?? 0)"
          sub="Sin facturar"
          to="/sales-orders?status=confirmed"
        />
        <MobileStatTile
          :icon="ShoppingCart"
          tone="warning"
          label="Pedidos tienda"
          :value="String(data?.pending.shop_orders ?? 0)"
          sub="Sin pagar"
          to="/shop-orders?status=pending"
        />
      </div>

      <!-- Stock alert -->
      <div v-if="stockAttention > 0" class="px-4 pt-3">
        <NuxtLink
          to="/inventory"
          class="flex items-center gap-3 rounded-2xl border border-red-100 bg-red-50 p-3.5 active:bg-red-100"
        >
          <div
            class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-red-100 text-red-500"
          >
            <AlertTriangle class="size-5" />
          </div>
          <div class="min-w-0 flex-1">
            <p class="text-sm font-bold text-red-700">
              {{ stockAttention }} productos requieren atención
            </p>
            <p class="text-xs text-red-500">
              {{ data?.inventory.out_of_stock ?? 0 }} sin stock ·
              {{ data?.inventory.low_stock ?? 0 }} bajo mínimo
            </p>
          </div>
          <ChevronRight class="size-5 shrink-0 text-red-300" />
        </NuxtLink>
      </div>

      <!-- Quick actions -->
      <MobileSectionHeader title="Accesos rápidos" class="pt-4" />
      <div class="grid grid-cols-4 gap-2 px-4">
        <NuxtLink
          v-for="action in quickActions"
          :key="action.to"
          :to="action.to"
          class="flex flex-col items-center gap-1.5 rounded-2xl border border-border bg-white py-3 transition-transform active:scale-95"
        >
          <div
            class="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary"
          >
            <component :is="action.icon" class="size-5" />
          </div>
          <span class="text-[11px] font-semibold text-foreground">{{
            action.label
          }}</span>
        </NuxtLink>
      </div>

      <!-- Recent shop orders -->
      <MobileSectionHeader
        title="Pedidos recientes"
        to="/shop-orders"
        class="pt-5"
      />
      <div v-if="loading && !data" class="flex justify-center py-10">
        <Loader2 class="size-6 animate-spin text-muted-foreground" />
      </div>
      <div
        v-else-if="!data?.recent_shop_orders.length"
        class="px-4 py-10 text-center text-sm text-muted-foreground"
      >
        Sin pedidos aún
      </div>
      <div
        v-else
        class="divide-y divide-border border-y border-border bg-white"
      >
        <MobileListItem
          v-for="order in data.recent_shop_orders"
          :key="order.id"
          :to="`/shop-orders/${order.id}`"
        >
          <template #leading>
            <div
              class="flex size-10 items-center justify-center rounded-xl bg-muted text-muted-foreground"
            >
              <ShoppingCart class="size-5" />
            </div>
          </template>
          <p class="truncate text-sm font-semibold text-foreground">
            {{
              order.customer_name || order.customer_email || "Cliente anónimo"
            }}
          </p>
          <p class="text-xs text-muted-foreground">
            {{ fmtDate(order.created_at) }}
          </p>
          <template #trailing>
            <div class="flex flex-col items-end gap-1">
              <span class="font-mono text-sm font-bold tabular-nums">{{
                fmtCurrency(order.total)
              }}</span>
              <span
                :class="[
                  'rounded px-1.5 py-px text-[10px] font-bold',
                  SHOP_STATUS[order.status]?.class,
                ]"
              >
                {{ SHOP_STATUS[order.status]?.label ?? order.status }}
              </span>
            </div>
          </template>
        </MobileListItem>
      </div>
    </div>
  </MobileScreen>
</template>
