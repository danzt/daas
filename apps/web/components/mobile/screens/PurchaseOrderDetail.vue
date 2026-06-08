<script setup lang="ts">
import {
  ShoppingCart,
  CheckCircle,
  XCircle,
  Package,
  Loader2,
  AlertTriangle,
} from "lucide-vue-next";

interface POLine {
  id: string;
  description: string;
  quantity_ordered: number;
  unit_cost: number;
  subtotal: number;
}

interface Supplier {
  id: string;
  name: string;
  rif: string;
}

interface PurchaseOrder {
  id: string;
  status: "draft" | "ordered" | "received" | "cancelled";
  notes: string;
  total: number;
  ordered_at: string | null;
  received_at: string | null;
  created_at: string;
  lines: POLine[];
  supplier?: Supplier;
}

const props = defineProps<{
  po: PurchaseOrder | null;
  loading: boolean;
  loadError: string;
  actionLoading: boolean;
  actionError: string;
}>();

const emit = defineEmits<{
  order: [];
  receive: [];
  cancel: [];
  retry: [];
}>();

const confirmingCancel = ref(false);

const STATUS: Record<string, { label: string; chip: string; icon: Component }> =
  {
    draft: {
      label: "Borrador",
      chip: "bg-gray-100 text-gray-600",
      icon: ShoppingCart,
    },
    ordered: {
      label: "Ordenada",
      chip: "bg-blue-100 text-blue-700",
      icon: Package,
    },
    received: {
      label: "Recibida",
      chip: "bg-emerald-100 text-emerald-700",
      icon: CheckCircle,
    },
    cancelled: {
      label: "Cancelada",
      chip: "bg-rose-100 text-rose-600",
      icon: XCircle,
    },
  };

function fmtDate(d: string | null) {
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
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
}

const hasActionBar = computed(
  () => props.po && !["received", "cancelled"].includes(props.po.status),
);
</script>

<template>
  <MobileScreen title="Orden de compra" back subtitle="Proveedores">
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

    <template v-else-if="po">
      <!-- Hero -->
      <div class="px-4 pt-4">
        <div class="rounded-2xl border border-border bg-white p-5 space-y-3">
          <div class="flex items-center justify-between gap-2">
            <span
              class="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-bold"
              :class="STATUS[po.status]?.chip"
            >
              <component :is="STATUS[po.status]?.icon" class="size-3.5" />
              {{ STATUS[po.status]?.label }}
            </span>
            <span class="font-mono text-xs text-muted-foreground">
              #{{ po.id.slice(0, 8) }}
            </span>
          </div>

          <div>
            <p class="text-2xl font-black tabular-nums text-foreground">
              {{ fmtCurrency(po.total) }}
            </p>
            <p class="mt-1 text-sm font-semibold text-foreground">
              {{ po.supplier?.name ?? "Proveedor desconocido" }}
            </p>
            <p
              v-if="po.supplier?.rif"
              class="font-mono text-xs text-muted-foreground"
            >
              RIF: {{ po.supplier.rif }}
            </p>
          </div>

          <div
            class="grid grid-cols-3 gap-2 pt-2 border-t border-border text-xs"
          >
            <div>
              <p class="text-muted-foreground">Creada</p>
              <p class="font-medium text-foreground">
                {{ fmtDate(po.created_at) }}
              </p>
            </div>
            <div>
              <p class="text-muted-foreground">Ordenada</p>
              <p class="font-medium text-foreground">
                {{ fmtDate(po.ordered_at) }}
              </p>
            </div>
            <div>
              <p class="text-muted-foreground">Recibida</p>
              <p class="font-medium text-foreground">
                {{ fmtDate(po.received_at) }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Notes -->
      <div v-if="po.notes" class="px-4 pt-4">
        <p class="rounded-2xl bg-muted/60 p-3.5 text-sm text-muted-foreground">
          {{ po.notes }}
        </p>
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
      <MobileSectionHeader
        :title="`Líneas (${po.lines?.length ?? 0})`"
        class="pt-5"
      />
      <div class="divide-y divide-border border-y border-border bg-white">
        <div
          v-for="line in po.lines"
          :key="line.id"
          class="flex items-start gap-3 px-4 py-3"
        >
          <span
            class="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-lg bg-muted text-xs font-bold text-muted-foreground"
          >
            {{ line.quantity_ordered }}×
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-foreground">
              {{ line.description }}
            </p>
            <p class="text-[11px] text-muted-foreground">
              {{ fmtCurrency(line.unit_cost) }} c/u
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
            {{ fmtCurrency(po.total) }}
          </span>
        </div>
      </div>

      <!-- Received confirmation -->
      <div v-if="po.status === 'received'" class="px-4 pt-4">
        <p class="flex items-center gap-2 text-sm text-emerald-600">
          <CheckCircle class="size-4" />
          Recibida el {{ fmtDate(po.received_at) }} — inventario actualizado
        </p>
      </div>

      <div v-if="po.status === 'cancelled'" class="px-4 pt-4">
        <p class="flex items-center gap-2 text-sm text-rose-600">
          <XCircle class="size-4" />
          Orden cancelada
        </p>
      </div>

      <!-- Spacer -->
      <div v-if="hasActionBar" class="h-24" />
    </template>

    <!-- Action bar -->
    <template v-if="hasActionBar && po">
      <div
        class="fixed inset-x-0 bottom-0 z-40 border-t border-border bg-white px-4 pb-[calc(0.75rem+env(safe-area-inset-bottom))] pt-3"
      >
        <!-- Confirming cancel -->
        <div v-if="confirmingCancel" class="flex gap-2">
          <button
            type="button"
            :disabled="actionLoading"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-red-500 text-sm font-bold text-white active:bg-red-600 disabled:opacity-50"
            @click="emit('cancel')"
          >
            <Loader2 v-if="actionLoading" class="mx-auto size-5 animate-spin" />
            <span v-else>Sí, cancelar</span>
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
          <!-- Cancel icon -->
          <button
            type="button"
            class="no-min-tap h-12 w-12 shrink-0 rounded-2xl border border-rose-200 text-rose-500 active:bg-rose-50"
            aria-label="Cancelar orden"
            @click="confirmingCancel = true"
          >
            <XCircle class="mx-auto size-5" />
          </button>

          <!-- draft → order -->
          <button
            v-if="po.status === 'draft'"
            type="button"
            :disabled="actionLoading"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-blue-600 text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('order')"
          >
            <Loader2 v-if="actionLoading" class="mx-auto size-5 animate-spin" />
            <span v-else class="flex items-center justify-center gap-2">
              <Package class="size-4" /> Confirmar orden
            </span>
          </button>

          <!-- ordered → receive -->
          <button
            v-else-if="po.status === 'ordered'"
            type="button"
            :disabled="actionLoading"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-emerald-600 text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('receive')"
          >
            <Loader2 v-if="actionLoading" class="mx-auto size-5 animate-spin" />
            <span v-else class="flex items-center justify-center gap-2">
              <CheckCircle class="size-4" /> Marcar recibida
            </span>
          </button>
        </div>
      </div>
    </template>
  </MobileScreen>
</template>
