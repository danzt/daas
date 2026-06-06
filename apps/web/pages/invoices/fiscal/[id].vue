<script setup lang="ts">
import {
  Receipt,
  ArrowLeft,
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle,
  RefreshCw,
  Loader2,
  Send,
  Ban,
  RotateCcw,
  Download,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import { useAuthStore } from "~/stores/auth";
import type { FiscalInvoice } from "~/components/invoices/FiscalInvoiceFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

const route = useRoute();
const router = useRouter();
const invoiceId = route.params.id as string;

// ─── Data ────────────────────────────────────────────────────────────────────
const invoice = ref<FiscalInvoice | null>(null);
const loading = ref(false);
const loadError = ref("");

// ─── Actions ─────────────────────────────────────────────────────────────────
const issuing = ref(false);
const cancelling = ref(false);
const retrying = ref(false);
const cancelNotes = ref("");
const cancelDialogOpen = ref(false);
const actionError = ref("");

// ─── Load ─────────────────────────────────────────────────────────────────────
async function loadInvoice() {
  loading.value = true;
  loadError.value = "";
  try {
    invoice.value = await useApiFetch<FiscalInvoice>(
      `/api/v1/invoices/fiscal/${invoiceId}`,
    );
  } catch {
    loadError.value = "No se pudo cargar la factura fiscal";
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
    invoice.value = await useApiFetch<FiscalInvoice>(
      `/api/v1/invoices/fiscal/${invoiceId}/issue`,
      { method: "POST" },
    );
  } catch (err: unknown) {
    const apiError = err as { data?: { detail?: string } };
    actionError.value = apiError?.data?.detail ?? "Error al emitir la factura";
  } finally {
    issuing.value = false;
  }
}

// ─── Retry ────────────────────────────────────────────────────────────────────
async function handleRetry() {
  retrying.value = true;
  actionError.value = "";
  try {
    invoice.value = await useApiFetch<FiscalInvoice>(
      `/api/v1/invoices/fiscal/${invoiceId}/retry`,
      { method: "POST" },
    );
  } catch (err: unknown) {
    const apiError = err as { data?: { detail?: string } };
    actionError.value =
      apiError?.data?.detail ?? "Error al reintentar el envío";
  } finally {
    retrying.value = false;
  }
}

// ─── Cancel ───────────────────────────────────────────────────────────────────
async function handleCancel() {
  cancelling.value = true;
  actionError.value = "";
  try {
    invoice.value = await useApiFetch<FiscalInvoice>(
      `/api/v1/invoices/fiscal/${invoiceId}/cancel`,
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
  pending_fiscal: {
    label: "Enviando a SENIAT...",
    icon: RefreshCw,
    class: "bg-blue-50 text-blue-600",
  },
  issued: {
    label: "Emitida",
    icon: CheckCircle,
    class: "bg-green-50 text-green-700",
  },
  failed: {
    label: "Fallida",
    icon: AlertCircle,
    class: "bg-orange-50 text-orange-600",
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

function formatTaxRate(rate: number) {
  return `${(rate * 100).toFixed(0)}%`;
}

// ─── PDF download ─────────────────────────────────────────────────────────────
const downloadingPDF = ref(false);

async function downloadPDF() {
  if (!invoice.value) return;
  downloadingPDF.value = true;
  try {
    const store = useAuthStore();
    const config = useRuntimeConfig();
    const url = `${config.public.apiBase}/api/v1/invoices/fiscal/${invoiceId}/pdf`;
    const resp = await fetch(url, {
      headers: { Authorization: `Bearer ${store.accessToken}` },
    });
    if (!resp.ok) throw new Error("error");
    const blob = await resp.blob();
    const link = document.createElement("a");
    link.href = URL.createObjectURL(blob);
    const name = invoice.value.fiscal_number ?? invoiceId;
    link.download = `factura-fiscal-${name}.pdf`;
    link.click();
    URL.revokeObjectURL(link.href);
  } finally {
    downloadingPDF.value = false;
  }
}
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6">
    <button
      type="button"
      class="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
      @click="router.push('/invoices/fiscal')"
    >
      <ArrowLeft class="w-4 h-4" />
      Volver a Facturas Fiscales
    </button>

    <div v-if="loading" class="flex items-center justify-center py-24">
      <Loader2 class="w-7 h-7 text-primary animate-spin" />
    </div>

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
      <div class="rounded-xl border bg-card shadow-sm p-6">
        <div
          class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4"
        >
          <div class="flex items-start gap-4">
            <div
              class="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center flex-shrink-0"
            >
              <Receipt class="w-6 h-6 text-primary" />
            </div>
            <div>
              <div class="flex items-center gap-3 flex-wrap">
                <h1
                  class="text-xl font-bold font-heading text-foreground font-mono"
                >
                  {{ invoice.fiscal_number ?? "Sin número fiscal" }}
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
              <!-- Fiscal machine info -->
              <div
                v-if="invoice.machine_serial"
                class="flex items-center gap-4 mt-2 text-xs text-muted-foreground"
              >
                <span
                  >Máquina:
                  <span class="font-mono">{{
                    invoice.machine_serial
                  }}</span></span
                >
                <span v-if="invoice.report_z_number">
                  Reporte Z:
                  <span class="font-mono">{{ invoice.report_z_number }}</span>
                </span>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex items-center gap-2 flex-shrink-0">
            <!-- Draft: Issue + Cancel -->
            <template v-if="invoice.status === 'draft'">
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
                {{ issuing ? "Enviando a SENIAT..." : "Emitir" }}
              </button>
            </template>

            <!-- Failed: Retry + Cancel -->
            <template v-else-if="invoice.status === 'failed'">
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
                :disabled="retrying || invoice.retry_count >= 3"
                class="flex items-center gap-2 h-9 px-5 bg-orange-500 text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer disabled:opacity-50"
                @click="handleRetry"
              >
                <Loader2 v-if="retrying" class="w-4 h-4 animate-spin" />
                <RotateCcw v-else class="w-4 h-4" />
                {{
                  retrying
                    ? "Reintentando..."
                    : `Reintentar (${invoice.retry_count}/3)`
                }}
              </button>
            </template>

            <!-- Issued: PDF download -->
            <template v-else-if="invoice.status === 'issued'">
              <button
                type="button"
                :disabled="downloadingPDF"
                class="flex items-center gap-2 h-9 px-4 border border-primary text-primary text-sm font-semibold rounded-lg hover:bg-primary/5 transition-all duration-200 cursor-pointer disabled:opacity-50"
                @click="downloadPDF"
              >
                <Loader2 v-if="downloadingPDF" class="w-4 h-4 animate-spin" />
                <Download v-else class="w-4 h-4" />
                PDF
              </button>
            </template>
          </div>
        </div>

        <!-- Fail reason -->
        <div
          v-if="invoice.fail_reason"
          class="mt-4 rounded-lg bg-orange-50 border border-orange-200 px-4 py-3"
        >
          <p class="text-xs font-semibold text-orange-700 mb-0.5">
            Error SENIAT
          </p>
          <p class="text-sm text-orange-700 font-mono">
            {{ invoice.fail_reason }}
          </p>
        </div>

        <!-- Action error -->
        <div
          v-if="actionError"
          class="mt-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3"
        >
          <p class="text-sm text-red-600">{{ actionError }}</p>
        </div>
      </div>

      <!-- 2-col: customer info + lines -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Customer -->
        <div class="rounded-xl border bg-card shadow-sm p-5">
          <h2 class="text-sm font-bold text-foreground mb-4">Cliente</h2>
          <div class="space-y-3">
            <div>
              <p class="text-xs text-muted-foreground">Nombre</p>
              <p class="text-sm font-semibold text-foreground">
                {{ invoice.customer_name || "Consumidor Final" }}
              </p>
            </div>
            <div>
              <p class="text-xs text-muted-foreground">Identificación</p>
              <p class="text-sm font-semibold text-foreground">
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
              <p class="text-sm text-foreground">{{ invoice.notes }}</p>
            </div>
          </div>
        </div>

        <!-- Lines + totals -->
        <div
          class="lg:col-span-2 rounded-xl border bg-card shadow-sm overflow-hidden"
        >
          <div class="px-5 py-4 border-b">
            <h2 class="text-sm font-bold text-foreground">Detalle con IVA</h2>
          </div>
          <div class="overflow-x-auto">
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
                    class="px-4 py-3 text-right text-xs font-semibold text-muted-foreground"
                  >
                    IVA
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
                    <p class="text-sm font-medium text-foreground">
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
                  <td class="px-4 py-3 text-right">
                    <span
                      class="text-xs tabular-nums text-muted-foreground font-mono"
                    >
                      {{ formatTaxRate(line.tax_rate) }}
                      <br />
                      <span class="text-foreground">{{
                        formatCurrency(line.tax_amount)
                      }}</span>
                    </span>
                  </td>
                  <td class="px-5 py-3 text-right">
                    <span
                      class="text-sm font-semibold tabular-nums text-foreground"
                    >
                      {{ formatCurrency(line.subtotal) }}
                    </span>
                  </td>
                </tr>
              </tbody>
              <tfoot class="border-t-2 bg-muted/10">
                <tr>
                  <td
                    colspan="4"
                    class="px-5 py-2 text-right text-xs text-muted-foreground"
                  >
                    Base imponible
                  </td>
                  <td
                    class="px-5 py-2 text-right text-sm tabular-nums text-muted-foreground"
                  >
                    {{ formatCurrency(invoice.subtotal_base) }}
                  </td>
                </tr>
                <tr>
                  <td
                    colspan="4"
                    class="px-5 py-2 text-right text-xs text-muted-foreground"
                  >
                    IVA total
                  </td>
                  <td
                    class="px-5 py-2 text-right text-sm tabular-nums text-muted-foreground"
                  >
                    {{ formatCurrency(invoice.tax_amount) }}
                  </td>
                </tr>
                <tr class="border-t">
                  <td
                    colspan="4"
                    class="px-5 py-4 text-right text-sm font-bold text-foreground"
                  >
                    Total
                  </td>
                  <td
                    class="px-5 py-4 text-right text-lg font-bold font-heading text-foreground tabular-nums"
                  >
                    {{ formatCurrency(invoice.total) }}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
      </div>
    </template>

    <!-- Cancel dialog -->
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
            <h3 class="text-lg font-bold font-heading text-foreground mb-1">
              Cancelar factura fiscal
            </h3>
            <p class="text-sm text-muted-foreground mb-5">
              Solo se pueden cancelar facturas en estado borrador o fallidas.
              ¿Confirmar?
            </p>
            <div class="mb-5">
              <label class="block text-sm font-semibold text-foreground mb-1.5">
                Motivo
                <span class="font-normal text-muted-foreground"
                  >(opcional)</span
                >
              </label>
              <textarea
                v-model="cancelNotes"
                rows="2"
                placeholder="Motivo de cancelación..."
                class="w-full px-4 py-3 border rounded-lg text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-muted-foreground resize-none"
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
                <span>{{ cancelling ? "Cancelando..." : "Confirmar" }}</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
