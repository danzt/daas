<script setup lang="ts">
import {
  ClipboardList,
  FileEdit,
  Truck,
  CheckCircle,
  XCircle,
  Loader2,
  FolderOpen,
} from "lucide-vue-next";
import type { Supplier } from "~/components/suppliers/SupplierFormModal.vue";

interface PurchaseOrder {
  id: string;
  supplier_id: string;
  status: "draft" | "ordered" | "received" | "cancelled";
  total: number;
  created_at: string;
  supplier?: Supplier;
}

const props = defineProps<{
  orders: PurchaseOrder[];
  loading: boolean;
  loadError: string;
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
    icon: FileEdit,
  },
  ordered: {
    label: "Ordenada",
    chip: "bg-blue-100 text-blue-700",
    bubble: "bg-blue-50",
    fg: "text-blue-600",
    icon: Truck,
  },
  received: {
    label: "Recibida",
    chip: "bg-emerald-100 text-emerald-700",
    bubble: "bg-emerald-50",
    fg: "text-emerald-600",
    icon: CheckCircle,
  },
  cancelled: {
    label: "Cancelada",
    chip: "bg-rose-100 text-rose-600",
    bubble: "bg-rose-50",
    fg: "text-rose-500",
    icon: XCircle,
  },
};

const segOptions = computed(() => [
  { key: "all", label: "Todas", count: props.orders.length },
  {
    key: "draft",
    label: "Borradores",
    count: props.orders.filter((o) => o.status === "draft").length,
  },
  {
    key: "ordered",
    label: "Ordenadas",
    count: props.orders.filter((o) => o.status === "ordered").length,
  },
  {
    key: "received",
    label: "Recibidas",
    count: props.orders.filter((o) => o.status === "received").length,
  },
]);

const filtered = computed(() => {
  let list = props.orders;
  if (statusFilter.value !== "all")
    list = list.filter((o) => o.status === statusFilter.value);
  const q = search.value.trim().toLowerCase();
  if (q)
    list = list.filter((o) =>
      (o.supplier?.name ?? "").toLowerCase().includes(q),
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
  <MobileScreen title="Órdenes de compra" back subtitle="Abastecimiento">
    <MobileSearchBar v-model="search" placeholder="Buscar por proveedor…" />
    <MobileSegment v-model="statusFilter" :options="segOptions" />

    <div
      v-if="loadError"
      class="mx-4 mt-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
    >
      {{ loadError }}
    </div>

    <div
      v-else-if="loading && !orders.length"
      class="flex justify-center py-16"
    >
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <div v-else-if="!filtered.length" class="px-6 py-16 text-center">
      <FolderOpen class="mx-auto mb-4 size-12 text-muted-foreground/20" />
      <p class="font-medium text-muted-foreground">Sin órdenes de compra</p>
    </div>

    <div v-else class="divide-y divide-border border-t border-border bg-white">
      <MobileListItem
        v-for="po in filtered"
        :key="po.id"
        :to="`/suppliers/purchase-orders/${po.id}`"
      >
        <template #leading>
          <div
            class="flex size-11 items-center justify-center rounded-2xl"
            :class="STATUS[po.status]?.bubble"
          >
            <component
              :is="STATUS[po.status]?.icon ?? ClipboardList"
              class="size-5"
              :class="STATUS[po.status]?.fg"
            />
          </div>
        </template>
        <p class="truncate text-sm font-semibold leading-tight text-foreground">
          {{ po.supplier?.name ?? "Proveedor" }}
        </p>
        <div class="mt-1 flex items-center gap-1.5">
          <span
            class="rounded px-1.5 py-px text-[10px] font-bold"
            :class="STATUS[po.status]?.chip"
            >{{ STATUS[po.status]?.label }}</span
          >
          <span class="text-[11px] text-muted-foreground">{{
            fmtDate(po.created_at)
          }}</span>
        </div>
        <template #trailing>
          <span
            class="font-mono text-sm font-bold tabular-nums text-foreground"
          >
            {{ fmtCurrency(po.total) }}
          </span>
        </template>
      </MobileListItem>
    </div>

    <MobileFab label="Nueva" @click="emit('create')" />
  </MobileScreen>
</template>
