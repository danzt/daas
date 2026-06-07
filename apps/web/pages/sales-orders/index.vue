<script setup lang="ts">
import {
  ShoppingBag,
  Plus,
  Search,
  X,
  Trash2,
  Loader2,
  FolderOpen,
  ChevronRight,
  Save,
} from "lucide-vue-next";
import { Sheet, SheetHeader, SheetFooter } from "~/components/ui/sheet";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { useApiFetch } from "~/composables/useAuth";
import type { Product } from "~/components/products/ProductFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

// ─── Types ────────────────────────────────────────────────────────────────────

interface OrderLine {
  id: string;
  product_id: string;
  description: string;
  quantity: number;
  unit_price: number;
  subtotal: number;
}

export interface SaleOrder {
  id: string;
  status: "draft" | "confirmed" | "invoiced" | "cancelled";
  customer_name: string;
  customer_id_type: string;
  customer_id_number: string;
  notes: string;
  total: number;
  confirmed_at: string | null;
  invoiced_at: string | null;
  invoice_id: string | null;
  created_at: string;
  lines: OrderLine[];
}

// ─── State ───────────────────────────────────────────────────────────────────
const orders = ref<SaleOrder[]>([]);
const products = ref<Product[]>([]);
const loading = ref(false);
const loadError = ref("");
const searchQuery = ref("");
const statusFilter = ref("all");
const sheetOpen = ref(false);
const saving = ref(false);
const saveError = ref("");

// ─── Form state ──────────────────────────────────────────────────────────────
interface FormLine {
  product_id: string;
  description: string;
  quantity: number;
  unit_price: number;
}

const form = reactive({
  customer_name: "",
  customer_id_type: "anonymous",
  customer_id_number: "",
  notes: "",
  lines: [] as FormLine[],
});

function addLine() {
  form.lines.push({
    product_id: "",
    description: "",
    quantity: 1,
    unit_price: 0,
  });
}

function removeLine(idx: number) {
  form.lines.splice(idx, 1);
}

function onProductChange(line: FormLine) {
  const product = products.value.find((p) => p.id === line.product_id);
  if (product) {
    if (!line.description) line.description = product.name;
    if (!line.unit_price) {
      line.unit_price = product.internal_price ?? product.fiscal_price ?? 0;
    }
  }
}

const formTotal = computed(() =>
  form.lines.reduce((acc, l) => acc + l.quantity * l.unit_price, 0),
);

// ─── Status config ────────────────────────────────────────────────────────────
const STATUS_CONFIG: Record<string, { label: string; class: string }> = {
  draft: { label: "Borrador", class: "bg-gray-100 text-muted-foreground" },
  confirmed: { label: "Confirmada", class: "bg-blue-100 text-blue-700" },
  invoiced: { label: "Facturada", class: "bg-emerald-100 text-emerald-700" },
  cancelled: { label: "Cancelada", class: "bg-red-100 text-red-600" },
};

const filterTabs = [
  { key: "all", label: "Todas" },
  { key: "draft", label: "Borradores" },
  { key: "confirmed", label: "Confirmadas" },
  { key: "invoiced", label: "Facturadas" },
  { key: "cancelled", label: "Canceladas" },
];

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
        o.id.toLowerCase().includes(q),
    );
  }
  return list;
});

const counts = computed(() => ({
  draft: orders.value.filter((o) => o.status === "draft").length,
  confirmed: orders.value.filter((o) => o.status === "confirmed").length,
  invoiced: orders.value.filter((o) => o.status === "invoiced").length,
}));

// ─── Formatters ───────────────────────────────────────────────────────────────
function fmtCurrency(n: number) {
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
}

function fmtDate(d: string | null) {
  if (!d) return "—";
  return new Date(d).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

// ─── API ─────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    const [ord, prods] = await Promise.all([
      useApiFetch<SaleOrder[]>("/api/v1/sales-orders"),
      useApiFetch<Product[]>("/api/v1/products"),
    ]);
    orders.value = ord;
    products.value = prods.filter((p) => p.active);
  } catch {
    loadError.value = "No se pudieron cargar las órdenes de venta";
  } finally {
    loading.value = false;
  }
}

function openSheet() {
  form.customer_name = "";
  form.customer_id_type = "anonymous";
  form.customer_id_number = "";
  form.notes = "";
  form.lines = [];
  addLine();
  saveError.value = "";
  sheetOpen.value = true;
}

async function createOrder() {
  if (form.lines.length === 0) {
    saveError.value = "Agregá al menos una línea";
    return;
  }
  for (const l of form.lines) {
    if (!l.product_id) {
      saveError.value = "Seleccioná un producto en cada línea";
      return;
    }
    if (l.quantity <= 0 || l.unit_price <= 0) {
      saveError.value = "Cantidad y precio deben ser mayores a 0";
      return;
    }
  }

  saving.value = true;
  saveError.value = "";
  try {
    const order = await useApiFetch<SaleOrder>("/api/v1/sales-orders", {
      method: "POST",
      body: {
        customer_name: form.customer_name,
        customer_id_type: form.customer_id_type,
        customer_id_number: form.customer_id_number,
        notes: form.notes,
        lines: form.lines.map((l) => ({
          product_id: l.product_id,
          description: l.description,
          quantity: l.quantity,
          unit_price: l.unit_price,
        })),
      },
    });
    orders.value.unshift(order);
    sheetOpen.value = false;
  } catch (err: unknown) {
    const e = err as { data?: { detail?: string } };
    saveError.value = e?.data?.detail ?? "Error al crear la orden";
  } finally {
    saving.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-foreground">Órdenes de venta</h1>
        <p class="text-sm text-muted-foreground mt-0.5">
          {{ counts.confirmed }} confirmada{{
            counts.confirmed !== 1 ? "s" : ""
          }}
          · {{ counts.draft }} borrador{{ counts.draft !== 1 ? "es" : "" }}
        </p>
      </div>
      <NuxtLink
        to="/sales-orders/new"
        class="inline-flex items-center gap-2 h-9 px-4 bg-primary text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200"
      >
        <Plus class="w-4 h-4" />
        Nueva orden
      </NuxtLink>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-3 gap-4">
      <div class="border bg-card rounded-xl p-4 shadow-sm">
        <p
          class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
        >
          Borradores
        </p>
        <p class="text-3xl font-bold text-foreground mt-1">
          {{ counts.draft }}
        </p>
      </div>
      <div class="border bg-card rounded-xl p-4 shadow-sm">
        <p
          class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
        >
          Confirmadas
        </p>
        <p class="text-3xl font-bold text-foreground mt-1">
          {{ counts.confirmed }}
        </p>
      </div>
      <div class="border bg-card rounded-xl p-4 shadow-sm">
        <p
          class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
        >
          Facturadas
        </p>
        <p class="text-3xl font-bold text-foreground mt-1">
          {{ counts.invoiced }}
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
          class="w-full h-10 pl-9 pr-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
          placeholder="Buscar por cliente..."
        />
      </div>
      <div class="flex gap-1 bg-gray-100 rounded-lg p-1">
        <button
          v-for="tab in filterTabs"
          :key="tab.key"
          :class="[
            'px-3 py-1.5 rounded-md text-sm font-medium transition-colors',
            statusFilter === tab.key
              ? 'bg-white text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground',
          ]"
          @click="statusFilter = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <!-- Error -->
    <div
      v-if="loadError"
      class="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3"
    >
      {{ loadError }}
    </div>

    <!-- Loading -->
    <div
      v-if="loading"
      class="flex items-center justify-center py-20 text-muted-foreground"
    >
      <Loader2 class="w-6 h-6 animate-spin mr-2" />
      Cargando órdenes...
    </div>

    <!-- Empty -->
    <div
      v-else-if="!loading && filteredOrders.length === 0"
      class="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3"
    >
      <FolderOpen class="w-12 h-12 opacity-30" />
      <p class="text-sm">
        {{
          searchQuery || statusFilter !== "all"
            ? "No hay órdenes que coincidan con el filtro"
            : "Aún no hay órdenes de venta"
        }}
      </p>
      <button
        v-if="!searchQuery && statusFilter === 'all'"
        class="text-sm text-primary hover:underline"
        @click="openSheet"
      >
        Crear la primera orden
      </button>
    </div>

    <!-- ── Mobile card list (xs/sm) ─────────────────────── -->
    <div
      v-else-if="!loading"
      class="sm:hidden border bg-card rounded-xl overflow-hidden shadow-sm divide-y divide-border"
    >
      <NuxtLink
        v-for="order in filteredOrders"
        :key="order.id"
        :to="`/sales-orders/${order.id}`"
        class="flex items-stretch gap-0 active:bg-gray-50 transition-colors duration-100"
      >
        <!-- Left status bar -->
        <div
          class="w-1 shrink-0 my-1.5 rounded-r"
          :class="{
            'bg-gray-400': order.status === 'draft',
            'bg-blue-500': order.status === 'confirmed',
            'bg-primary': order.status === 'invoiced',
            'bg-red-400': order.status === 'cancelled',
          }"
        />

        <div class="flex flex-1 items-center gap-3 px-4 py-3.5 min-w-0">
          <!-- Status icon -->
          <div
            class="no-min-tap shrink-0 w-11 h-11 rounded-2xl flex items-center justify-center"
            :class="{
              'bg-gray-100': order.status === 'draft',
              'bg-blue-50': order.status === 'confirmed',
              'bg-primary/10': order.status === 'invoiced',
              'bg-red-50': order.status === 'cancelled',
            }"
          >
            <ShoppingBag
              class="w-5 h-5"
              :class="{
                'text-muted-foreground': order.status === 'draft',
                'text-blue-600': order.status === 'confirmed',
                'text-primary': order.status === 'invoiced',
                'text-red-500': order.status === 'cancelled',
              }"
            />
          </div>

          <!-- Order info -->
          <div class="flex-1 min-w-0">
            <p
              class="font-semibold text-sm text-foreground truncate leading-tight"
            >
              {{ order.customer_name || "Cliente anónimo" }}
            </p>
            <div class="flex items-center gap-1.5 mt-0.5">
              <span
                :class="[
                  'inline-flex items-center px-1.5 py-px rounded text-[10px] font-bold',
                  STATUS_CONFIG[order.status]?.class,
                ]"
              >
                {{ STATUS_CONFIG[order.status]?.label }}
              </span>
              <span class="text-[11px] text-muted-foreground">
                {{ fmtDate(order.created_at) }}
              </span>
            </div>
            <p
              v-if="order.notes"
              class="text-[11px] text-muted-foreground truncate mt-0.5"
            >
              {{ order.notes }}
            </p>
          </div>

          <!-- Total + chevron -->
          <div class="flex items-center gap-1 shrink-0">
            <span class="font-mono font-bold text-sm text-foreground">
              {{ fmtCurrency(order.total) }}
            </span>
            <ChevronRight class="w-4 h-4 text-muted-foreground/50" />
          </div>
        </div>
      </NuxtLink>
    </div>

    <!-- ── Desktop table (sm+) ────────────────────────── -->
    <div
      v-else-if="!loading"
      class="hidden sm:block border bg-card rounded-xl overflow-hidden shadow-sm"
    >
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border bg-gray-50/50">
              <th
                class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground"
              >
                Cliente
              </th>
              <th
                class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground"
              >
                Estado
              </th>
              <th
                class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground hidden md:table-cell"
              >
                Fecha
              </th>
              <th
                class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted-foreground"
              >
                Total
              </th>
              <th class="px-4 py-3" />
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr
              v-for="order in filteredOrders"
              :key="order.id"
              class="hover:bg-gray-50 transition-colors group"
            >
              <td class="px-4 py-3">
                <p class="font-medium text-foreground">
                  {{ order.customer_name || "Cliente anónimo" }}
                </p>
                <p
                  v-if="order.notes"
                  class="text-xs text-muted-foreground truncate max-w-48 mt-0.5"
                >
                  {{ order.notes }}
                </p>
              </td>
              <td class="px-4 py-3">
                <span
                  :class="[
                    'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                    STATUS_CONFIG[order.status]?.class,
                  ]"
                >
                  {{ STATUS_CONFIG[order.status]?.label }}
                </span>
              </td>
              <td
                class="px-4 py-3 text-muted-foreground hidden md:table-cell text-xs"
              >
                {{ fmtDate(order.created_at) }}
              </td>
              <td
                class="px-4 py-3 text-right font-mono font-semibold text-foreground"
              >
                {{ fmtCurrency(order.total) }}
              </td>
              <td class="px-4 py-3 text-right">
                <NuxtLink
                  :to="`/sales-orders/${order.id}`"
                  class="opacity-0 group-hover:opacity-100 inline-flex p-1.5 rounded-lg hover:bg-gray-100 transition-all text-muted-foreground"
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

  <!-- ─── Create Sheet ──────────────────────────────────────────────────────── -->
  <Sheet v-model:open="sheetOpen">
    <!-- Header -->
    <SheetHeader>
      <div class="flex items-center gap-3">
        <div class="p-2 bg-primary/10 rounded-lg">
          <ShoppingBag class="w-5 h-5 text-primary" />
        </div>
        <div>
          <h2 class="text-base font-semibold text-foreground">
            Nueva orden de venta
          </h2>
          <p class="text-xs text-muted-foreground mt-0.5">
            Completá los datos del cliente y las líneas
          </p>
        </div>
      </div>
      <button
        class="p-1.5 rounded-lg hover:bg-gray-100 transition-colors text-muted-foreground"
        @click="sheetOpen = false"
      >
        <X class="w-4 h-4" />
      </button>
    </SheetHeader>

    <!-- Body -->
    <div class="flex-1 overflow-y-auto px-6 py-5 space-y-6">
      <!-- Error -->
      <div
        v-if="saveError"
        class="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3"
      >
        {{ saveError }}
      </div>

      <!-- Customer section -->
      <section class="space-y-4">
        <h3
          class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
        >
          Datos del cliente
        </h3>

        <div class="grid grid-cols-2 gap-3">
          <div class="space-y-1.5">
            <Label for="customer_name">Nombre</Label>
            <Input
              id="customer_name"
              v-model="form.customer_name"
              placeholder="Nombre del cliente..."
            />
          </div>
          <div class="space-y-1.5">
            <Label for="customer_id_type">Tipo de doc.</Label>
            <select
              id="customer_id_type"
              v-model="form.customer_id_type"
              class="w-full h-11 px-4 border border-input rounded-lg text-base bg-white focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
            >
              <option value="anonymous">Anónimo</option>
              <option value="cedula">Cédula</option>
              <option value="rif">RIF</option>
              <option value="passport">Pasaporte</option>
            </select>
          </div>
        </div>

        <div v-if="form.customer_id_type !== 'anonymous'" class="space-y-1.5">
          <Label for="customer_id_number">Número de documento</Label>
          <Input
            id="customer_id_number"
            v-model="form.customer_id_number"
            placeholder="Ej. V-12345678"
          />
        </div>

        <div class="space-y-1.5">
          <Label for="notes">Notas</Label>
          <textarea
            id="notes"
            v-model="form.notes"
            rows="2"
            class="w-full px-4 py-2.5 border border-input rounded-lg text-sm bg-white focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 resize-none placeholder:text-muted-foreground"
            placeholder="Observaciones opcionales..."
          />
        </div>
      </section>

      <!-- Lines section -->
      <section class="space-y-3">
        <div class="flex items-center justify-between">
          <h3
            class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
          >
            Líneas
            <span class="text-red-500 normal-case font-normal tracking-normal"
              >*</span
            >
          </h3>
          <button
            class="text-xs font-medium text-primary hover:underline flex items-center gap-1"
            @click="addLine"
          >
            <Plus class="w-3 h-3" />
            Agregar línea
          </button>
        </div>

        <div
          v-for="(line, idx) in form.lines"
          :key="idx"
          class="border border-border rounded-xl p-4 space-y-3 bg-gray-50/50"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold text-muted-foreground"
              >Línea {{ idx + 1 }}</span
            >
            <button
              class="p-1 rounded-md hover:bg-gray-200 text-muted-foreground transition-colors"
              @click="removeLine(idx)"
            >
              <Trash2 class="w-3.5 h-3.5" />
            </button>
          </div>

          <!-- Product select -->
          <div class="space-y-1.5">
            <Label>Producto</Label>
            <select
              v-model="line.product_id"
              class="w-full h-11 px-4 border border-input rounded-lg text-sm bg-white focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
              @change="onProductChange(line)"
            >
              <option value="">Seleccionar producto...</option>
              <option v-for="prod in products" :key="prod.id" :value="prod.id">
                {{ prod.name }}
              </option>
            </select>
          </div>

          <!-- Description -->
          <div class="space-y-1.5">
            <Label>Descripción</Label>
            <Input
              v-model="line.description"
              placeholder="Descripción en la orden..."
            />
          </div>

          <!-- Qty + Price -->
          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1.5">
              <Label>Cantidad</Label>
              <Input
                v-model="line.quantity"
                type="number"
                min="0.001"
                step="any"
              />
            </div>
            <div class="space-y-1.5">
              <Label>Precio unitario</Label>
              <Input
                v-model="line.unit_price"
                type="number"
                min="0"
                step="any"
              />
            </div>
          </div>

          <div class="flex justify-end">
            <span class="text-xs text-muted-foreground">
              Subtotal:
              <span class="font-mono font-semibold text-foreground ml-1">
                {{ fmtCurrency(line.quantity * line.unit_price) }}
              </span>
            </span>
          </div>
        </div>

        <!-- Total -->
        <div
          v-if="form.lines.length > 0"
          class="flex items-center justify-between px-4 py-3 bg-primary/5 border border-primary/20 rounded-xl"
        >
          <span class="text-sm font-semibold text-foreground">Total</span>
          <span class="font-mono font-bold text-primary text-xl">
            {{ fmtCurrency(formTotal) }}
          </span>
        </div>
      </section>
    </div>

    <!-- Footer -->
    <SheetFooter>
      <Button variant="outline" @click="sheetOpen = false">Cancelar</Button>
      <Button :disabled="saving" @click="createOrder">
        <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
        <Save v-else class="w-4 h-4" />
        {{ saving ? "Creando..." : "Crear orden" }}
      </Button>
    </SheetFooter>
  </Sheet>
</template>
