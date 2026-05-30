<script setup lang="ts">
import {
  BarChart2,
  TrendingUp,
  Package,
  ShoppingCart,
  RefreshCw,
  Loader2,
  AlertTriangle,
  TrendingDown,
  Boxes,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

// ─── Types ────────────────────────────────────────────────────────────────────

interface DayRevenue {
  date: string;
  revenue: number;
  count: number;
}

interface ProductRevenue {
  product_id: string;
  product_name: string;
  total_sold: number;
  revenue: number;
}

interface SalesSummary {
  from: string;
  to: string;
  total_revenue: number;
  total_invoices: number;
  internal_count: number;
  fiscal_count: number;
  by_day: DayRevenue[];
  top_products: ProductRevenue[];
}

interface StockItem {
  product_id: string;
  product_name: string;
  sku: string;
  category_name: string;
  quantity_on_hand: number;
  is_fiscal: boolean;
  active: boolean;
}

interface InventorySnapshot {
  generated_at: string;
  total_products: number;
  low_stock_count: number;
  out_of_stock_count: number;
  items: StockItem[];
}

interface SupplierSpend {
  supplier_id: string;
  supplier_name: string;
  total_spend: number;
  order_count: number;
}

interface PurchaseSummary {
  from: string;
  to: string;
  total_spend: number;
  total_orders: number;
  received_orders: number;
  by_supplier: SupplierSpend[];
}

// ─── State ───────────────────────────────────────────────────────────────────
const sales = ref<SalesSummary | null>(null);
const inventory = ref<InventorySnapshot | null>(null);
const purchases = ref<PurchaseSummary | null>(null);
const loading = ref(false);
const loadError = ref("");

// Date range
const today = new Date();
const thirtyDaysAgo = new Date(today);
thirtyDaysAgo.setDate(today.getDate() - 30);

const fromDate = ref(thirtyDaysAgo.toISOString().slice(0, 10));
const toDate = ref(today.toISOString().slice(0, 10));

const route = useRoute();
const tabFromQuery = route.query.tab as string | undefined;
const activeTab = ref<"sales" | "inventory" | "purchases">(
  tabFromQuery === "inventory" || tabFromQuery === "purchases"
    ? tabFromQuery
    : "sales",
);

// ─── Computed helpers ─────────────────────────────────────────────────────────
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
  });
}

// Chart bar heights (max height = 100%)
const maxDayRevenue = computed(() =>
  sales.value ? Math.max(...sales.value.by_day.map((d) => d.revenue), 1) : 1,
);

function barHeight(revenue: number) {
  return Math.round((revenue / maxDayRevenue.value) * 100);
}

// ─── API ─────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    const params = `?from=${fromDate.value}&to=${toDate.value}`;
    const [s, inv, p] = await Promise.all([
      useApiFetch<SalesSummary>(`/api/v1/reports/sales${params}`),
      useApiFetch<InventorySnapshot>(`/api/v1/reports/inventory`),
      useApiFetch<PurchaseSummary>(`/api/v1/reports/purchases${params}`),
    ]);
    sales.value = s;
    inventory.value = inv;
    purchases.value = p;
  } catch {
    loadError.value = "No se pudieron cargar los reportes";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold text-foreground">Reportes</h1>
        <p class="text-sm text-muted-foreground mt-0.5">
          Análisis de ventas, inventario y compras
        </p>
      </div>

      <!-- Date range + refresh -->
      <div class="flex items-center gap-2 flex-wrap">
        <div class="flex items-center gap-1 bg-muted/50 rounded-lg p-1 text-sm">
          <input
            v-model="fromDate"
            type="date"
            class="bg-transparent px-2 py-1 text-foreground focus:outline-none"
          />
          <span class="text-muted-foreground">→</span>
          <input
            v-model="toDate"
            type="date"
            class="bg-transparent px-2 py-1 text-foreground focus:outline-none"
          />
        </div>
        <button
          :disabled="loading"
          class="flex items-center gap-2 px-3 py-2 rounded-lg bg-primary text-primary-foreground text-sm hover:bg-primary/90 transition-colors disabled:opacity-50"
          @click="load"
        >
          <RefreshCw :class="['w-4 h-4', loading && 'animate-spin']" />
          Actualizar
        </button>
      </div>
    </div>

    <!-- Error -->
    <div
      v-if="loadError"
      class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3 flex items-center gap-2"
    >
      <AlertTriangle class="w-4 h-4" />
      {{ loadError }}
    </div>

    <!-- Loading -->
    <div
      v-if="loading"
      class="flex items-center justify-center py-20 text-muted-foreground"
    >
      <Loader2 class="w-6 h-6 animate-spin mr-2" />
      Cargando reportes...
    </div>

    <template v-else-if="sales && inventory && purchases">
      <!-- Summary KPIs -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="bg-card border border-border rounded-xl p-5">
          <div class="flex items-center justify-between mb-3">
            <p class="text-sm text-muted-foreground">Ingresos</p>
            <div class="p-2 bg-emerald-100 dark:bg-emerald-900/30 rounded-lg">
              <TrendingUp
                class="w-4 h-4 text-emerald-600 dark:text-emerald-400"
              />
            </div>
          </div>
          <p class="text-2xl font-bold text-foreground font-mono">
            {{ fmtCurrency(sales.total_revenue) }}
          </p>
          <p class="text-xs text-muted-foreground mt-1">
            {{ sales.total_invoices }} factura{{
              sales.total_invoices !== 1 ? "s" : ""
            }}
          </p>
        </div>

        <div class="bg-card border border-border rounded-xl p-5">
          <div class="flex items-center justify-between mb-3">
            <p class="text-sm text-muted-foreground">Compras</p>
            <div class="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
              <TrendingDown class="w-4 h-4 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
          <p class="text-2xl font-bold text-foreground font-mono">
            {{ fmtCurrency(purchases.total_spend) }}
          </p>
          <p class="text-xs text-muted-foreground mt-1">
            {{ purchases.received_orders }}/{{ purchases.total_orders }} órdenes
            recibidas
          </p>
        </div>

        <div class="bg-card border border-border rounded-xl p-5">
          <div class="flex items-center justify-between mb-3">
            <p class="text-sm text-muted-foreground">Productos</p>
            <div class="p-2 bg-primary/10 rounded-lg">
              <Boxes class="w-4 h-4 text-primary" />
            </div>
          </div>
          <p class="text-2xl font-bold text-foreground">
            {{ inventory.total_products }}
          </p>
          <p class="text-xs text-muted-foreground mt-1">
            {{ inventory.out_of_stock_count }} sin stock ·
            {{ inventory.low_stock_count }} bajo
          </p>
        </div>

        <div class="bg-card border border-border rounded-xl p-5">
          <div class="flex items-center justify-between mb-3">
            <p class="text-sm text-muted-foreground">Facturas internas</p>
            <div class="p-2 bg-orange-100 dark:bg-orange-900/30 rounded-lg">
              <BarChart2 class="w-4 h-4 text-orange-500" />
            </div>
          </div>
          <p class="text-2xl font-bold text-foreground">
            {{ sales.internal_count }}
          </p>
          <p class="text-xs text-muted-foreground mt-1">
            {{ sales.fiscal_count }} fiscal{{
              sales.fiscal_count !== 1 ? "es" : ""
            }}
          </p>
        </div>
      </div>

      <!-- Tabs -->
      <div class="flex gap-1 bg-muted/50 rounded-lg p-1 w-fit">
        <button
          v-for="tab in [
            { key: 'sales', label: 'Ventas', icon: TrendingUp },
            { key: 'inventory', label: 'Inventario', icon: Package },
            { key: 'purchases', label: 'Compras', icon: ShoppingCart },
          ]"
          :key="tab.key"
          :class="[
            'flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors',
            activeTab === tab.key
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground',
          ]"
          @click="activeTab = tab.key as 'sales' | 'inventory' | 'purchases'"
        >
          <component :is="tab.icon" class="w-4 h-4" />
          {{ tab.label }}
        </button>
      </div>

      <!-- ── Sales Tab ──────────────────────────────────────────────────────── -->
      <div v-if="activeTab === 'sales'" class="space-y-5">
        <!-- Revenue chart -->
        <div class="bg-card border border-border rounded-xl p-6">
          <h3 class="font-semibold text-foreground mb-4">Ingresos por día</h3>
          <div
            v-if="sales.by_day.length === 0"
            class="text-center py-10 text-muted-foreground text-sm"
          >
            Sin ventas en el período seleccionado
          </div>
          <div v-else class="flex items-end gap-1.5 h-40 overflow-x-auto pb-2">
            <div
              v-for="day in sales.by_day"
              :key="day.date"
              class="flex flex-col items-center gap-1 min-w-6 flex-1"
              :title="`${fmtDate(day.date)}: ${fmtCurrency(day.revenue)}`"
            >
              <div
                class="w-full bg-primary/80 hover:bg-primary rounded-t transition-colors cursor-default"
                :style="`height: ${barHeight(day.revenue)}%`"
              />
              <span
                class="text-[9px] text-muted-foreground rotate-45 origin-left whitespace-nowrap"
              >
                {{
                  new Date(day.date).toLocaleDateString("es-VE", {
                    day: "2-digit",
                    month: "short",
                  })
                }}
              </span>
            </div>
          </div>
        </div>

        <!-- Top products -->
        <div class="bg-card border border-border rounded-xl overflow-hidden">
          <div class="px-6 py-4 border-b border-border">
            <h3 class="font-semibold text-foreground">
              Top productos por ventas
            </h3>
          </div>
          <div
            v-if="sales.top_products.length === 0"
            class="text-center py-10 text-muted-foreground text-sm"
          >
            Sin datos de productos en el período
          </div>
          <table v-else class="w-full text-sm">
            <thead>
              <tr class="border-b border-border bg-muted/30">
                <th
                  class="px-4 py-3 text-left font-medium text-muted-foreground"
                >
                  #
                </th>
                <th
                  class="px-4 py-3 text-left font-medium text-muted-foreground"
                >
                  Producto
                </th>
                <th
                  class="px-4 py-3 text-right font-medium text-muted-foreground hidden md:table-cell"
                >
                  Unidades
                </th>
                <th
                  class="px-4 py-3 text-right font-medium text-muted-foreground"
                >
                  Ingresos
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <tr
                v-for="(prod, idx) in sales.top_products"
                :key="prod.product_id"
                class="hover:bg-muted/20 transition-colors"
              >
                <td class="px-4 py-3 text-muted-foreground font-mono text-xs">
                  {{ idx + 1 }}
                </td>
                <td class="px-4 py-3 font-medium text-foreground">
                  {{ prod.product_name }}
                </td>
                <td
                  class="px-4 py-3 text-right text-muted-foreground hidden md:table-cell font-mono"
                >
                  {{ prod.total_sold }}
                </td>
                <td
                  class="px-4 py-3 text-right font-mono font-semibold text-foreground"
                >
                  {{ fmtCurrency(prod.revenue) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- ── Inventory Tab ──────────────────────────────────────────────────── -->
      <div v-if="activeTab === 'inventory'" class="space-y-4">
        <!-- Alert badges -->
        <div
          v-if="
            inventory.out_of_stock_count > 0 || inventory.low_stock_count > 0
          "
          class="flex gap-3 flex-wrap"
        >
          <div
            v-if="inventory.out_of_stock_count > 0"
            class="flex items-center gap-2 px-4 py-2.5 bg-rose-50 dark:bg-rose-900/20 border border-rose-200 dark:border-rose-800 rounded-xl text-sm text-rose-700 dark:text-rose-400"
          >
            <AlertTriangle class="w-4 h-4" />
            {{ inventory.out_of_stock_count }} producto{{
              inventory.out_of_stock_count !== 1 ? "s" : ""
            }}
            sin stock
          </div>
          <div
            v-if="inventory.low_stock_count > 0"
            class="flex items-center gap-2 px-4 py-2.5 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-xl text-sm text-amber-700 dark:text-amber-400"
          >
            <AlertTriangle class="w-4 h-4" />
            {{ inventory.low_stock_count }} con bajo stock
          </div>
        </div>

        <div class="bg-card border border-border rounded-xl overflow-hidden">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border bg-muted/30">
                <th
                  class="px-4 py-3 text-left font-medium text-muted-foreground"
                >
                  Producto
                </th>
                <th
                  class="px-4 py-3 text-left font-medium text-muted-foreground hidden md:table-cell"
                >
                  Categoría
                </th>
                <th
                  class="px-4 py-3 text-left font-medium text-muted-foreground hidden lg:table-cell"
                >
                  Tipo
                </th>
                <th
                  class="px-4 py-3 text-right font-medium text-muted-foreground"
                >
                  Stock
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <tr
                v-for="item in inventory.items"
                :key="item.product_id"
                class="hover:bg-muted/20 transition-colors"
                :class="!item.active ? 'opacity-50' : ''"
              >
                <td class="px-4 py-3">
                  <p class="font-medium text-foreground">
                    {{ item.product_name }}
                  </p>
                  <p
                    v-if="item.sku"
                    class="text-xs text-muted-foreground font-mono"
                  >
                    {{ item.sku }}
                  </p>
                </td>
                <td
                  class="px-4 py-3 text-muted-foreground hidden md:table-cell"
                >
                  {{ item.category_name || "—" }}
                </td>
                <td class="px-4 py-3 hidden lg:table-cell">
                  <span
                    :class="[
                      'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                      item.is_fiscal
                        ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                        : 'bg-muted text-muted-foreground',
                    ]"
                  >
                    {{ item.is_fiscal ? "Fiscal" : "Interno" }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right">
                  <span
                    :class="[
                      'font-mono font-semibold',
                      item.quantity_on_hand === 0
                        ? 'text-rose-600 dark:text-rose-400'
                        : item.quantity_on_hand < 5
                          ? 'text-amber-600 dark:text-amber-400'
                          : 'text-foreground',
                    ]"
                  >
                    {{ item.quantity_on_hand }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- ── Purchases Tab ──────────────────────────────────────────────────── -->
      <div v-if="activeTab === 'purchases'" class="space-y-4">
        <!-- Summary bar -->
        <div class="grid grid-cols-3 gap-4">
          <div class="bg-card border border-border rounded-xl p-4 text-center">
            <p class="text-sm text-muted-foreground">Total gastado</p>
            <p class="text-xl font-bold text-foreground font-mono mt-1">
              {{ fmtCurrency(purchases.total_spend) }}
            </p>
          </div>
          <div class="bg-card border border-border rounded-xl p-4 text-center">
            <p class="text-sm text-muted-foreground">Órdenes</p>
            <p class="text-xl font-bold text-foreground mt-1">
              {{ purchases.total_orders }}
            </p>
          </div>
          <div class="bg-card border border-border rounded-xl p-4 text-center">
            <p class="text-sm text-muted-foreground">Recibidas</p>
            <p class="text-xl font-bold text-emerald-600 mt-1">
              {{ purchases.received_orders }}
            </p>
          </div>
        </div>

        <!-- By supplier -->
        <div class="bg-card border border-border rounded-xl overflow-hidden">
          <div class="px-6 py-4 border-b border-border">
            <h3 class="font-semibold text-foreground">Gasto por proveedor</h3>
          </div>
          <div
            v-if="purchases.by_supplier.length === 0"
            class="text-center py-10 text-muted-foreground text-sm"
          >
            Sin compras en el período seleccionado
          </div>
          <table v-else class="w-full text-sm">
            <thead>
              <tr class="border-b border-border bg-muted/30">
                <th
                  class="px-4 py-3 text-left font-medium text-muted-foreground"
                >
                  Proveedor
                </th>
                <th
                  class="px-4 py-3 text-right font-medium text-muted-foreground hidden md:table-cell"
                >
                  Órdenes
                </th>
                <th
                  class="px-4 py-3 text-right font-medium text-muted-foreground"
                >
                  Monto
                </th>
                <th
                  class="px-4 py-3 text-right font-medium text-muted-foreground hidden lg:table-cell"
                >
                  % del total
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <tr
                v-for="sup in purchases.by_supplier"
                :key="sup.supplier_id"
                class="hover:bg-muted/20 transition-colors"
              >
                <td class="px-4 py-3 font-medium text-foreground">
                  {{ sup.supplier_name }}
                </td>
                <td
                  class="px-4 py-3 text-right text-muted-foreground hidden md:table-cell"
                >
                  {{ sup.order_count }}
                </td>
                <td
                  class="px-4 py-3 text-right font-mono font-semibold text-foreground"
                >
                  {{ fmtCurrency(sup.total_spend) }}
                </td>
                <td
                  class="px-4 py-3 text-right text-muted-foreground hidden lg:table-cell"
                >
                  <div class="flex items-center justify-end gap-2">
                    <div
                      class="w-16 h-1.5 bg-muted rounded-full overflow-hidden"
                    >
                      <div
                        class="h-full bg-blue-500 rounded-full"
                        :style="`width: ${purchases.total_spend > 0 ? Math.round((sup.total_spend / purchases.total_spend) * 100) : 0}%`"
                      />
                    </div>
                    <span class="text-xs w-8 text-right">
                      {{
                        purchases.total_spend > 0
                          ? Math.round(
                              (sup.total_spend / purchases.total_spend) * 100,
                            )
                          : 0
                      }}%
                    </span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>
