<script setup lang="ts">
import { ShoppingBag, FolderOpen, Loader2 } from "lucide-vue-next";

interface SaleOrder {
  id: string;
  status: "draft" | "confirmed" | "invoiced" | "cancelled";
  customer_name: string;
  total: number;
  notes: string;
  created_at: string;
}

const props = defineProps<{
  orders: SaleOrder[];
  loading: boolean;
  loadError: string;
}>();

const search = ref("");
const statusFilter = ref("all");

const STATUS: Record<string, { label: string; chip: string; accent: string }> =
  {
    draft: {
      label: "Borrador",
      chip: "bg-gray-100 text-gray-600",
      accent: "bg-gray-400",
    },
    confirmed: {
      label: "Confirmada",
      chip: "bg-blue-100 text-blue-700",
      accent: "bg-blue-500",
    },
    invoiced: {
      label: "Facturada",
      chip: "bg-emerald-100 text-emerald-700",
      accent: "bg-primary",
    },
    cancelled: {
      label: "Cancelada",
      chip: "bg-rose-100 text-rose-600",
      accent: "bg-red-400",
    },
  };

const segOptions = computed(() => [
  { key: "all", label: "Todas", count: props.orders.length },
  {
    key: "draft",
    label: "Borradores",
    count: props.orders.filter((o) => o.status === "draft").length,
  },
  {
    key: "confirmed",
    label: "Confirmadas",
    count: props.orders.filter((o) => o.status === "confirmed").length,
  },
  {
    key: "invoiced",
    label: "Facturadas",
    count: props.orders.filter((o) => o.status === "invoiced").length,
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
        o.id.toLowerCase().includes(q),
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
  <MobileScreen title="Órdenes de venta" subtitle="Ventas B2B">
    <MobileSearchBar v-model="search" placeholder="Buscar por cliente…" />
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
      <p class="font-medium text-muted-foreground">
        {{
          search || statusFilter !== "all"
            ? "Sin órdenes que coincidan"
            : "Aún no hay órdenes de venta"
        }}
      </p>
    </div>

    <!-- List -->
    <div v-else class="divide-y divide-border border-t border-border bg-white">
      <MobileListItem
        v-for="order in filtered"
        :key="order.id"
        :to="`/sales-orders/${order.id}`"
      >
        <template #leading>
          <div class="relative">
            <div
              class="flex size-11 items-center justify-center rounded-2xl"
              :class="{
                'bg-gray-100': order.status === 'draft',
                'bg-blue-50': order.status === 'confirmed',
                'bg-primary/10': order.status === 'invoiced',
                'bg-red-50': order.status === 'cancelled',
              }"
            >
              <ShoppingBag
                class="size-5"
                :class="{
                  'text-gray-500': order.status === 'draft',
                  'text-blue-600': order.status === 'confirmed',
                  'text-primary': order.status === 'invoiced',
                  'text-red-500': order.status === 'cancelled',
                }"
              />
            </div>
          </div>
        </template>

        <p class="truncate text-sm font-semibold leading-tight text-foreground">
          {{ order.customer_name || "Cliente anónimo" }}
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

    <!-- FAB -->
    <MobileFab to="/sales-orders/new" label="Nueva orden" />
  </MobileScreen>
</template>
