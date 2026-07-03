<script setup lang="ts">
import {
  RefreshCw,
  Download,
  Loader2,
  AlertTriangle,
  TrendingUp,
  Boxes,
  ShoppingCart,
  Receipt,
} from "lucide-vue-next";

interface ProductRevenue {
  product_id: string;
  product_name: string;
  total_sold: number;
  revenue: number;
}
interface SalesSummary {
  total_revenue: number;
  total_invoices: number;
  internal_count: number;
  fiscal_count: number;
  top_products: ProductRevenue[];
}
interface InventorySnapshot {
  total_products: number;
  low_stock_count: number;
  out_of_stock_count: number;
}
interface SupplierSpend {
  supplier_id: string;
  supplier_name: string;
  total_spend: number;
  order_count: number;
}
interface PurchaseInvoiceEntry {
  date: string;
  number: string;
  supplier_name: string;
  supplier_rif: string;
  tax_base: number;
  tax_amount: number;
  total: number;
}

interface PurchaseSummary {
  total_spend: number;
  total_orders: number;
  received_orders: number;
  by_supplier: SupplierSpend[];
  total_tax_base: number;
  total_tax_credit: number;
  invoices: PurchaseInvoiceEntry[];
}

defineProps<{
  sales: SalesSummary | null;
  inventory: InventorySnapshot | null;
  purchases: PurchaseSummary | null;
  loading: boolean;
  loadError: string;
}>();

const emit = defineEmits<{
  reload: [];
  exportSales: [];
  exportInventory: [];
  exportPurchases: [];
}>();

const fromDate = defineModel<string>("fromDate", { required: true });
const toDate = defineModel<string>("toDate", { required: true });

const tab = ref("sales");
const tabs = [
  { key: "sales", label: "Ventas" },
  { key: "inventory", label: "Inventario" },
  { key: "purchases", label: "Compras" },
];

function fmt(n: number) {
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
  <MobileScreen title="Reportes" subtitle="Resumen del negocio">
    <template #action>
      <button
        type="button"
        aria-label="Actualizar"
        :disabled="loading"
        class="no-min-tap flex size-10 items-center justify-center rounded-full text-foreground active:bg-accent disabled:opacity-50"
        @click="emit('reload')"
      >
        <RefreshCw :class="['size-5', loading ? 'animate-spin' : '']" />
      </button>
    </template>

    <MobileSegment v-model="tab" :options="tabs" />

    <!-- Date range -->
    <div class="px-4 pt-1">
      <div
        class="flex items-center gap-2 rounded-2xl border border-border bg-white p-2"
      >
        <input
          v-model="fromDate"
          type="date"
          aria-label="Desde"
          class="h-10 w-0 flex-1 rounded-xl bg-muted/50 px-2 text-xs text-foreground focus:outline-none"
          @change="emit('reload')"
        />
        <span class="shrink-0 text-xs text-muted-foreground">→</span>
        <input
          v-model="toDate"
          type="date"
          aria-label="Hasta"
          class="h-10 w-0 flex-1 rounded-xl bg-muted/50 px-2 text-xs text-foreground focus:outline-none"
          @change="emit('reload')"
        />
      </div>
    </div>

    <!-- Error -->
    <div v-if="loadError" class="px-4 py-12 text-center">
      <AlertTriangle class="mx-auto mb-3 size-10 text-destructive" />
      <p class="font-medium text-destructive">{{ loadError }}</p>
    </div>

    <!-- Loading -->
    <div v-else-if="loading" class="flex justify-center py-16">
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <template v-else>
      <!-- ── Sales ── -->
      <div v-if="tab === 'sales'" class="px-4 pt-4">
        <div class="rounded-2xl bg-primary p-5 text-white">
          <p class="text-sm text-white/70">Ingresos del período</p>
          <p class="mt-1 text-[30px] font-black leading-none tabular-nums">
            {{ fmt(sales?.total_revenue ?? 0) }}
          </p>
          <div class="mt-3 flex gap-2">
            <span
              class="rounded-full bg-white/15 px-2.5 py-1 text-xs font-semibold"
              >{{ sales?.total_invoices ?? 0 }} facturas</span
            >
            <span
              class="rounded-full bg-white/15 px-2.5 py-1 text-xs font-semibold"
              >{{ sales?.fiscal_count ?? 0 }} fiscales</span
            >
          </div>
        </div>

        <MobileSectionHeader title="Top productos" class="px-0 pt-5" />
        <div
          v-if="sales?.top_products?.length"
          class="overflow-hidden rounded-2xl border border-border bg-white"
        >
          <div
            v-for="(prod, i) in sales.top_products.slice(0, 8)"
            :key="prod.product_id"
            class="flex items-center gap-3 border-b border-border px-4 py-3 last:border-0"
          >
            <span
              class="flex size-7 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-xs font-black text-primary"
              >{{ i + 1 }}</span
            >
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-semibold text-foreground">
                {{ prod.product_name }}
              </p>
              <p class="text-[11px] text-muted-foreground">
                {{ prod.total_sold }} uds vendidas
              </p>
            </div>
            <span class="font-mono text-sm font-bold tabular-nums">{{
              fmt(prod.revenue)
            }}</span>
          </div>
        </div>
        <p v-else class="px-1 py-6 text-center text-sm text-muted-foreground">
          Sin ventas en el período
        </p>

        <button
          type="button"
          class="mt-4 flex w-full items-center justify-center gap-2 rounded-2xl border border-primary/30 bg-primary/5 py-3 text-sm font-bold text-primary active:bg-primary/10"
          @click="emit('exportSales')"
        >
          <Download class="size-4" /> Exportar CSV
        </button>
      </div>

      <!-- ── Inventory ── -->
      <div v-else-if="tab === 'inventory'" class="px-4 pt-4">
        <div class="grid grid-cols-3 gap-3">
          <div
            class="rounded-2xl border border-border bg-white p-3.5 text-center"
          >
            <Boxes class="mx-auto mb-1 size-5 text-muted-foreground" />
            <p class="text-[22px] font-black tabular-nums text-foreground">
              {{ inventory?.total_products ?? 0 }}
            </p>
            <p class="text-[11px] text-muted-foreground">Productos</p>
          </div>
          <div
            class="rounded-2xl border border-orange-100 bg-orange-50 p-3.5 text-center"
          >
            <p class="text-[22px] font-black tabular-nums text-orange-600">
              {{ inventory?.low_stock_count ?? 0 }}
            </p>
            <p class="text-[11px] text-orange-500">Stock bajo</p>
          </div>
          <div
            class="rounded-2xl border border-red-100 bg-red-50 p-3.5 text-center"
          >
            <p class="text-[22px] font-black tabular-nums text-red-600">
              {{ inventory?.out_of_stock_count ?? 0 }}
            </p>
            <p class="text-[11px] text-red-500">Sin stock</p>
          </div>
        </div>
        <button
          type="button"
          class="mt-4 flex w-full items-center justify-center gap-2 rounded-2xl border border-primary/30 bg-primary/5 py-3 text-sm font-bold text-primary active:bg-primary/10"
          @click="emit('exportInventory')"
        >
          <Download class="size-4" /> Exportar inventario CSV
        </button>
      </div>

      <!-- ── Purchases ── -->
      <div v-else class="px-4 pt-4">
        <div class="rounded-2xl bg-foreground p-5 text-white">
          <p class="text-sm text-white/60">Gasto en compras</p>
          <p class="mt-1 text-[30px] font-black leading-none tabular-nums">
            {{ fmt(purchases?.total_spend ?? 0) }}
          </p>
          <div class="mt-3 flex gap-2">
            <span
              class="rounded-full bg-white/15 px-2.5 py-1 text-xs font-semibold"
              >{{ purchases?.total_orders ?? 0 }} órdenes</span
            >
            <span
              class="rounded-full bg-white/15 px-2.5 py-1 text-xs font-semibold"
              >{{ purchases?.received_orders ?? 0 }} recibidas</span
            >
          </div>
          <div
            class="mt-3 flex items-center justify-between border-t border-white/15 pt-3"
          >
            <span class="text-sm text-white/60">Crédito fiscal (IVA)</span>
            <span class="font-mono text-lg font-black tabular-nums">{{
              fmt(purchases?.total_tax_credit ?? 0)
            }}</span>
          </div>
        </div>

        <MobileSectionHeader title="Por proveedor" class="px-0 pt-5" />
        <div
          v-if="purchases?.by_supplier?.length"
          class="overflow-hidden rounded-2xl border border-border bg-white"
        >
          <div
            v-for="sup in purchases.by_supplier.slice(0, 8)"
            :key="sup.supplier_id"
            class="flex items-center gap-3 border-b border-border px-4 py-3 last:border-0"
          >
            <div
              class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground"
            >
              <ShoppingCart class="size-4" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-semibold text-foreground">
                {{ sup.supplier_name }}
              </p>
              <p class="text-[11px] text-muted-foreground">
                {{ sup.order_count }} órdenes
              </p>
            </div>
            <span class="font-mono text-sm font-bold tabular-nums">{{
              fmt(sup.total_spend)
            }}</span>
          </div>
        </div>
        <p v-else class="px-1 py-6 text-center text-sm text-muted-foreground">
          Sin compras en el período
        </p>

        <!-- Libro de compras -->
        <MobileSectionHeader title="Libro de compras" class="px-0 pt-5" />
        <div
          v-if="purchases?.invoices?.length"
          class="overflow-hidden rounded-2xl border border-border bg-white"
        >
          <div
            v-for="(inv, idx) in purchases.invoices"
            :key="idx"
            class="flex items-center gap-3 border-b border-border px-4 py-3 last:border-0"
          >
            <div
              class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"
            >
              <Receipt class="size-4" />
            </div>
            <div class="min-w-0 flex-1">
              <p
                class="truncate font-mono text-sm font-semibold text-foreground"
              >
                {{ inv.number }}
              </p>
              <p class="truncate text-[11px] text-muted-foreground">
                {{ fmtDate(inv.date) }} · {{ inv.supplier_name }}
              </p>
            </div>
            <div class="text-right">
              <p class="font-mono text-sm font-bold tabular-nums text-primary">
                {{ fmt(inv.tax_amount) }}
              </p>
              <p class="text-[10px] text-muted-foreground">IVA</p>
            </div>
          </div>
        </div>
        <p v-else class="px-1 py-6 text-center text-sm text-muted-foreground">
          Sin facturas de compra en el período
        </p>

        <button
          type="button"
          class="mt-4 flex w-full items-center justify-center gap-2 rounded-2xl border border-primary/30 bg-primary/5 py-3 text-sm font-bold text-primary active:bg-primary/10"
          @click="emit('exportPurchases')"
        >
          <Download class="size-4" /> Exportar compras CSV
        </button>
      </div>
    </template>
  </MobileScreen>
</template>
