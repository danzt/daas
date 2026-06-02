<script setup lang="ts">
import {
  ArrowLeft,
  ShoppingBag,
  CheckCircle,
  XCircle,
  FileText,
  Loader2,
  AlertTriangle,
  Receipt,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

interface OrderLine {
  id: string;
  product_id: string;
  description: string;
  quantity: number;
  unit_price: number;
  subtotal: number;
  sort_order: number;
}

interface SaleOrder {
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
  updated_at: string;
  lines: OrderLine[];
}

const route = useRoute();
const orderId = route.params.id as string;

const order = ref<SaleOrder | null>(null);
const loading = ref(false);
const loadError = ref("");
const actionLoading = ref(false);
const actionError = ref("");
const confirmCancel = ref(false);

const STATUS_CONFIG: Record<
  string,
  { label: string; class: string; icon: Component }
> = {
  draft: {
    label: "Borrador",
    class: "bg-muted text-muted-foreground",
    icon: ShoppingBag,
  },
  confirmed: {
    label: "Confirmada",
    class: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
    icon: CheckCircle,
  },
  invoiced: {
    label: "Facturada",
    class:
      "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400",
    icon: Receipt,
  },
  cancelled: {
    label: "Cancelada",
    class: "bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400",
    icon: XCircle,
  },
};

const ID_TYPE_LABEL: Record<string, string> = {
  anonymous: "Anónimo",
  cedula: "Cédula",
  rif: "RIF",
  passport: "Pasaporte",
};

function fmtDate(d: string | null) {
  if (!d) return "—";
  return new Date(d).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function fmtCurrency(n: number) {
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
}

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    order.value = await useApiFetch<SaleOrder>(
      `/api/v1/sales-orders/${orderId}`,
    );
  } catch {
    loadError.value = "No se pudo cargar la orden de venta";
  } finally {
    loading.value = false;
  }
}

async function performAction(action: "confirm" | "invoice" | "cancel") {
  actionLoading.value = true;
  actionError.value = "";
  confirmCancel.value = false;
  try {
    order.value = await useApiFetch<SaleOrder>(
      `/api/v1/sales-orders/${orderId}/${action}`,
      { method: "POST" },
    );
  } catch (err: unknown) {
    const e = err as { data?: { detail?: string } };
    actionError.value = e?.data?.detail ?? "Ocurrió un error";
  } finally {
    actionLoading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6">
    <!-- Back -->
    <div class="flex items-center gap-3">
      <NuxtLink
        to="/sales-orders"
        class="p-2 rounded-lg hover:bg-muted transition-colors text-muted-foreground"
      >
        <ArrowLeft class="w-4 h-4" />
      </NuxtLink>
      <h1 class="text-xl font-bold text-foreground">Orden de venta</h1>
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
      Cargando...
    </div>

    <template v-else-if="order">
      <!-- Header card -->
      <div class="border bg-card rounded-xl p-6 space-y-5">
        <!-- Status + total -->
        <div class="flex items-start justify-between gap-4 flex-wrap">
          <div>
            <div class="flex items-center gap-3 mb-2">
              <span
                :class="[
                  'inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-sm font-medium',
                  STATUS_CONFIG[order.status]?.class,
                ]"
              >
                <component
                  :is="STATUS_CONFIG[order.status]?.icon"
                  class="w-4 h-4"
                />
                {{ STATUS_CONFIG[order.status]?.label }}
              </span>
            </div>
            <p class="text-lg font-semibold text-foreground">
              {{ order.customer_name || "Cliente anónimo" }}
            </p>
            <p
              v-if="
                order.customer_id_type !== 'anonymous' &&
                order.customer_id_number
              "
              class="text-sm text-muted-foreground font-mono"
            >
              {{ ID_TYPE_LABEL[order.customer_id_type] }}:
              {{ order.customer_id_number }}
            </p>
          </div>
          <div class="text-right">
            <p class="text-3xl font-bold text-foreground font-mono">
              {{ fmtCurrency(order.total) }}
            </p>
            <p class="text-sm text-muted-foreground mt-0.5">
              Total de la orden
            </p>
          </div>
        </div>

        <!-- Dates -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
          <div>
            <p class="text-muted-foreground text-xs">Creada</p>
            <p class="text-foreground font-medium">
              {{ fmtDate(order.created_at) }}
            </p>
          </div>
          <div>
            <p class="text-muted-foreground text-xs">Confirmada</p>
            <p class="text-foreground font-medium">
              {{ fmtDate(order.confirmed_at) }}
            </p>
          </div>
          <div>
            <p class="text-muted-foreground text-xs">Facturada</p>
            <p class="text-foreground font-medium">
              {{ fmtDate(order.invoiced_at) }}
            </p>
          </div>
        </div>

        <div
          v-if="order.notes"
          class="p-3 bg-muted/40 rounded-lg text-sm text-muted-foreground"
        >
          {{ order.notes }}
        </div>

        <!-- Invoice link -->
        <div
          v-if="order.invoice_id"
          class="flex items-center gap-2 text-sm text-emerald-600 dark:text-emerald-400"
        >
          <Receipt class="w-4 h-4" />
          Factura interna generada ·
          <NuxtLink
            :to="`/invoices/${order.invoice_id}`"
            class="underline hover:no-underline font-medium"
          >
            Ver factura
          </NuxtLink>
        </div>

        <!-- Action error -->
        <div
          v-if="actionError"
          class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3 flex items-center gap-2"
        >
          <AlertTriangle class="w-4 h-4 shrink-0" />
          {{ actionError }}
        </div>

        <!-- Actions -->
        <div
          v-if="order.status !== 'invoiced' && order.status !== 'cancelled'"
          class="flex items-center gap-3 flex-wrap"
        >
          <!-- Confirm (draft → confirmed) -->
          <button
            v-if="order.status === 'draft'"
            :disabled="actionLoading"
            class="flex items-center gap-2 px-4 py-2 text-sm rounded-lg bg-blue-600 text-white hover:bg-blue-700 transition-colors disabled:opacity-50"
            @click="performAction('confirm')"
          >
            <Loader2 v-if="actionLoading" class="w-4 h-4 animate-spin" />
            <CheckCircle v-else class="w-4 h-4" />
            Confirmar orden
          </button>

          <!-- Invoice (confirmed → invoiced) -->
          <button
            v-if="order.status === 'confirmed'"
            :disabled="actionLoading"
            class="flex items-center gap-2 px-4 py-2 text-sm rounded-lg bg-emerald-600 text-white hover:bg-emerald-700 transition-colors disabled:opacity-50"
            @click="performAction('invoice')"
          >
            <Loader2 v-if="actionLoading" class="w-4 h-4 animate-spin" />
            <FileText v-else class="w-4 h-4" />
            Generar factura
          </button>

          <!-- Cancel -->
          <div v-if="!confirmCancel">
            <button
              :disabled="actionLoading"
              class="flex items-center gap-2 px-4 py-2 text-sm rounded-lg border border-destructive/50 text-destructive hover:bg-destructive/10 transition-colors disabled:opacity-50"
              @click="confirmCancel = true"
            >
              <XCircle class="w-4 h-4" />
              Cancelar orden
            </button>
          </div>
          <div v-else class="flex items-center gap-2">
            <span class="text-sm text-muted-foreground">¿Estás seguro?</span>
            <button
              :disabled="actionLoading"
              class="flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg bg-destructive text-destructive-foreground hover:bg-destructive/90 transition-colors disabled:opacity-50"
              @click="performAction('cancel')"
            >
              <Loader2 v-if="actionLoading" class="w-3.5 h-3.5 animate-spin" />
              Sí, cancelar
            </button>
            <button
              class="px-3 py-1.5 text-sm rounded-lg border border-border hover:bg-muted transition-colors"
              @click="confirmCancel = false"
            >
              No
            </button>
          </div>
        </div>

        <!-- Final state info -->
        <div
          v-if="order.status === 'invoiced'"
          class="flex items-center gap-2 text-sm text-emerald-600 dark:text-emerald-400"
        >
          <CheckCircle class="w-4 h-4" />
          Facturada el {{ fmtDate(order.invoiced_at) }} — stock deducido y
          factura generada
        </div>
        <div
          v-if="order.status === 'cancelled'"
          class="flex items-center gap-2 text-sm text-rose-600 dark:text-rose-400"
        >
          <XCircle class="w-4 h-4" />
          Orden cancelada — stock restaurado automáticamente si estaba
          confirmada
        </div>
      </div>

      <!-- Lines table -->
      <div class="border bg-card rounded-xl overflow-hidden">
        <div class="px-6 py-4 border-b border-border">
          <h3 class="font-semibold text-foreground">
            Líneas de la orden ({{ order.lines?.length ?? 0 }})
          </h3>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border bg-muted/30">
                <th
                  class="px-4 py-3 text-left font-medium text-muted-foreground"
                >
                  Descripción
                </th>
                <th
                  class="px-4 py-3 text-right font-medium text-muted-foreground"
                >
                  Cantidad
                </th>
                <th
                  class="px-4 py-3 text-right font-medium text-muted-foreground hidden md:table-cell"
                >
                  Precio unit.
                </th>
                <th
                  class="px-4 py-3 text-right font-medium text-muted-foreground"
                >
                  Subtotal
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <tr
                v-for="line in order.lines"
                :key="line.id"
                class="hover:bg-muted/20 transition-colors"
              >
                <td class="px-4 py-3 text-foreground">
                  {{ line.description }}
                </td>
                <td
                  class="px-4 py-3 text-right font-mono text-muted-foreground"
                >
                  {{ line.quantity }}
                </td>
                <td
                  class="px-4 py-3 text-right font-mono text-muted-foreground hidden md:table-cell"
                >
                  {{ fmtCurrency(line.unit_price) }}
                </td>
                <td
                  class="px-4 py-3 text-right font-mono font-medium text-foreground"
                >
                  {{ fmtCurrency(line.subtotal) }}
                </td>
              </tr>
            </tbody>
            <tfoot>
              <tr class="border-t-2 border-border bg-muted/20">
                <td
                  colspan="3"
                  class="px-4 py-3 text-right text-sm font-medium text-muted-foreground"
                >
                  Total
                </td>
                <td
                  class="px-4 py-3 text-right font-mono font-bold text-lg text-foreground"
                >
                  {{ fmtCurrency(order.total) }}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>
