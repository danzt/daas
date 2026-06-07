<script setup lang="ts">
import {
  Warehouse,
  History,
  AlertTriangle,
  FolderOpen,
  Search,
} from "lucide-vue-next";
import type { Product } from "~/components/products/ProductFormModal.vue";

export interface StockRow {
  product_id: string;
  tenant_id: string;
  quantity_on_hand: number;
  last_updated_at: string;
  product: Product | null;
}

const props = defineProps<{
  rows: StockRow[];
  loading: boolean;
  loadError: string;
  isOwner: boolean;
  lowStockThreshold: number;
}>();

const emit = defineEmits<{ adjust: [row: StockRow]; retry: [] }>();

const search = ref("");
const filter = ref("all"); // all | low

const stats = computed(() => {
  const total = props.rows.length;
  const low = props.rows.filter(
    (r) =>
      r.quantity_on_hand > 0 && r.quantity_on_hand < props.lowStockThreshold,
  ).length;
  const out = props.rows.filter((r) => r.quantity_on_hand === 0).length;
  const value = props.rows.reduce((acc, r) => {
    const price = r.product?.is_fiscal
      ? (r.product?.fiscal_price ?? 0)
      : (r.product?.internal_price ?? 0);
    return acc + price * r.quantity_on_hand;
  }, 0);
  return { total, low, out, value };
});

const segOptions = computed(() => [
  { key: "all", label: "Todos", count: props.rows.length },
  { key: "low", label: "Stock bajo", count: stats.value.low + stats.value.out },
]);

const filtered = computed(() => {
  let list = props.rows.filter((r) => r.product);
  const q = search.value.trim().toLowerCase();
  if (q) {
    list = list.filter(
      (r) =>
        (r.product?.name ?? "").toLowerCase().includes(q) ||
        (r.product?.sku ?? "").toLowerCase().includes(q),
    );
  }
  if (filter.value === "low") {
    list = list.filter((r) => r.quantity_on_hand < props.lowStockThreshold);
  }
  return list;
});

function fmtMoney(n: number) {
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
}

function priceOf(p: Product | null) {
  if (!p) return 0;
  return p.is_fiscal ? (p.fiscal_price ?? 0) : (p.internal_price ?? 0);
}

function sev(qty: number): "out" | "low" | "ok" {
  if (qty === 0) return "out";
  if (qty < props.lowStockThreshold) return "low";
  return "ok";
}
</script>

<template>
  <MobileScreen title="Inventario" subtitle="Stock y movimientos">
    <template #action>
      <NuxtLink
        to="/inventory/movements"
        aria-label="Historial"
        class="no-min-tap flex size-10 items-center justify-center rounded-full text-foreground active:bg-accent"
      >
        <History class="size-5" />
      </NuxtLink>
    </template>

    <!-- Stat strip -->
    <div class="no-scrollbar flex gap-3 overflow-x-auto px-4 pb-1 pt-3">
      <div
        class="flex min-w-[120px] shrink-0 flex-col rounded-2xl border border-border bg-white p-3.5"
      >
        <span class="text-[22px] font-black tabular-nums text-foreground">{{
          stats.total
        }}</span>
        <span class="text-xs text-muted-foreground">Productos</span>
      </div>
      <div
        class="flex min-w-[120px] shrink-0 flex-col rounded-2xl border border-orange-100 bg-orange-50 p-3.5"
      >
        <span class="text-[22px] font-black tabular-nums text-orange-600">{{
          stats.low
        }}</span>
        <span class="text-xs text-orange-500">Stock bajo</span>
      </div>
      <div
        class="flex min-w-[120px] shrink-0 flex-col rounded-2xl border border-red-100 bg-red-50 p-3.5"
      >
        <span class="text-[22px] font-black tabular-nums text-red-600">{{
          stats.out
        }}</span>
        <span class="text-xs text-red-500">Sin stock</span>
      </div>
      <div
        class="flex min-w-[140px] shrink-0 flex-col rounded-2xl border border-border bg-white p-3.5"
      >
        <span class="text-[22px] font-black tabular-nums text-foreground"
          >${{ fmtMoney(stats.value) }}</span
        >
        <span class="text-xs text-muted-foreground">Valor</span>
      </div>
    </div>

    <!-- Search + filter -->
    <MobileSearchBar v-model="search" placeholder="Buscar por nombre o SKU…" />
    <MobileSegment v-model="filter" :options="segOptions" />

    <!-- Error -->
    <div v-if="loadError" class="px-4 py-12 text-center">
      <AlertTriangle class="mx-auto mb-3 size-10 text-destructive" />
      <p class="font-medium text-destructive">{{ loadError }}</p>
      <button
        type="button"
        class="mt-4 h-10 rounded-xl border border-border px-5 text-sm font-semibold text-muted-foreground active:bg-accent"
        @click="emit('retry')"
      >
        Reintentar
      </button>
    </div>

    <!-- Loading -->
    <div v-else-if="loading && !rows.length" class="space-y-2 px-4 pt-2">
      <div
        v-for="n in 7"
        :key="n"
        class="h-[72px] animate-pulse rounded-2xl bg-muted"
      />
    </div>

    <!-- Empty -->
    <div v-else-if="!filtered.length" class="px-6 py-16 text-center">
      <FolderOpen class="mx-auto mb-4 size-12 text-muted-foreground/20" />
      <p class="font-medium text-muted-foreground">Sin resultados</p>
    </div>

    <!-- Stock list -->
    <div v-else class="divide-y divide-border border-t border-border bg-white">
      <MobileListItem
        v-for="row in filtered"
        :key="row.product_id"
        :chevron="false"
        :accent="
          sev(row.quantity_on_hand) === 'out'
            ? 'danger'
            : sev(row.quantity_on_hand) === 'low'
              ? 'warning'
              : 'success'
        "
      >
        <template #leading>
          <div
            class="flex size-12 flex-col items-center justify-center rounded-2xl"
            :class="{
              'bg-red-50': sev(row.quantity_on_hand) === 'out',
              'bg-orange-50': sev(row.quantity_on_hand) === 'low',
              'bg-emerald-50': sev(row.quantity_on_hand) === 'ok',
            }"
          >
            <span
              class="text-lg font-black leading-none tabular-nums"
              :class="{
                'text-red-600': sev(row.quantity_on_hand) === 'out',
                'text-orange-600': sev(row.quantity_on_hand) === 'low',
                'text-emerald-700': sev(row.quantity_on_hand) === 'ok',
              }"
              >{{ row.quantity_on_hand.toFixed(0) }}</span
            >
            <span
              class="text-[9px] font-semibold uppercase leading-none"
              :class="{
                'text-red-400': sev(row.quantity_on_hand) === 'out',
                'text-orange-400': sev(row.quantity_on_hand) === 'low',
                'text-emerald-500': sev(row.quantity_on_hand) === 'ok',
              }"
              >uds</span
            >
          </div>
        </template>

        <p class="truncate text-sm font-semibold leading-tight text-foreground">
          {{ row.product?.name ?? "—" }}
        </p>
        <div class="mt-1 flex items-center gap-1.5">
          <span
            v-if="row.product?.sku"
            class="rounded bg-muted px-1.5 py-px font-mono text-[10px] text-muted-foreground"
            >{{ row.product.sku }}</span
          >
          <span class="text-[11px] text-muted-foreground"
            >${{ fmtMoney(priceOf(row.product)) }}</span
          >
        </div>

        <template #trailing>
          <button
            v-if="isOwner"
            type="button"
            class="no-min-tap h-9 rounded-xl border border-primary/30 bg-primary/5 px-3.5 text-xs font-bold text-primary active:bg-primary/15"
            @click.stop="emit('adjust', row)"
          >
            Ajustar
          </button>
        </template>
      </MobileListItem>
    </div>
  </MobileScreen>
</template>
