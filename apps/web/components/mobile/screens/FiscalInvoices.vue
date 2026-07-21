<script setup lang="ts">
import {
  Clock,
  CheckCircle,
  XCircle,
  AlertTriangle,
  RefreshCw,
  Loader2,
  FolderOpen,
  FileText,
} from "lucide-vue-next";
import type { FiscalInvoice } from "~/components/invoices/FiscalInvoiceFormModal.vue";

const props = defineProps<{
  invoices: FiscalInvoice[];
  loading: boolean;
  loadError: string;
  retryingIds: string[];
}>();

const emit = defineEmits<{
  create: [];
  retry: [];
  retryInvoice: [inv: FiscalInvoice];
}>();

const search = ref("");
const statusFilter = ref("all");

const STATUS: Record<
  string,
  { label: string; chip: string; bubble: string; fg: string; icon: Component }
> = {
  draft: {
    label: "Borrador",
    chip: "bg-gray-100 text-gray-600",
    bubble: "bg-gray-100",
    fg: "text-gray-500",
    icon: FileText,
  },
  pending_fiscal: {
    label: "Procesando",
    chip: "bg-amber-50 text-amber-700",
    bubble: "bg-amber-50",
    fg: "text-amber-600",
    icon: Clock,
  },
  issued: {
    label: "Emitida",
    chip: "bg-emerald-100 text-emerald-700",
    bubble: "bg-emerald-50",
    fg: "text-emerald-600",
    icon: CheckCircle,
  },
  failed: {
    label: "Falló",
    chip: "bg-red-100 text-red-700",
    bubble: "bg-red-50",
    fg: "text-red-500",
    icon: AlertTriangle,
  },
  cancelled: {
    label: "Anulada",
    chip: "bg-rose-100 text-rose-600",
    bubble: "bg-rose-50",
    fg: "text-rose-500",
    icon: XCircle,
  },
};

const segOptions = computed(() => [
  { key: "all", label: "Todas", count: props.invoices.length },
  {
    key: "issued",
    label: "Emitidas",
    count: props.invoices.filter((i) => i.status === "issued").length,
  },
  {
    key: "pending_fiscal",
    label: "Procesando",
    count: props.invoices.filter((i) => i.status === "pending_fiscal").length,
  },
  {
    key: "failed",
    label: "Fallidas",
    count: props.invoices.filter((i) => i.status === "failed").length,
  },
]);

const filtered = computed(() => {
  let list = props.invoices;
  if (statusFilter.value !== "all")
    list = list.filter((i) => i.status === statusFilter.value);
  const q = search.value.trim().toLowerCase();
  if (q)
    list = list.filter(
      (i) =>
        (i.customer_name ?? "").toLowerCase().includes(q) ||
        (i.fiscal_number ?? "").toLowerCase().includes(q),
    );
  return list;
});

function fmtCurrency(n: number) {
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
}
function fmtDate(d: string) {
  return new Date(d).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "short",
  });
}
</script>

<template>
  <MobileScreen
    title="Facturas fiscales"
    back
    subtitle="SENIAT / máquina fiscal"
  >
    <MobileSearchBar v-model="search" placeholder="Buscar por cliente o N°…" />
    <MobileSegment v-model="statusFilter" :options="segOptions" />

    <div
      v-if="loadError"
      class="mx-4 mt-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
    >
      {{ loadError }}
    </div>

    <div
      v-else-if="loading && !invoices.length"
      class="flex justify-center py-16"
    >
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <div v-else-if="!filtered.length" class="px-6 py-16 text-center">
      <FolderOpen class="mx-auto mb-4 size-12 text-muted-foreground/20" />
      <p class="font-medium text-muted-foreground">Sin facturas fiscales</p>
    </div>

    <div v-else class="divide-y divide-border border-t border-border bg-white">
      <div v-for="inv in filtered" :key="inv.id">
        <MobileListItem
          :to="`/invoices/fiscal/${inv.id}`"
          :chevron="inv.status !== 'failed'"
        >
          <template #leading>
            <div
              class="flex size-11 items-center justify-center rounded-2xl"
              :class="STATUS[inv.status]?.bubble"
            >
              <component
                :is="STATUS[inv.status]?.icon ?? FileText"
                class="size-5"
                :class="STATUS[inv.status]?.fg"
              />
            </div>
          </template>
          <p
            class="truncate text-sm font-semibold leading-tight text-foreground"
          >
            {{ inv.customer_name || "Consumidor Final" }}
          </p>
          <div class="mt-1 flex items-center gap-1.5">
            <span
              v-if="inv.fiscal_number"
              class="rounded bg-muted px-1.5 py-px font-mono text-[10px] text-muted-foreground"
              >N° {{ inv.fiscal_number }}</span
            >
            <span
              class="rounded px-1.5 py-px text-[10px] font-bold"
              :class="STATUS[inv.status]?.chip"
              >{{ STATUS[inv.status]?.label }}</span
            >
            <span class="text-[11px] text-muted-foreground">{{
              fmtDate(inv.created_at)
            }}</span>
          </div>
          <template #trailing>
            <span
              class="font-mono text-sm font-bold tabular-nums text-foreground"
            >
              {{ fmtCurrency(inv.total) }}
            </span>
          </template>
        </MobileListItem>

        <!-- Retry row for failed invoices -->
        <div
          v-if="inv.status === 'failed'"
          class="flex items-center gap-2 bg-red-50/60 px-4 pb-3"
        >
          <p class="min-w-0 flex-1 truncate text-[11px] text-red-500">
            {{ inv.fail_reason || "Error al transmitir al SENIAT" }}
          </p>
          <button
            type="button"
            :disabled="retryingIds.includes(inv.id)"
            class="no-min-tap flex h-9 shrink-0 items-center gap-1 rounded-xl bg-red-500 px-3 text-xs font-bold text-white active:bg-red-600 disabled:opacity-50"
            @click="emit('retryInvoice', inv)"
          >
            <RefreshCw
              :class="[
                'size-3.5',
                retryingIds.includes(inv.id) ? 'animate-spin' : '',
              ]"
            />
            Reintentar
          </button>
        </div>
      </div>
    </div>

    <MobileFab label="Nueva" @click="emit('create')" />
  </MobileScreen>
</template>
