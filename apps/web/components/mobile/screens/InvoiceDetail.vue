<script setup lang="ts">
import {
  Clock,
  CheckCircle,
  XCircle,
  Download,
  Loader2,
  AlertTriangle,
} from "lucide-vue-next";
import type { Invoice } from "~/components/invoices/InvoiceFormModal.vue";

const props = defineProps<{
  invoice: Invoice | null;
  loading: boolean;
  loadError: string;
  issuing: boolean;
  cancelling: boolean;
  actionError: string;
}>();

const emit = defineEmits<{
  issue: [];
  cancel: [];
  downloadPdf: [];
  retry: [];
}>();

const confirmingCancel = ref(false);

const STATUS: Record<string, { label: string; chip: string; icon: Component }> =
  {
    draft: {
      label: "Borrador",
      chip: "bg-gray-100 text-gray-600",
      icon: Clock,
    },
    issued: {
      label: "Emitida",
      chip: "bg-emerald-100 text-emerald-700",
      icon: CheckCircle,
    },
    cancelled: {
      label: "Anulada",
      chip: "bg-rose-100 text-rose-600",
      icon: XCircle,
    },
  };

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
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const hasActions = computed(
  () => props.invoice?.status === "draft" || props.invoice?.status === "issued",
);
</script>

<template>
  <MobileScreen title="Factura" back subtitle="Facturación interna">
    <div v-if="loading" class="flex justify-center py-20">
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <div v-else-if="loadError" class="px-4 py-12 text-center">
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

    <template v-else-if="invoice">
      <!-- Hero -->
      <div class="px-4 pt-4">
        <div class="rounded-2xl border border-border bg-white p-5">
          <div class="flex items-center justify-between">
            <span
              class="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-bold"
              :class="STATUS[invoice.status]?.chip"
            >
              <component :is="STATUS[invoice.status]?.icon" class="size-3.5" />
              {{ STATUS[invoice.status]?.label }}
            </span>
            <span
              v-if="invoice.correlative"
              class="font-mono text-sm font-bold text-muted-foreground"
              >N° {{ invoice.correlative }}</span
            >
          </div>
          <p class="mt-3 text-3xl font-black tabular-nums text-foreground">
            {{ fmtCurrency(invoice.total) }}
          </p>
          <p class="mt-1 text-sm font-semibold text-foreground">
            {{ invoice.customer_name || "Consumidor Final" }}
          </p>
          <p
            v-if="invoice.customer_id_number"
            class="font-mono text-xs text-muted-foreground"
          >
            {{ invoice.customer_id_type }}: {{ invoice.customer_id_number }}
          </p>
        </div>
      </div>

      <div v-if="actionError" class="px-4 pt-3">
        <div
          class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
        >
          {{ actionError }}
        </div>
      </div>

      <!-- Lines -->
      <MobileSectionHeader title="Detalle" class="pt-5" />
      <div class="divide-y divide-border border-y border-border bg-white">
        <div
          v-for="line in invoice.lines"
          :key="line.id"
          class="flex items-start gap-3 px-4 py-3"
        >
          <span
            class="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-lg bg-muted text-xs font-bold text-muted-foreground"
            >{{ line.quantity }}×</span
          >
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-foreground">
              {{ line.description }}
            </p>
            <p class="text-[11px] text-muted-foreground">
              {{ fmtCurrency(line.unit_price) }} c/u
            </p>
          </div>
          <span
            class="font-mono text-sm font-bold tabular-nums text-foreground"
          >
            {{ fmtCurrency(line.subtotal) }}
          </span>
        </div>
        <div class="flex items-center justify-between px-4 py-2.5">
          <span class="text-sm text-muted-foreground">Subtotal</span>
          <span class="font-mono text-sm tabular-nums text-foreground">{{
            fmtCurrency(invoice.subtotal)
          }}</span>
        </div>
        <div class="flex items-center justify-between px-4 py-3.5">
          <span class="text-sm font-bold text-foreground">Total</span>
          <span
            class="font-mono text-lg font-black tabular-nums text-primary"
            >{{ fmtCurrency(invoice.total) }}</span
          >
        </div>
      </div>

      <!-- Meta -->
      <div class="space-y-2 px-4 pt-5 text-xs text-muted-foreground">
        <div class="flex justify-between">
          <span>Creada</span><span>{{ fmtDate(invoice.created_at) }}</span>
        </div>
        <div v-if="invoice.issued_at" class="flex justify-between">
          <span>Emitida</span><span>{{ fmtDate(invoice.issued_at) }}</span>
        </div>
      </div>

      <div v-if="invoice.notes" class="px-4 pt-4">
        <p class="rounded-2xl bg-muted/60 p-3.5 text-sm text-muted-foreground">
          {{ invoice.notes }}
        </p>
      </div>

      <!-- Action bar -->
      <div
        v-if="hasActions"
        class="fixed inset-x-0 bottom-0 z-40 border-t border-border bg-white px-4 pb-[calc(0.75rem+env(safe-area-inset-bottom))] pt-3"
      >
        <div v-if="confirmingCancel" class="flex gap-2">
          <button
            type="button"
            :disabled="cancelling"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-red-500 text-sm font-bold text-white active:bg-red-600 disabled:opacity-50"
            @click="emit('cancel')"
          >
            <Loader2 v-if="cancelling" class="mx-auto size-5 animate-spin" />
            <span v-else>Sí, anular factura</span>
          </button>
          <button
            type="button"
            class="no-min-tap h-12 rounded-2xl border border-border px-5 text-sm font-bold text-muted-foreground active:bg-accent"
            @click="confirmingCancel = false"
          >
            No
          </button>
        </div>

        <div v-else class="flex gap-2">
          <button
            type="button"
            class="no-min-tap h-12 w-12 shrink-0 rounded-2xl border border-red-200 text-red-500 active:bg-red-50"
            aria-label="Anular factura"
            @click="confirmingCancel = true"
          >
            <XCircle class="mx-auto size-5" />
          </button>
          <button
            v-if="invoice.status === 'draft'"
            type="button"
            :disabled="issuing"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-primary text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('issue')"
          >
            <Loader2 v-if="issuing" class="mx-auto size-5 animate-spin" />
            <span v-else>Emitir factura</span>
          </button>
          <button
            v-else
            type="button"
            class="no-min-tap flex h-12 flex-1 items-center justify-center gap-2 rounded-2xl bg-primary text-sm font-bold text-white active:opacity-90"
            @click="emit('downloadPdf')"
          >
            <Download class="size-5" /> Descargar PDF
          </button>
        </div>
      </div>
    </template>
  </MobileScreen>
</template>
