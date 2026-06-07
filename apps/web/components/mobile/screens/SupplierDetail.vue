<script setup lang="ts">
import {
  Pencil,
  Phone,
  Mail,
  MapPin,
  User,
  Loader2,
  AlertTriangle,
  ClipboardList,
} from "lucide-vue-next";
import type { Supplier } from "~/components/suppliers/SupplierFormModal.vue";

interface PurchaseOrder {
  id: string;
  status: "draft" | "ordered" | "received" | "cancelled";
  total: number;
  created_at: string;
}

const props = defineProps<{
  supplier: Supplier | null;
  orders: PurchaseOrder[];
  loading: boolean;
  loadError: string;
}>();

const emit = defineEmits<{ edit: []; retry: [] }>();

const PO_STATUS: Record<string, { label: string; chip: string }> = {
  draft: { label: "Borrador", chip: "bg-gray-100 text-gray-600" },
  ordered: { label: "Ordenada", chip: "bg-blue-100 text-blue-700" },
  received: { label: "Recibida", chip: "bg-emerald-100 text-emerald-700" },
  cancelled: { label: "Cancelada", chip: "bg-rose-100 text-rose-600" },
};

const initials = computed(() =>
  (props.supplier?.name || "?")
    .split(" ")
    .slice(0, 2)
    .map((w) => w[0])
    .join("")
    .toUpperCase(),
);

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
    year: "numeric",
  });
}
</script>

<template>
  <MobileScreen title="Proveedor" back subtitle="Detalle">
    <template #action>
      <button
        type="button"
        aria-label="Editar"
        class="no-min-tap flex size-10 items-center justify-center rounded-full text-primary active:bg-accent"
        @click="emit('edit')"
      >
        <Pencil class="size-5" />
      </button>
    </template>

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

    <template v-else-if="supplier">
      <!-- Hero -->
      <div class="px-4 pt-4">
        <div
          class="flex items-center gap-3 rounded-2xl border border-border bg-white p-4"
        >
          <div
            class="flex size-14 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-lg font-black text-primary"
          >
            {{ initials }}
          </div>
          <div class="min-w-0 flex-1">
            <p class="truncate text-base font-bold text-foreground">
              {{ supplier.name }}
            </p>
            <p
              v-if="supplier.rif"
              class="font-mono text-xs text-muted-foreground"
            >
              {{ supplier.rif }}
            </p>
            <span
              class="mt-1 inline-block rounded-full px-2 py-px text-[10px] font-bold"
              :class="
                supplier.active
                  ? 'bg-emerald-100 text-emerald-700'
                  : 'bg-muted text-muted-foreground'
              "
              >{{ supplier.active ? "Activo" : "Inactivo" }}</span
            >
          </div>
        </div>
      </div>

      <!-- Contact -->
      <MobileSectionHeader title="Contacto" class="pt-5" />
      <div class="divide-y divide-border border-y border-border bg-white">
        <div
          v-if="supplier.contact_name"
          class="flex items-center gap-3 px-4 py-3"
        >
          <User class="size-5 shrink-0 text-muted-foreground" />
          <span class="text-sm text-foreground">{{
            supplier.contact_name
          }}</span>
        </div>
        <a
          v-if="supplier.phone"
          :href="`tel:${supplier.phone}`"
          class="flex items-center gap-3 px-4 py-3 active:bg-gray-50"
        >
          <Phone class="size-5 shrink-0 text-primary" />
          <span class="text-sm font-medium text-primary">{{
            supplier.phone
          }}</span>
        </a>
        <a
          v-if="supplier.email"
          :href="`mailto:${supplier.email}`"
          class="flex items-center gap-3 px-4 py-3 active:bg-gray-50"
        >
          <Mail class="size-5 shrink-0 text-primary" />
          <span class="truncate text-sm font-medium text-primary">{{
            supplier.email
          }}</span>
        </a>
        <div v-if="supplier.address" class="flex items-start gap-3 px-4 py-3">
          <MapPin class="mt-0.5 size-5 shrink-0 text-muted-foreground" />
          <span class="text-sm text-foreground">{{ supplier.address }}</span>
        </div>
      </div>

      <!-- Purchase orders -->
      <MobileSectionHeader
        title="Órdenes de compra"
        to="/suppliers/purchase-orders"
        class="pt-5"
      />
      <div
        v-if="orders.length"
        class="divide-y divide-border border-y border-border bg-white"
      >
        <MobileListItem
          v-for="po in orders"
          :key="po.id"
          :to="`/suppliers/purchase-orders/${po.id}`"
        >
          <template #leading>
            <div
              class="flex size-10 items-center justify-center rounded-xl bg-muted text-muted-foreground"
            >
              <ClipboardList class="size-5" />
            </div>
          </template>
          <p class="text-sm font-semibold text-foreground">
            Orden #{{ po.id.slice(0, 8) }}
          </p>
          <div class="mt-1 flex items-center gap-1.5">
            <span
              class="rounded px-1.5 py-px text-[10px] font-bold"
              :class="PO_STATUS[po.status]?.chip"
              >{{ PO_STATUS[po.status]?.label }}</span
            >
            <span class="text-[11px] text-muted-foreground">{{
              fmtDate(po.created_at)
            }}</span>
          </div>
          <template #trailing>
            <span class="font-mono text-sm font-bold tabular-nums">{{
              fmtCurrency(po.total)
            }}</span>
          </template>
        </MobileListItem>
      </div>
      <p v-else class="px-4 py-8 text-center text-sm text-muted-foreground">
        Sin órdenes de compra
      </p>
    </template>
  </MobileScreen>
</template>
