<script setup lang="ts">
import {
  FileText,
  Clock,
  CheckCircle,
  XCircle,
  FolderOpen,
  Loader2,
} from "lucide-vue-next";
import type { Invoice } from "~/components/invoices/InvoiceFormModal.vue";

const props = defineProps<{
  invoices: Invoice[];
  loading: boolean;
  loadError: string;
  isOwner: boolean;
}>();

const emit = defineEmits<{ create: []; retry: [] }>();

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
    icon: Clock,
  },
  issued: {
    label: "Emitida",
    chip: "bg-emerald-100 text-emerald-700",
    bubble: "bg-emerald-50",
    fg: "text-emerald-600",
    icon: CheckCircle,
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
    key: "draft",
    label: "Borradores",
    count: props.invoices.filter((i) => i.status === "draft").length,
  },
  {
    key: "issued",
    label: "Emitidas",
    count: props.invoices.filter((i) => i.status === "issued").length,
  },
  {
    key: "cancelled",
    label: "Anuladas",
    count: props.invoices.filter((i) => i.status === "cancelled").length,
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
        (i.correlative ?? "").toLowerCase().includes(q),
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
  <MobileScreen title="Facturas" subtitle="Facturación interna">
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
      <p class="font-medium text-muted-foreground">Sin facturas</p>
    </div>

    <div v-else class="divide-y divide-border border-t border-border bg-white">
      <MobileListItem
        v-for="inv in filtered"
        :key="inv.id"
        :to="`/invoices/${inv.id}`"
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

        <p class="truncate text-sm font-semibold leading-tight text-foreground">
          {{ inv.customer_name || "Consumidor Final" }}
        </p>
        <div class="mt-1 flex items-center gap-1.5">
          <span
            v-if="inv.correlative"
            class="rounded bg-muted px-1.5 py-px font-mono text-[10px] text-muted-foreground"
            >N° {{ inv.correlative }}</span
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
    </div>

    <MobileFab v-if="isOwner" label="Nueva" @click="emit('create')" />
  </MobileScreen>
</template>
