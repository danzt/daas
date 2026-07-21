<script setup lang="ts">
import {
  Clock,
  CreditCard,
  Package,
  CheckCircle2,
  XCircle,
  FolderOpen,
  Loader2,
} from "lucide-vue-next";

interface ShopOrder {
  id: string;
  customer_name: string;
  customer_email: string;
  total: number;
  status: "pending" | "paid" | "fulfilled" | "delivered" | "cancelled";
  created_at: string;
}

const props = defineProps<{
  orders: ShopOrder[];
  loading: boolean;
  loadError: string;
}>();

const search = ref("");
const statusFilter = ref("all");

const STATUS: Record<
  string,
  { label: string; chip: string; bubble: string; fg: string; icon: Component }
> = {
  pending: {
    label: "Pendiente",
    chip: "bg-amber-50 text-amber-700",
    bubble: "bg-amber-50",
    fg: "text-amber-600",
    icon: Clock,
  },
  paid: {
    label: "Pagada",
    chip: "bg-blue-100 text-blue-700",
    bubble: "bg-blue-50",
    fg: "text-blue-600",
    icon: CreditCard,
  },
  fulfilled: {
    label: "Despachada",
    chip: "bg-purple-100 text-purple-700",
    bubble: "bg-purple-50",
    fg: "text-purple-600",
    icon: Package,
  },
  delivered: {
    label: "Entregada",
    chip: "bg-emerald-100 text-emerald-700",
    bubble: "bg-emerald-50",
    fg: "text-emerald-600",
    icon: CheckCircle2,
  },
  cancelled: {
    label: "Cancelada",
    chip: "bg-rose-100 text-rose-600",
    bubble: "bg-rose-50",
    fg: "text-rose-500",
    icon: XCircle,
  },
};

const revenue = computed(() =>
  props.orders
    .filter((o) => o.status !== "cancelled")
    .reduce((s, o) => s + o.total, 0),
);
const pendingRevenue = computed(() =>
  props.orders
    .filter((o) => o.status === "pending")
    .reduce((s, o) => s + o.total, 0),
);

const segOptions = computed(() => [
  { key: "all", label: "Todas", count: props.orders.length },
  {
    key: "pending",
    label: "Pendientes",
    count: props.orders.filter((o) => o.status === "pending").length,
  },
  {
    key: "paid",
    label: "Pagadas",
    count: props.orders.filter((o) => o.status === "paid").length,
  },
  {
    key: "fulfilled",
    label: "Despachadas",
    count: props.orders.filter((o) => o.status === "fulfilled").length,
  },
  {
    key: "delivered",
    label: "Entregadas",
    count: props.orders.filter((o) => o.status === "delivered").length,
  },
]);

const filtered = computed(() => {
  let list = props.orders;
  if (statusFilter.value !== "all")
    list = list.filter((o) => o.status === statusFilter.value);
  const q = search.value.trim().toLowerCase();
  if (q)
    list = list.filter(
      (o) =>
        (o.customer_name ?? "").toLowerCase().includes(q) ||
        (o.customer_email ?? "").toLowerCase().includes(q),
    );
  return list;
});

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
</script>

<template>
  <MobileScreen title="Pedidos tienda" subtitle="Ventas online B2C">
    <!-- Revenue strip -->
    <div class="grid grid-cols-2 gap-3 px-4 pb-1 pt-3">
      <div class="rounded-2xl border border-border bg-white p-3.5">
        <p class="text-[19px] font-black tabular-nums text-foreground">
          ${{ fmtCurrency(revenue) }}
        </p>
        <p class="text-xs text-muted-foreground">Ingresos</p>
      </div>
      <div class="rounded-2xl border border-amber-100 bg-amber-50 p-3.5">
        <p class="text-[19px] font-black tabular-nums text-amber-600">
          ${{ fmtCurrency(pendingRevenue) }}
        </p>
        <p class="text-xs text-amber-500">Por cobrar</p>
      </div>
    </div>

    <MobileSearchBar
      v-model="search"
      placeholder="Buscar por cliente o email…"
    />
    <MobileSegment v-model="statusFilter" :options="segOptions" />

    <!-- Error -->
    <div
      v-if="loadError"
      class="mx-4 mt-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
    >
      {{ loadError }}
    </div>

    <!-- Loading -->
    <div
      v-else-if="loading && !orders.length"
      class="flex justify-center py-16"
    >
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <!-- Empty -->
    <div v-else-if="!filtered.length" class="px-6 py-16 text-center">
      <FolderOpen class="mx-auto mb-4 size-12 text-muted-foreground/20" />
      <p class="font-medium text-muted-foreground">Sin pedidos</p>
    </div>

    <!-- List -->
    <div v-else class="divide-y divide-border border-t border-border bg-white">
      <MobileListItem
        v-for="order in filtered"
        :key="order.id"
        :to="`/shop-orders/${order.id}`"
      >
        <template #leading>
          <div
            class="flex size-11 items-center justify-center rounded-2xl"
            :class="STATUS[order.status]?.bubble"
          >
            <component
              :is="STATUS[order.status]?.icon"
              class="size-5"
              :class="STATUS[order.status]?.fg"
            />
          </div>
        </template>

        <p class="truncate text-sm font-semibold leading-tight text-foreground">
          {{ order.customer_name || order.customer_email || "Cliente anónimo" }}
        </p>
        <div class="mt-1 flex items-center gap-1.5">
          <span
            class="rounded px-1.5 py-px text-[10px] font-bold"
            :class="STATUS[order.status]?.chip"
            >{{ STATUS[order.status]?.label }}</span
          >
          <span class="text-[11px] text-muted-foreground">{{
            fmtDate(order.created_at)
          }}</span>
        </div>

        <template #trailing>
          <span
            class="font-mono text-sm font-bold tabular-nums text-foreground"
          >
            {{ fmtCurrency(order.total) }}
          </span>
        </template>
      </MobileListItem>
    </div>
  </MobileScreen>
</template>
