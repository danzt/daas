<script setup lang="ts">
import {
  Receipt,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  RefreshCw,
  RotateCcw,
  Send,
  Ban,
  Download,
  Loader2,
  AlertTriangle,
} from "lucide-vue-next";
import type { FiscalInvoice } from "~/components/invoices/FiscalInvoiceFormModal.vue";

const props = defineProps<{
  invoice: FiscalInvoice | null;
  loading: boolean;
  loadError: string;
  issuing: boolean;
  retrying: boolean;
  cancelling: boolean;
  downloadingPDF: boolean;
  actionError: string;
}>();

const emit = defineEmits<{
  issue: [];
  retry: [];
  cancel: [notes: string];
  downloadPdf: [];
  retryLoad: [];
}>();

const confirmingCancel = ref(false);
const cancelNotes = ref("");

const STATUS: Record<string, { label: string; chip: string; icon: Component }> =
  {
    draft: {
      label: "Borrador",
      chip: "bg-gray-100 text-gray-600",
      icon: Clock,
    },
    pending_fiscal: {
      label: "Enviando a SENIAT…",
      chip: "bg-blue-100 text-blue-700",
      icon: RefreshCw,
    },
    issued: {
      label: "Emitida",
      chip: "bg-emerald-100 text-emerald-700",
      icon: CheckCircle,
    },
    failed: {
      label: "Fallida",
      chip: "bg-orange-100 text-orange-600",
      icon: AlertCircle,
    },
    cancelled: {
      label: "Cancelada",
      chip: "bg-rose-100 text-rose-600",
      icon: XCircle,
    },
  };

const CUSTOMER_ID_LABELS: Record<string, string> = {
  cedula: "Cédula",
  rif: "RIF",
  passport: "Pasaporte",
  anonymous: "Anónimo",
};

function fmtDate(d: string | null | undefined) {
  if (!d) return "—";
  return new Date(d).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function fmtCurrency(n: number) {
  return n.toFixed(2);
}

function fmtTaxRate(r: number) {
  return `${(r * 100).toFixed(0)}%`;
}

function submitCancel() {
  emit("cancel", cancelNotes.value.trim());
  cancelNotes.value = "";
  confirmingCancel.value = false;
}

const hasActionBar = computed(
  () =>
    props.invoice &&
    ["draft", "failed", "issued"].includes(props.invoice.status),
);
</script>

<template>
  <MobileScreen title="Factura fiscal" back subtitle="SENIAT">
    <div v-if="loading" class="flex justify-center py-20">
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <div v-else-if="loadError" class="px-4 py-12 text-center">
      <AlertTriangle class="mx-auto mb-3 size-10 text-destructive" />
      <p class="font-medium text-destructive">{{ loadError }}</p>
      <button
        type="button"
        class="mt-4 h-10 rounded-xl border border-border px-5 text-sm font-semibold text-muted-foreground active:bg-accent"
        @click="emit('retryLoad')"
      >
        Reintentar
      </button>
    </div>

    <template v-else-if="invoice">
      <!-- Hero -->
      <div class="px-4 pt-4">
        <div class="rounded-2xl border border-border bg-white p-5 space-y-3">
          <div class="flex items-center justify-between gap-2">
            <span
              class="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-bold"
              :class="STATUS[invoice.status]?.chip"
            >
              <component :is="STATUS[invoice.status]?.icon" class="size-3.5" />
              {{ STATUS[invoice.status]?.label }}
            </span>
            <span
              v-if="invoice.fiscal_number"
              class="font-mono text-xs font-bold text-muted-foreground"
            >
              {{ invoice.fiscal_number }}
            </span>
          </div>

          <p class="text-2xl font-black tabular-nums text-foreground">
            {{ fmtCurrency(invoice.total) }}
          </p>

          <div>
            <p class="text-sm font-semibold text-foreground">
              {{ invoice.customer_name || "Consumidor Final" }}
            </p>
            <p
              v-if="invoice.customer_id_number"
              class="font-mono text-xs text-muted-foreground"
            >
              {{
                CUSTOMER_ID_LABELS[invoice.customer_id_type] ??
                invoice.customer_id_type
              }}
              {{ invoice.customer_id_number }}
            </p>
          </div>

          <div
            class="grid grid-cols-2 gap-2 pt-2 border-t border-border text-xs"
          >
            <div>
              <p class="text-muted-foreground">Creada</p>
              <p class="font-medium text-foreground">
                {{ fmtDate(invoice.created_at) }}
              </p>
            </div>
            <div v-if="invoice.issued_at">
              <p class="text-muted-foreground">Emitida</p>
              <p class="font-medium text-foreground">
                {{ fmtDate(invoice.issued_at) }}
              </p>
            </div>
          </div>

          <!-- Machine info -->
          <div
            v-if="invoice.machine_serial"
            class="pt-2 border-t border-border text-xs text-muted-foreground space-y-1"
          >
            <div class="flex gap-2">
              <span class="w-20 shrink-0">Máquina</span>
              <span class="font-mono">{{ invoice.machine_serial }}</span>
            </div>
            <div v-if="invoice.report_z_number" class="flex gap-2">
              <span class="w-20 shrink-0">Reporte Z</span>
              <span class="font-mono">{{ invoice.report_z_number }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Fail reason -->
      <div v-if="invoice.fail_reason" class="px-4 pt-3">
        <div class="rounded-xl border border-orange-200 bg-orange-50 px-4 py-3">
          <p class="text-xs font-bold text-orange-700">Error SENIAT</p>
          <p class="mt-1 font-mono text-sm text-orange-700">
            {{ invoice.fail_reason }}
          </p>
        </div>
      </div>

      <!-- Action error -->
      <div v-if="actionError" class="px-4 pt-3">
        <div
          class="flex items-center gap-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
        >
          <AlertTriangle class="size-4 shrink-0" />
          {{ actionError }}
        </div>
      </div>

      <!-- Lines -->
      <MobileSectionHeader title="Detalle con IVA" class="pt-5" />
      <div class="divide-y divide-border border-y border-border bg-white">
        <div
          v-for="line in invoice.lines"
          :key="line.id"
          class="flex items-start gap-3 px-4 py-3"
        >
          <span
            class="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-lg bg-muted text-xs font-bold text-muted-foreground"
          >
            {{ line.quantity }}×
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-foreground">
              {{ line.description }}
            </p>
            <p class="text-[11px] text-muted-foreground">
              {{ fmtCurrency(line.unit_price) }} c/u · IVA
              {{ fmtTaxRate(line.tax_rate) }}
              ({{ fmtCurrency(line.tax_amount) }})
            </p>
          </div>
          <span
            class="font-mono text-sm font-bold tabular-nums text-foreground"
          >
            {{ fmtCurrency(line.subtotal) }}
          </span>
        </div>

        <div class="flex items-center justify-between px-4 py-2.5 text-sm">
          <span class="text-muted-foreground">Base imponible</span>
          <span class="font-mono tabular-nums">{{
            fmtCurrency(invoice.subtotal_base)
          }}</span>
        </div>
        <div class="flex items-center justify-between px-4 py-2.5 text-sm">
          <span class="text-muted-foreground">IVA total</span>
          <span class="font-mono tabular-nums">{{
            fmtCurrency(invoice.tax_amount)
          }}</span>
        </div>
        <div class="flex items-center justify-between px-4 py-3.5">
          <span class="text-sm font-bold text-foreground">Total</span>
          <span class="font-mono text-lg font-black tabular-nums text-primary">
            {{ fmtCurrency(invoice.total) }}
          </span>
        </div>
      </div>

      <!-- Notes -->
      <div v-if="invoice.notes" class="px-4 pt-4">
        <p class="rounded-2xl bg-muted/60 p-3.5 text-sm text-muted-foreground">
          {{ invoice.notes }}
        </p>
      </div>

      <!-- Spacer for action bar -->
      <div v-if="hasActionBar" class="h-24" />
    </template>

    <!-- Action bar -->
    <template v-if="hasActionBar && invoice">
      <div
        class="fixed inset-x-0 bottom-0 z-40 border-t border-border bg-white px-4 pb-[calc(0.75rem+env(safe-area-inset-bottom))] pt-3"
      >
        <!-- Cancel confirm: inline notes input -->
        <div v-if="confirmingCancel" class="space-y-2">
          <input
            v-model="cancelNotes"
            class="h-10 w-full rounded-xl border border-border px-3 text-sm placeholder:text-muted-foreground focus:border-primary focus:outline-none"
            placeholder="Motivo (opcional)…"
          />
          <div class="flex gap-2">
            <button
              type="button"
              :disabled="cancelling"
              class="no-min-tap h-12 flex-1 rounded-2xl bg-red-500 text-sm font-bold text-white active:bg-red-600 disabled:opacity-50"
              @click="submitCancel"
            >
              <Loader2 v-if="cancelling" class="mx-auto size-5 animate-spin" />
              <span v-else>Confirmar cancelación</span>
            </button>
            <button
              type="button"
              class="no-min-tap h-12 rounded-2xl border border-border px-5 text-sm font-bold text-muted-foreground active:bg-accent"
              @click="confirmingCancel = false"
            >
              No
            </button>
          </div>
        </div>

        <!-- Normal action bar -->
        <div v-else class="flex gap-2">
          <!-- Cancel icon (draft or failed) -->
          <button
            v-if="invoice.status === 'draft' || invoice.status === 'failed'"
            type="button"
            class="no-min-tap h-12 w-12 shrink-0 rounded-2xl border border-rose-200 text-rose-500 active:bg-rose-50"
            aria-label="Cancelar factura"
            @click="confirmingCancel = true"
          >
            <Ban class="mx-auto size-5" />
          </button>

          <!-- draft → issue -->
          <button
            v-if="invoice.status === 'draft'"
            type="button"
            :disabled="issuing"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-primary text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('issue')"
          >
            <Loader2 v-if="issuing" class="mx-auto size-5 animate-spin" />
            <span v-else class="flex items-center justify-center gap-2">
              <Send class="size-4" /> Emitir a SENIAT
            </span>
          </button>

          <!-- failed → retry -->
          <button
            v-else-if="invoice.status === 'failed'"
            type="button"
            :disabled="retrying || invoice.retry_count >= 3"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-orange-500 text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('retry')"
          >
            <Loader2 v-if="retrying" class="mx-auto size-5 animate-spin" />
            <span v-else class="flex items-center justify-center gap-2">
              <RotateCcw class="size-4" />
              Reintentar ({{ invoice.retry_count }}/3)
            </span>
          </button>

          <!-- issued → PDF -->
          <button
            v-else-if="invoice.status === 'issued'"
            type="button"
            :disabled="downloadingPDF"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-primary text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('downloadPdf')"
          >
            <Loader2
              v-if="downloadingPDF"
              class="mx-auto size-5 animate-spin"
            />
            <span v-else class="flex items-center justify-center gap-2">
              <Download class="size-4" /> Descargar PDF
            </span>
          </button>
        </div>
      </div>
    </template>
  </MobileScreen>
</template>
