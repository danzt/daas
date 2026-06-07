<script setup lang="ts">
import {
  ArrowLeft,
  Building2,
  Phone,
  Mail,
  MapPin,
  FileText,
  Pencil,
  ShoppingCart,
  Loader2,
  ChevronRight,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import SupplierFormModal from "~/components/suppliers/SupplierFormModal.vue";
import type { Supplier } from "~/components/suppliers/SupplierFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

interface POLine {
  id: string;
  product_id: string;
  description: string;
  quantity_ordered: number;
  unit_cost: number;
  subtotal: number;
}

interface PurchaseOrder {
  id: string;
  supplier_id: string;
  status: "draft" | "ordered" | "received" | "cancelled";
  notes: string;
  total: number;
  ordered_at: string | null;
  received_at: string | null;
  created_at: string;
  lines: POLine[];
}

const route = useRoute();
const { isMobile } = useMobileMode();
const supplierId = route.params.id as string;

const supplier = ref<Supplier | null>(null);
const orders = ref<PurchaseOrder[]>([]);
const loading = ref(false);
const loadError = ref("");
const formModalOpen = ref(false);

const STATUS_CONFIG: Record<string, { label: string; class: string }> = {
  draft: { label: "Borrador", class: "bg-muted text-muted-foreground" },
  ordered: {
    label: "Ordenada",
    class: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
  },
  received: {
    label: "Recibida",
    class:
      "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400",
  },
  cancelled: {
    label: "Cancelada",
    class: "bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400",
  },
};

function fmtDate(d: string | null) {
  if (!d) return "—";
  return new Date(d).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

function fmtCurrency(n: number) {
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
}

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    const [sup, pos] = await Promise.all([
      useApiFetch<Supplier>(`/api/v1/suppliers/${supplierId}`),
      useApiFetch<PurchaseOrder[]>(
        `/api/v1/purchase-orders?supplier_id=${supplierId}`,
      ),
    ]);
    supplier.value = sup;
    orders.value = pos;
  } catch {
    loadError.value = "No se pudo cargar la información del proveedor";
  } finally {
    loading.value = false;
  }
}

function onSaved(updated: Supplier) {
  supplier.value = updated;
  formModalOpen.value = false;
}

onMounted(load);
</script>

<template>
  <!-- MOBILE -->
  <MobileScreensSupplierDetail
    v-if="isMobile"
    :supplier="supplier"
    :orders="orders"
    :loading="loading"
    :load-error="loadError"
    @edit="formModalOpen = true"
    @retry="load"
  />

  <!-- WEB -->
  <div v-else class="p-4 sm:p-6 space-y-6">
    <!-- Back -->
    <div class="flex items-center gap-3">
      <NuxtLink
        to="/suppliers"
        class="p-2 rounded-lg hover:bg-muted transition-colors text-muted-foreground"
      >
        <ArrowLeft class="w-4 h-4" />
      </NuxtLink>
      <h1 class="text-xl font-bold text-foreground">
        {{ supplier?.name ?? "Proveedor" }}
      </h1>
    </div>

    <!-- Error -->
    <div
      v-if="loadError"
      class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3"
    >
      {{ loadError }}
    </div>

    <!-- Loading -->
    <div
      v-if="loading"
      class="flex items-center justify-center py-20 text-muted-foreground"
    >
      <Loader2 class="w-6 h-6 animate-spin mr-2" />
      Cargando...
    </div>

    <template v-else-if="supplier">
      <!-- Info card -->
      <div class="border bg-card rounded-xl p-6">
        <div class="flex items-start justify-between gap-4">
          <div class="flex items-center gap-4">
            <div
              class="w-14 h-14 rounded-xl bg-primary/10 flex items-center justify-center shrink-0"
            >
              <Building2 class="w-7 h-7 text-primary" />
            </div>
            <div>
              <h2 class="text-xl font-semibold text-foreground">
                {{ supplier.name }}
              </h2>
              <p
                v-if="supplier.rif"
                class="text-sm text-muted-foreground font-mono mt-0.5"
              >
                RIF: {{ supplier.rif }}
              </p>
              <span
                :class="[
                  'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium mt-1',
                  supplier.active
                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                    : 'bg-muted text-muted-foreground',
                ]"
              >
                {{ supplier.active ? "Activo" : "Inactivo" }}
              </span>
            </div>
          </div>
          <button
            class="flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg border border-border hover:bg-muted transition-colors text-foreground"
            @click="formModalOpen = true"
          >
            <Pencil class="w-3.5 h-3.5" />
            Editar
          </button>
        </div>

        <!-- Contact info -->
        <div class="mt-6 grid grid-cols-1 md:grid-cols-3 gap-4">
          <div
            v-if="supplier.contact_name"
            class="flex items-center gap-2 text-sm text-muted-foreground"
          >
            <Building2 class="w-4 h-4 shrink-0" />
            {{ supplier.contact_name }}
          </div>
          <div
            v-if="supplier.phone"
            class="flex items-center gap-2 text-sm text-muted-foreground"
          >
            <Phone class="w-4 h-4 shrink-0" />
            {{ supplier.phone }}
          </div>
          <div
            v-if="supplier.email"
            class="flex items-center gap-2 text-sm text-muted-foreground"
          >
            <Mail class="w-4 h-4 shrink-0" />
            {{ supplier.email }}
          </div>
          <div
            v-if="supplier.address"
            class="flex items-center gap-2 text-sm text-muted-foreground md:col-span-2"
          >
            <MapPin class="w-4 h-4 shrink-0" />
            {{ supplier.address }}
          </div>
        </div>

        <div
          v-if="supplier.notes"
          class="mt-4 p-3 bg-muted/40 rounded-lg text-sm text-muted-foreground"
        >
          <FileText class="w-4 h-4 inline mr-1" />
          {{ supplier.notes }}
        </div>
      </div>

      <!-- Purchase orders -->
      <div>
        <div class="flex items-center justify-between mb-3">
          <h3
            class="text-base font-semibold text-foreground flex items-center gap-2"
          >
            <ShoppingCart class="w-4 h-4 text-primary" />
            Órdenes de compra ({{ orders.length }})
          </h3>
          <NuxtLink
            to="/suppliers/purchase-orders"
            class="text-sm text-primary hover:underline flex items-center gap-1"
          >
            Ver todas <ChevronRight class="w-3.5 h-3.5" />
          </NuxtLink>
        </div>

        <div
          v-if="orders.length === 0"
          class="border bg-card rounded-xl flex flex-col items-center justify-center py-12 text-muted-foreground gap-2"
        >
          <ShoppingCart class="w-10 h-10 opacity-30" />
          <p class="text-sm">Sin órdenes de compra para este proveedor</p>
          <NuxtLink
            to="/suppliers/purchase-orders"
            class="text-sm text-primary hover:underline"
          >
            Crear una orden
          </NuxtLink>
        </div>

        <div v-else class="border bg-card rounded-xl overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-border bg-muted/30">
                  <th
                    class="px-4 py-3 text-left font-medium text-muted-foreground"
                  >
                    Estado
                  </th>
                  <th
                    class="px-4 py-3 text-left font-medium text-muted-foreground hidden md:table-cell"
                  >
                    Fecha
                  </th>
                  <th
                    class="px-4 py-3 text-right font-medium text-muted-foreground"
                  >
                    Total
                  </th>
                  <th class="px-4 py-3" />
                </tr>
              </thead>
              <tbody class="divide-y divide-border">
                <tr
                  v-for="po in orders"
                  :key="po.id"
                  class="hover:bg-muted/20 transition-colors group"
                >
                  <td class="px-4 py-3">
                    <span
                      :class="[
                        'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                        STATUS_CONFIG[po.status]?.class,
                      ]"
                    >
                      {{ STATUS_CONFIG[po.status]?.label }}
                    </span>
                  </td>
                  <td
                    class="px-4 py-3 text-muted-foreground hidden md:table-cell"
                  >
                    {{ fmtDate(po.created_at) }}
                  </td>
                  <td
                    class="px-4 py-3 text-right font-mono font-medium text-foreground"
                  >
                    {{ fmtCurrency(po.total) }}
                  </td>
                  <td class="px-4 py-3 text-right">
                    <NuxtLink
                      :to="`/suppliers/purchase-orders/${po.id}`"
                      class="opacity-0 group-hover:opacity-100 p-1.5 rounded-lg hover:bg-muted transition-all text-muted-foreground inline-flex"
                    >
                      <ChevronRight class="w-4 h-4" />
                    </NuxtLink>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </template>

    <!-- Edit modal -->
    <SupplierFormModal
      :open="formModalOpen"
      :supplier="supplier"
      @close="formModalOpen = false"
      @saved="onSaved"
    />
  </div>
</template>
