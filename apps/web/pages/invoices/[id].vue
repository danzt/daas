<script setup lang="ts">
import {
  FileText,
  ArrowLeft,
  CheckCircle,
  XCircle,
  Clock,
  Loader2,
  Send,
  Ban,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import type { Invoice } from "~/components/invoices/InvoiceFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

const route = useRoute();
const router = useRouter();
const invoiceId = route.params.id as string;

// ─── Data ────────────────────────────────────────────────────────────────────
const invoice = ref<Invoice | null>(null);
const loading = ref(false);
const loadError = ref("");

// ─── Actions ─────────────────────────────────────────────────────────────────
const issuing = ref(false);
const cancelling = ref(false);
const cancelNotes = ref("");
const cancelDialogOpen = ref(false);
const actionError = ref("");

// ─── Load ─────────────────────────────────────────────────────────────────────
async function loadInvoice() {
  loading.value = true;
  loadError.value = "";
  try {
    invoice.value = await useApiFetch<Invoice>(
      `/api/v1/invoices/internal/${invoiceId}`,
    );
  } catch {
    loadError.value = "No se pudo cargar la factura";
  } finally {
    loading.value = false;
  }
}

onMounted(loadInvoice);

// ─── Issue ────────────────────────────────────────────────────────────────────
async function handleIssue() {
  issuing.value = true;
  actionError.value = "";
  try {
    invoice.value = await useApiFetch<Invoice>(
      `/api/v1/invoices/internal/${invoiceId}/issue`,
      { method: "POST" },
    );
  } catch (err: unknown) {
    const apiError = err as { data?: { detail?: string } };
    actionError.value = apiError?.data?.detail ?? "Error al emitir la factura";
  } finally {
    issuing.value = false;
  }
}

// ─── Cancel ───────────────────────────────────────────────────────────────────
async function handleCancel() {
  cancelling.value = true;
  actionError.value = "";
  try {
    invoice.value = await useApiFetch<Invoice>(
      `/api/v1/invoices/internal/${invoiceId}/cancel`,
      { method: "POST", body: { notes: cancelNotes.value.trim() } },
    );
    cancelDialogOpen.value = false;
    cancelNotes.value = "";
  } catch (err: unknown) {
    const apiError = err as { data?: { detail?: string } };
    actionError.value =
      apiError?.data?.detail ?? "Error al cancelar la factura";
  } finally {
    cancelling.value = false;
  }
}

// ─── Helpers ─────────────────────────────────────────────────────────────────
const STATUS_CONFIG = {
  draft: {
    label: "Borrador",
    icon: Clock,
    class: "bg-muted text-muted-foreground",
  },
  issued: {
    label: "Emitida",
    icon: CheckCircle,
    class: "bg-green-50 text-green-700",
  },
  cancelled: {
    label: "Cancelada",
    icon: XCircle,
    class: "bg-red-50 text-red-600",
  },
} as const;

const CUSTOMER_ID_LABELS: Record<string, string> = {
  cedula: "Cédula",
  rif: "RIF",
  passport: "Pasaporte",
  anonymous: "Anónimo",
};

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatCurrency(value: number) {
  return value.toFixed(2);
}
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Back nav -->
    <button
      type="button"
      class="flex items-center gap-2 text-sm text-muted-foreground hover:text-text-brand transition-colors cursor-pointer"
      @click="router.push('/invoices')"
    >
      <ArrowLeft class="w-4 h-4" />
      Volver a Facturas
    </button>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-24">
      <Loader2 class="w-7 h-7 text-primary animate-spin" />
    </div>

    <!-- Error -->
    <div
      v-else-if="loadError"
      class="flex flex-col items-center justify-center py-24 gap-3"
    >
      <p class="text-sm text-red-600">{{ loadError }}</p>
      <button
        type="button"
        class="text-xs text-primary font-semibold hover:underline cursor-pointer"
        @click="loadInvoice"
      >
        Reintentar
      </button>
    </div>

    <template v-else-if="invoice">
      <!-- Header card -->
      <div class="bg-card rounded-xl border shadow-sm p-6">
        <div
          class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4"
        >
          <!-- Left: invoice info -->
          <div class="flex items-start gap-4">
            <div
              class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center flex-shrink-0"
            >
              <FileText class="w-6 h-6 text-primary" />
            </div>
            <div>
              <div class="flex items-center gap-3">
                <h1 class="text-xl font-bold font-heading text-text-brand">
                  {{ invoice.correlative || "Borrador" }}
                </h1>
                <span
                  :class="[
                    'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold',
                    STATUS_CONFIG[invoice.status].class,
                  ]"
                >
                  <component
                    :is="STATUS_CONFIG[invoice.status].icon"
                    class="w-3 h-3"
                  />
                  {{ STATUS_CONFIG[invoice.status].label }}
                </span>
              </div>
              <p class="text-sm text-muted-foreground mt-1">
                Creada el {{ formatDate(invoice.created_at) }}
              </p>
              <p v-if="invoice.issued_at" class="text-xs text-muted-foreground">
                Emitida el {{ formatDate(invoice.issued_at) }}
              </p>
            </div>
          </div>

          <!-- Actions -->
          <div
            v-if="invoice.status === 'draft'"
            class="flex items-center gap-2 flex-shrink-0"
          >
            <button
              type="button"
              class="flex items-center gap-2 h-9 px-4 border text-sm font-semibold text-muted-foreground rounded-lg hover:bg-red-50 hover:text-red-600 hover:border-red-200 transition-all duration-200 cursor-pointer"
              @click="cancelDialogOpen = true"
            >
              <Ban class="w-4 h-4" />
              Cancelar
            </button>
            <button
              type="button"
              :disabled="issuing"
              class="flex items-center gap-2 h-9 px-5 bg-primary text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer disabled:opacity-50"
              @click="handleIssue"
            >
              <Loader2 v-if="issuing" class="w-4 h-4 animate-spin" />
              <Send v-else class="w-4 h-4" />
              {{ issuing ? "Emitiendo..." : "Emitir" }}
            </button>
          </div>

          <div
            v-else-if="invoice.status === 'issued'"
            class="flex items-center gap-2 flex-shrink-0"
          >
            <button
              type="button"
              class="flex items-center gap-2 h-9 px-4 border text-sm font-semibold text-muted-foreground rounded-lg hover:bg-red-50 hover:text-red-600 hover:border-red-200 transition-all duration-200 cursor-pointer"
              @click="cancelDialogOpen = true"
            >
              <Ban class="w-4 h-4" />
              Cancelar
            </button>
          </div>
        </div>

        <!-- Action error -->
        <div
          v-if="actionError"
          class="mt-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3"
        >
          <p class="text-sm text-red-600">{{ actionError }}</p>
        </div>
      </div>

      <!-- 2-col layout: customer + lines -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Customer info -->
        <div class="bg-card rounded-xl border shadow-sm p-5">
          <h2 class="text-sm font-bold text-text-brand mb-4">Cliente</h2>
          <div class="space-y-3">
            <div>
              <p class="text-xs text-muted-foreground">Nombre</p>
              <p class="text-sm font-semibold text-text-brand">
                {{ invoice.customer_name || "Consumidor Final" }}
              </p>
            </div>
            <div>
              <p class="text-xs text-muted-foreground">Identificación</p>
              <p class="text-sm font-semibold text-text-brand">
                {{
                  CUSTOMER_ID_LABELS[invoice.customer_id_type] ??
                  invoice.customer_id_type
                }}
                <span v-if="invoice.customer_id_number" class="ml-1 font-mono">
                  {{ invoice.customer_id_number }}
                </span>
              </p>
            </div>
            <div v-if="invoice.notes">
              <p class="text-xs text-muted-foreground">Notas</p>
              <p class="text-sm text-text-brand">{{ invoice.notes }}</p>
            </div>
          </div>
        </div>

        <!-- Lines table -->
        <div
          class="lg:col-span-2 bg-card rounded-xl border shadow-sm overflow-hidden"
        >
          <div class="px-5 py-4 border-b">
            <h2 class="text-sm font-bold text-text-brand">Detalle</h2>
          </div>
          <table class="w-full">
            <thead>
              <tr class="bg-muted/40 border-b">
                <th
                  class="px-5 py-3 text-left text-xs font-semibold text-muted-foreground"
                >
                  Descripción
                </th>
                <th
                  class="px-4 py-3 text-right text-xs font-semibold text-muted-foreground"
                >
                  Cant.
                </th>
                <th
                  class="px-4 py-3 text-right text-xs font-semibold text-muted-foreground"
                >
                  P. Unit.
                </th>
                <th
                  class="px-5 py-3 text-right text-xs font-semibold text-muted-foreground"
                >
                  Subtotal
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <tr
                v-for="line in invoice.lines"
                :key="line.id"
                class="hover:bg-muted/20 transition-colors"
              >
                <td class="px-5 py-3">
                  <p class="text-sm font-medium text-text-brand">
                    {{ line.description }}
                  </p>
                </td>
                <td class="px-4 py-3 text-right">
                  <span class="text-sm tabular-nums text-muted-foreground">
                    {{ line.quantity }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right">
                  <span class="text-sm tabular-nums text-muted-foreground">
                    {{ formatCurrency(line.unit_price) }}
                  </span>
                </td>
                <td class="px-5 py-3 text-right">
                  <span
                    class="text-sm font-semibold tabular-nums text-text-brand"
                  >
                    {{ formatCurrency(line.subtotal) }}
                  </span>
                </td>
              </tr>
            </tbody>
            <tfoot>
              <tr class="border-t-2 bg-muted/20">
                <td
                  colspan="3"
                  class="px-5 py-4 text-right text-sm font-bold text-text-brand"
                >
                  Total
                </td>
                <td
                  class="px-5 py-4 text-right text-lg font-bold font-heading text-text-brand tabular-nums"
                >
                  {{ formatCurrency(invoice.total) }}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </template>

    <!-- Cancel confirmation dialog -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-200"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-200"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="cancelDialogOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4"
          role="dialog"
          aria-modal="true"
        >
          <div
            class="absolute inset-0 bg-black/50"
            @click="cancelDialogOpen = false"
          />
          <div
            class="relative z-10 w-full max-w-md bg-white rounded-xl shadow-xl p-6"
          >
            <h3 class="text-lg font-bold font-heading text-text-brand mb-1">
              Cancelar factura
            </h3>
            <p class="text-sm text-muted-foreground mb-5">
              Esta acción revertirá el stock descontado. ¿Confirmar?
            </p>
            <div class="mb-5">
              <label class="block text-sm font-semibold text-text-brand mb-1.5">
                Motivo
                <span class="font-normal text-muted-foreground"
                  >(opcional)</span
                >
              </label>
              <textarea
                v-model="cancelNotes"
                rows="2"
                placeholder="Motivo de cancelación..."
                class="w-full px-4 py-3 border rounded-lg text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-gray-400 resize-none"
              />
            </div>
            <div class="flex justify-end gap-3">
              <button
                type="button"
                class="h-10 px-5 border text-muted-foreground text-sm font-semibold rounded-lg hover:bg-muted transition-all duration-200 cursor-pointer"
                @click="cancelDialogOpen = false"
              >
                Volver
              </button>
              <button
                type="button"
                :disabled="cancelling"
                class="flex items-center gap-2 h-10 px-5 bg-red-600 text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer disabled:opacity-50"
                @click="handleCancel"
              >
                <Loader2 v-if="cancelling" class="w-4 h-4 animate-spin" />
                <span>{{
                  cancelling ? "Cancelando..." : "Confirmar cancelación"
                }}</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
