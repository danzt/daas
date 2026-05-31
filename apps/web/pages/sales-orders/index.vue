<script setup lang="ts">
import {
  ShoppingBag,
  Plus,
  Search,
  X,
  Save,
  Trash2,
  Loader2,
  FolderOpen,
  ChevronRight,
} from "lucide-vue-next";
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
const formOpen = ref(false);
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
    if (!line.unit_price || line.unit_price === 0) {
      line.unit_price = product.internal_price ?? product.fiscal_price ?? 0;
    }
  }
}

const formTotal = computed(() =>
  form.lines.reduce((acc, l) => acc + l.quantity * l.unit_price, 0),
);

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

// ─── Status config ────────────────────────────────────────────────────────────
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

function openForm() {
  form.customer_name = "";
  form.customer_id_type = "anonymous";
  form.customer_id_number = "";
  form.notes = "";
  form.lines = [];
  addLine();
  saveError.value = "";
  formOpen.value = true;
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
    formOpen.value = false;
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
  <div class="p-6 space-y-6">
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
      <button
        class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors"
        @click="openForm"
      >
        <Plus class="w-4 h-4" />
        Nueva orden
      </button>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-3 gap-4">
      <div class="bg-card border border-border rounded-xl p-4">
        <p class="text-sm text-muted-foreground">Borradores</p>
        <p class="text-2xl font-bold text-foreground mt-1">
          {{ counts.draft }}
        </p>
      </div>
      <div class="bg-card border border-border rounded-xl p-4">
        <p class="text-sm text-muted-foreground">Confirmadas</p>
        <p class="text-2xl font-bold text-blue-600 mt-1">
          {{ counts.confirmed }}
        </p>
      </div>
      <div class="bg-card border border-border rounded-xl p-4">
        <p class="text-sm text-muted-foreground">Facturadas</p>
        <p class="text-2xl font-bold text-emerald-600 mt-1">
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
          class="w-full h-10 pl-9 pr-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
          placeholder="Buscar por cliente..."
        />
      </div>
      <div class="flex gap-1 bg-muted/50 rounded-lg p-1">
        <button
          v-for="tab in filterTabs"
          :key="tab.key"
          :class="[
            'px-3 py-1.5 rounded-md text-sm font-medium transition-colors',
            statusFilter === tab.key
              ? 'bg-background text-foreground shadow-sm'
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
      class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3"
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
        @click="openForm"
      >
        Crear la primera orden
      </button>
    </div>

    <!-- Table -->
    <div
      v-else-if="!loading"
      class="bg-card border border-border rounded-xl overflow-hidden"
    >
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border bg-muted/30">
            <th class="px-4 py-3 text-left font-medium text-muted-foreground">
              Cliente
            </th>
            <th class="px-4 py-3 text-left font-medium text-muted-foreground">
              Estado
            </th>
            <th
              class="px-4 py-3 text-left font-medium text-muted-foreground hidden md:table-cell"
            >
              Fecha
            </th>
            <th class="px-4 py-3 text-right font-medium text-muted-foreground">
              Total
            </th>
            <th class="px-4 py-3" />
          </tr>
        </thead>
        <tbody class="divide-y divide-border">
          <tr
            v-for="order in filteredOrders"
            :key="order.id"
            class="hover:bg-muted/20 transition-colors group"
          >
            <td class="px-4 py-3">
              <p class="font-medium text-foreground">
                {{ order.customer_name || "Cliente anónimo" }}
              </p>
              <p
                v-if="order.notes"
                class="text-xs text-muted-foreground truncate max-w-48"
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
              class="px-4 py-3 text-right font-mono font-medium text-foreground"
            >
              {{ fmtCurrency(order.total) }}
            </td>
            <td class="px-4 py-3 text-right">
              <NuxtLink
                :to="`/sales-orders/${order.id}`"
                class="opacity-0 group-hover:opacity-100 p-1.5 rounded-lg hover:bg-muted transition-all text-muted-foreground inline-flex"
              >
                <ChevronRight class="w-4 h-4" />
              </NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- New order slide-over -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="formOpen"
          class="fixed inset-0 z-50 flex items-start justify-end bg-black/40"
          @click.self="formOpen = false"
        >
          <div
            class="bg-card border-l border-border h-full w-full max-w-xl flex flex-col shadow-2xl overflow-hidden"
          >
            <!-- Header -->
            <div
              class="flex items-center justify-between px-6 py-4 border-b border-border shrink-0"
            >
              <div class="flex items-center gap-3">
                <div class="p-2 bg-primary/10 rounded-lg">
                  <ShoppingBag class="w-5 h-5 text-primary" />
                </div>
                <h2 class="text-lg font-semibold text-foreground">
                  Nueva orden de venta
                </h2>
              </div>
              <button
                class="p-2 rounded-lg hover:bg-muted transition-colors text-muted-foreground"
                @click="formOpen = false"
              >
                <X class="w-4 h-4" />
              </button>
            </div>

            <!-- Body -->
            <div class="flex-1 overflow-y-auto px-6 py-5 space-y-5">
              <div
                v-if="saveError"
                class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3"
              >
                {{ saveError }}
              </div>

              <!-- Customer -->
              <div class="space-y-3">
                <p class="text-sm font-medium text-foreground">Cliente</p>
                <div class="grid grid-cols-2 gap-3">
                  <div class="space-y-1.5">
                    <label class="text-xs font-medium text-muted-foreground"
                      >Nombre</label
                    >
                    <input
                      v-model="form.customer_name"
                      class="w-full h-9 px-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
                      placeholder="Nombre del cliente..."
                    />
                  </div>
                  <div class="space-y-1.5">
                    <label class="text-xs font-medium text-muted-foreground"
                      >Tipo de doc.</label
                    >
                    <select
                      v-model="form.customer_id_type"
                      class="w-full h-9 px-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
                    >
                      <option value="anonymous">Anónimo</option>
                      <option value="cedula">Cédula</option>
                      <option value="rif">RIF</option>
                      <option value="passport">Pasaporte</option>
                    </select>
                  </div>
                </div>
                <div
                  v-if="form.customer_id_type !== 'anonymous'"
                  class="space-y-1.5"
                >
                  <label class="text-xs font-medium text-muted-foreground"
                    >Número de documento</label
                  >
                  <input
                    v-model="form.customer_id_number"
                    class="w-full h-9 px-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
                    placeholder="Número..."
                  />
                </div>
              </div>

              <!-- Notes -->
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground">Notas</label>
                <textarea
                  v-model="form.notes"
                  rows="2"
                  class="w-full px-3 py-2 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/40 resize-none"
                  placeholder="Observaciones..."
                />
              </div>

              <!-- Lines -->
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <label class="text-sm font-medium text-foreground">
                    Líneas <span class="text-destructive">*</span>
                  </label>
                  <button
                    class="text-xs text-primary hover:underline flex items-center gap-1"
                    @click="addLine"
                  >
                    <Plus class="w-3 h-3" />
                    Agregar línea
                  </button>
                </div>

                <div
                  v-for="(line, idx) in form.lines"
                  :key="idx"
                  class="bg-muted/30 border border-border rounded-lg p-4 space-y-3"
                >
                  <div class="flex items-center justify-between">
                    <span class="text-xs font-medium text-muted-foreground"
                      >Línea {{ idx + 1 }}</span
                    >
                    <button
                      class="p-1 rounded hover:bg-muted text-muted-foreground transition-colors"
                      @click="removeLine(idx)"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>

                  <div class="space-y-1.5">
                    <label class="text-xs font-medium text-muted-foreground"
                      >Producto</label
                    >
                    <select
                      v-model="line.product_id"
                      class="w-full h-9 px-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
                      @change="onProductChange(line)"
                    >
                      <option value="">Seleccionar producto...</option>
                      <option
                        v-for="prod in products"
                        :key="prod.id"
                        :value="prod.id"
                      >
                        {{ prod.name }}
                      </option>
                    </select>
                  </div>

                  <div class="space-y-1.5">
                    <label class="text-xs font-medium text-muted-foreground"
                      >Descripción</label
                    >
                    <input
                      v-model="line.description"
                      class="w-full h-9 px-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
                      placeholder="Descripción en la orden..."
                    />
                  </div>

                  <div class="grid grid-cols-2 gap-3">
                    <div class="space-y-1.5">
                      <label class="text-xs font-medium text-muted-foreground"
                        >Cantidad</label
                      >
                      <input
                        v-model.number="line.quantity"
                        type="number"
                        min="0.001"
                        step="any"
                        class="w-full h-9 px-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
                      />
                    </div>
                    <div class="space-y-1.5">
                      <label class="text-xs font-medium text-muted-foreground"
                        >Precio unitario</label
                      >
                      <input
                        v-model.number="line.unit_price"
                        type="number"
                        min="0"
                        step="any"
                        class="w-full h-9 px-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
                      />
                    </div>
                  </div>

                  <div class="text-right text-xs text-muted-foreground">
                    Subtotal:
                    <span class="font-mono font-medium text-foreground">
                      {{ fmtCurrency(line.quantity * line.unit_price) }}
                    </span>
                  </div>
                </div>

                <!-- Total -->
                <div
                  v-if="form.lines.length > 0"
                  class="flex items-center justify-between px-4 py-3 bg-primary/5 rounded-lg border border-primary/20"
                >
                  <span class="text-sm font-medium text-foreground">Total</span>
                  <span class="font-mono font-bold text-primary text-lg">
                    {{ fmtCurrency(formTotal) }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Footer -->
            <div
              class="flex items-center justify-end gap-3 px-6 py-4 border-t border-border shrink-0"
            >
              <button
                class="px-4 py-2 text-sm rounded-lg border border-border hover:bg-muted transition-colors text-foreground"
                @click="formOpen = false"
              >
                Cancelar
              </button>
              <button
                :disabled="saving"
                class="flex items-center gap-2 px-4 py-2 text-sm rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50"
                @click="createOrder"
              >
                <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                <Save v-else class="w-4 h-4" />
                {{ saving ? "Creando..." : "Crear orden" }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
