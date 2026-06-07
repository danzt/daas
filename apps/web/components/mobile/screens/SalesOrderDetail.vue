<script setup lang="ts">
import {
  ShoppingBag,
  CheckCircle,
  XCircle,
  Receipt,
  Loader2,
  AlertTriangle,
  FileText,
} from "lucide-vue-next";

interface OrderLine {
  id: string;
  description: string;
  quantity: number;
  unit_price: number;
  subtotal: number;
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
  lines: OrderLine[];
}

const props = defineProps<{
  order: SaleOrder | null;
  loading: boolean;
  loadError: string;
  actionLoading: boolean;
  actionError: string;
}>();

const emit = defineEmits<{
  confirm: [];
  invoice: [];
  cancel: [];
  retry: [];
}>();

const confirmingCancel = ref(false);

const STATUS: Record<string, { label: string; chip: string; icon: Component }> =
  {
    draft: {
      label: "Borrador",
      chip: "bg-gray-100 text-gray-600",
      icon: ShoppingBag,
    },
    confirmed: {
      label: "Confirmada",
      chip: "bg-blue-100 text-blue-700",
      icon: CheckCircle,
    },
    invoiced: {
      label: "Facturada",
      chip: "bg-emerald-100 text-emerald-700",
      icon: Receipt,
    },
    cancelled: {
      label: "Cancelada",
      chip: "bg-rose-100 text-rose-600",
      icon: XCircle,
    },
  };

const ID_TYPE: Record<string, string> = {
  anonymous: "Anónimo",
  cedula: "Cédula",
  rif: "RIF",
  passport: "Pasaporte",
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
  () => props.order?.status === "draft" || props.order?.status === "confirmed",
);
</script>

<template>
  <MobileScreen title="Orden de venta" back subtitle="Detalle B2B">
    <!-- Loading -->
    <div v-if="loading" class="flex justify-center py-20">
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <!-- Error -->
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

    <template v-else-if="order">
      <!-- Status + total hero -->
      <div class="px-4 pt-4">
        <div class="rounded-2xl border border-border bg-white p-5">
          <span
            class="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-bold"
            :class="STATUS[order.status]?.chip"
          >
            <component :is="STATUS[order.status]?.icon" class="size-3.5" />
            {{ STATUS[order.status]?.label }}
          </span>
          <p class="mt-3 text-3xl font-black tabular-nums text-foreground">
            {{ fmtCurrency(order.total) }}
          </p>
          <p class="mt-1 text-sm font-semibold text-foreground">
            {{ order.customer_name || "Cliente anónimo" }}
          </p>
          <p
            v-if="
              order.customer_id_type !== 'anonymous' && order.customer_id_number
            "
            class="font-mono text-xs text-muted-foreground"
          >
            {{ ID_TYPE[order.customer_id_type] ?? order.customer_id_type }}:
            {{ order.customer_id_number }}
          </p>
        </div>
      </div>

      <!-- Action error -->
      <div v-if="actionError" class="px-4 pt-3">
        <div
          class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
        >
          {{ actionError }}
        </div>
      </div>

      <!-- Line items -->
      <MobileSectionHeader title="Productos" class="pt-5" />
      <div class="divide-y divide-border border-y border-border bg-white">
        <div
          v-for="line in order.lines"
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
        <div class="flex items-center justify-between px-4 py-3.5">
          <span class="text-sm font-bold text-foreground">Total</span>
          <span class="font-mono text-lg font-black tabular-nums text-primary">
            {{ fmtCurrency(order.total) }}
          </span>
        </div>
      </div>

      <!-- Meta -->
      <div class="space-y-2 px-4 pt-5 text-xs text-muted-foreground">
        <div class="flex justify-between">
          <span>Creada</span><span>{{ fmtDate(order.created_at) }}</span>
        </div>
        <div v-if="order.confirmed_at" class="flex justify-between">
          <span>Confirmada</span><span>{{ fmtDate(order.confirmed_at) }}</span>
        </div>
        <div v-if="order.invoiced_at" class="flex justify-between">
          <span>Facturada</span><span>{{ fmtDate(order.invoiced_at) }}</span>
        </div>
      </div>

      <!-- Notes -->
      <div v-if="order.notes" class="px-4 pt-4">
        <p class="rounded-2xl bg-muted/60 p-3.5 text-sm text-muted-foreground">
          {{ order.notes }}
        </p>
      </div>

      <!-- Link to invoice -->
      <div v-if="order.invoice_id" class="px-4 pt-4">
        <NuxtLink
          :to="`/invoices/${order.invoice_id}`"
          class="flex items-center justify-center gap-2 rounded-2xl border border-primary/30 bg-primary/5 py-3 text-sm font-bold text-primary active:bg-primary/10"
        >
          <FileText class="size-4" /> Ver factura
        </NuxtLink>
      </div>

      <!-- Sticky action bar -->
      <div
        v-if="hasActions"
        class="fixed inset-x-0 bottom-0 z-40 border-t border-border bg-white px-4 pb-[calc(0.75rem+env(safe-area-inset-bottom))] pt-3"
      >
        <!-- Cancel confirmation -->
        <div v-if="confirmingCancel" class="flex gap-2">
          <button
            type="button"
            :disabled="actionLoading"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-red-500 text-sm font-bold text-white active:bg-red-600 disabled:opacity-50"
            @click="emit('cancel')"
          >
            <Loader2 v-if="actionLoading" class="mx-auto size-5 animate-spin" />
            <span v-else>Sí, cancelar orden</span>
          </button>
          <button
            type="button"
            class="no-min-tap h-12 rounded-2xl border border-border px-5 text-sm font-bold text-muted-foreground active:bg-accent"
            @click="confirmingCancel = false"
          >
            No
          </button>
        </div>

        <!-- Primary actions -->
        <div v-else class="flex gap-2">
          <button
            type="button"
            class="no-min-tap h-12 w-12 shrink-0 rounded-2xl border border-red-200 text-red-500 active:bg-red-50"
            aria-label="Cancelar orden"
            @click="confirmingCancel = true"
          >
            <XCircle class="mx-auto size-5" />
          </button>
          <button
            v-if="order.status === 'draft'"
            type="button"
            :disabled="actionLoading"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-primary text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('confirm')"
          >
            <Loader2 v-if="actionLoading" class="mx-auto size-5 animate-spin" />
            <span v-else>Confirmar orden</span>
          </button>
          <button
            v-else-if="order.status === 'confirmed'"
            type="button"
            :disabled="actionLoading"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-cta text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('invoice')"
          >
            <Loader2 v-if="actionLoading" class="mx-auto size-5 animate-spin" />
            <span v-else>Generar factura</span>
          </button>
        </div>
      </div>
    </template>
  </MobileScreen>
</template>
