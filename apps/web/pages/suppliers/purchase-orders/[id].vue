<script setup lang="ts">
import {
  ArrowLeft,
  ShoppingCart,
  CheckCircle,
  XCircle,
  Package,
  Loader2,
  AlertTriangle,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

interface POLine {
  id: string;
  product_id: string;
  description: string;
  quantity_ordered: number;
  unit_cost: number;
  subtotal: number;
  sort_order: number;
}

interface Supplier {
  id: string;
  name: string;
  rif: string;
  contact_name: string;
  email: string;
  phone: string;
}

interface PurchaseOrder {
  id: string;
  supplier_id: string;
  status: "draft" | "ordered" | "received" | "cancelled";
  notes: string;
  total: number;
  ordered_at: string | null;
  received_at: string | null;
  created_at: string;
  updated_at: string;
  lines: POLine[];
  supplier?: Supplier;
}

const { isMobile } = useMobileMode();
const route = useRoute();
const poId = route.params.id as string;

const po = ref<PurchaseOrder | null>(null);
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
    icon: ShoppingCart,
  },
  ordered: {
    label: "Ordenada",
    class: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
    icon: Package,
  },
  received: {
    label: "Recibida",
    class:
      "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400",
    icon: CheckCircle,
  },
  cancelled: {
    label: "Cancelada",
    class: "bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400",
    icon: XCircle,
  },
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
    po.value = await useApiFetch<PurchaseOrder>(
      `/api/v1/purchase-orders/${poId}`,
    );
  } catch {
    loadError.value = "No se pudo cargar la orden de compra";
  } finally {
    loading.value = false;
  }
}

async function performAction(action: "order" | "receive" | "cancel") {
  actionLoading.value = true;
  actionError.value = "";
  confirmCancel.value = false;
  try {
    po.value = await useApiFetch<PurchaseOrder>(
      `/api/v1/purchase-orders/${poId}/${action}`,
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
  <MobileScreensPurchaseOrderDetail
    v-if="isMobile"
    :po="po"
    :loading="loading"
    :load-error="loadError"
    :action-loading="actionLoading"
    :action-error="actionError"
    @order="performAction('order')"
    @receive="performAction('receive')"
    @cancel="performAction('cancel')"
    @retry="load"
  />
  <div v-else class="p-4 sm:p-6 space-y-6">
    <!-- Back -->
    <div class="flex items-center gap-3">
      <NuxtLink
        to="/suppliers/purchase-orders"
        class="p-2 rounded-lg hover:bg-muted transition-colors text-muted-foreground"
      >
        <ArrowLeft class="w-4 h-4" />
      </NuxtLink>
      <h1 class="text-xl font-bold text-foreground">Orden de compra</h1>
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

    <template v-else-if="po">
      <!-- Header card -->
      <div class="border bg-card rounded-xl p-6 space-y-5">
        <!-- Status + supplier -->
        <div class="flex items-start justify-between gap-4 flex-wrap">
          <div>
            <div class="flex items-center gap-3 mb-2">
              <span
                :class="[
                  'inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-sm font-medium',
                  STATUS_CONFIG[po.status]?.class,
                ]"
              >
                <component
                  :is="STATUS_CONFIG[po.status]?.icon"
                  class="w-4 h-4"
                />
                {{ STATUS_CONFIG[po.status]?.label }}
              </span>
            </div>
            <p class="text-lg font-semibold text-foreground">
              {{ po.supplier?.name ?? "Proveedor desconocido" }}
            </p>
            <p
              v-if="po.supplier?.rif"
              class="text-sm text-muted-foreground font-mono"
            >
              RIF: {{ po.supplier.rif }}
            </p>
          </div>
          <div class="text-right">
            <p class="text-3xl font-bold text-foreground font-mono">
              {{ fmtCurrency(po.total) }}
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
              {{ fmtDate(po.created_at) }}
            </p>
          </div>
          <div>
            <p class="text-muted-foreground text-xs">Ordenada</p>
            <p class="text-foreground font-medium">
              {{ fmtDate(po.ordered_at) }}
            </p>
          </div>
          <div>
            <p class="text-muted-foreground text-xs">Recibida</p>
            <p class="text-foreground font-medium">
              {{ fmtDate(po.received_at) }}
            </p>
          </div>
        </div>

        <div
          v-if="po.notes"
          class="p-3 bg-muted/40 rounded-lg text-sm text-muted-foreground"
        >
          {{ po.notes }}
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
          v-if="po.status !== 'received' && po.status !== 'cancelled'"
          class="flex items-center gap-3 flex-wrap"
        >
          <!-- Order (draft → ordered) -->
          <button
            v-if="po.status === 'draft'"
            :disabled="actionLoading"
            class="flex items-center gap-2 px-4 py-2 text-sm rounded-lg bg-blue-600 text-white hover:bg-blue-700 transition-colors disabled:opacity-50"
            @click="performAction('order')"
          >
            <Loader2 v-if="actionLoading" class="w-4 h-4 animate-spin" />
            <Package v-else class="w-4 h-4" />
            Confirmar orden
          </button>

          <!-- Receive (ordered → received) -->
          <button
            v-if="po.status === 'ordered'"
            :disabled="actionLoading"
            class="flex items-center gap-2 px-4 py-2 text-sm rounded-lg bg-emerald-600 text-white hover:bg-emerald-700 transition-colors disabled:opacity-50"
            @click="performAction('receive')"
          >
            <Loader2 v-if="actionLoading" class="w-4 h-4 animate-spin" />
            <CheckCircle v-else class="w-4 h-4" />
            Marcar como recibida
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

        <!-- Received info -->
        <div
          v-if="po.status === 'received'"
          class="flex items-center gap-2 text-sm text-emerald-600 dark:text-emerald-400"
        >
          <CheckCircle class="w-4 h-4" />
          Recibida el {{ fmtDate(po.received_at) }} — inventario actualizado
          automáticamente
        </div>
      </div>

      <!-- Lines table -->
      <div class="border bg-card rounded-xl overflow-hidden">
        <div class="px-6 py-4 border-b border-border">
          <h3 class="font-semibold text-foreground">
            Líneas de la orden ({{ po.lines?.length ?? 0 }})
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
                  Costo unit.
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
                v-for="line in po.lines"
                :key="line.id"
                class="hover:bg-muted/20 transition-colors"
              >
                <td class="px-4 py-3 text-foreground">
                  {{ line.description }}
                </td>
                <td
                  class="px-4 py-3 text-right font-mono text-muted-foreground"
                >
                  {{ line.quantity_ordered }}
                </td>
                <td
                  class="px-4 py-3 text-right font-mono text-muted-foreground hidden md:table-cell"
                >
                  {{ fmtCurrency(line.unit_cost) }}
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
                  {{ fmtCurrency(po.total) }}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>
