<script setup lang="ts">
import {
  Smartphone,
  Building2,
  DollarSign,
  Bitcoin,
  Banknote,
  MoreHorizontal,
  Plus,
  Pencil,
  Trash2,
  Loader2,
  CreditCard,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import PaymentMethodFormModal from "~/components/settings/PaymentMethodFormModal.vue";
import type { PaymentMethod } from "~/components/settings/PaymentMethodFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

// ─── State ────────────────────────────────────────────────────────────────────

const methods = ref<PaymentMethod[]>([]);
const loading = ref(false);
const loadError = ref("");
const formModalOpen = ref(false);
const editingMethod = ref<PaymentMethod | null>(null);
const deletingId = ref<string | null>(null);

// ─── Computed ─────────────────────────────────────────────────────────────────

const activeCount = computed(
  () => methods.value.filter((m) => m.active).length,
);
const totalCount = computed(() => methods.value.length);

// ─── Helpers ─────────────────────────────────────────────────────────────────

const TYPE_LABELS: Record<string, string> = {
  pago_movil: "Pago Móvil",
  transfer_bank: "Transferencia",
  zelle: "Zelle",
  paypal: "PayPal",
  usdt: "USDT",
  cash: "Efectivo",
  other: "Otro",
};

const TYPE_ICONS: Record<string, typeof Smartphone> = {
  pago_movil: Smartphone,
  transfer_bank: Building2,
  zelle: DollarSign,
  paypal: DollarSign,
  usdt: Bitcoin,
  cash: Banknote,
  other: MoreHorizontal,
};

const TYPE_COLORS: Record<string, string> = {
  pago_movil: "bg-blue-100 text-blue-700",
  transfer_bank: "bg-violet-100 text-violet-700",
  zelle: "bg-yellow-100 text-yellow-700",
  paypal: "bg-sky-100 text-sky-700",
  usdt: "bg-emerald-100 text-emerald-700",
  cash: "bg-green-100 text-green-700",
  other: "bg-muted text-muted-foreground",
};

function methodIcon(type: string) {
  return TYPE_ICONS[type] ?? MoreHorizontal;
}

function typeLabel(type: string) {
  return TYPE_LABELS[type] ?? type;
}

function typeBadgeClass(type: string) {
  return TYPE_COLORS[type] ?? "bg-muted text-muted-foreground";
}

/**
 * Returns a concise key-detail string to display on the card.
 * Shows first 4 chars of sensitive fields and masks the rest.
 */
function keyDetail(m: PaymentMethod): string {
  const d = m.details;
  const t = m.type;
  if (t === "pago_movil") {
    const bank = d.bank ?? "";
    const phone = d.phone ?? "";
    return [bank, phone].filter(Boolean).join(" · ");
  } else if (t === "transfer_bank") {
    const bank = d.bank ?? "";
    const acc = d.account_number ?? "";
    const masked = acc.length > 4 ? "****" + acc.slice(-4) : acc;
    return [bank, masked].filter(Boolean).join(" · ");
  } else if (t === "zelle") {
    return d.email ?? "";
  } else if (t === "paypal") {
    return d.email ?? "";
  } else if (t === "usdt") {
    const wallet = d.wallet_address ?? "";
    const short =
      wallet.length > 8
        ? wallet.slice(0, 4) + "..." + wallet.slice(-4)
        : wallet;
    return [d.network, short].filter(Boolean).join(" · ");
  } else if (t === "cash") {
    return d.notes ? "Con instrucciones" : "Sin instrucciones adicionales";
  } else {
    return d.notes ?? "";
  }
}

function effectiveLabel(m: PaymentMethod): string {
  return m.label || typeLabel(m.type);
}

// ─── API ──────────────────────────────────────────────────────────────────────

async function loadMethods() {
  loading.value = true;
  loadError.value = "";
  try {
    methods.value = await useApiFetch<PaymentMethod[]>(
      "/api/v1/payment-methods",
    );
  } catch {
    loadError.value = "No se pudieron cargar los métodos de pago";
  } finally {
    loading.value = false;
  }
}

async function toggleActive(m: PaymentMethod) {
  try {
    const updated = await useApiFetch<PaymentMethod>(
      `/api/v1/payment-methods/${m.id}`,
      {
        method: "PUT",
        body: { active: !m.active },
      },
    );
    const idx = methods.value.findIndex((x) => x.id === updated.id);
    if (idx >= 0) methods.value[idx] = updated;
  } catch {
    // silently fail — user sees no change, can retry
  }
}

async function deleteMethod(id: string) {
  deletingId.value = id;
  try {
    await useApiFetch(`/api/v1/payment-methods/${id}`, { method: "DELETE" });
    methods.value = methods.value.filter((m) => m.id !== id);
  } catch {
    // ignore — keep in list
  } finally {
    deletingId.value = null;
  }
}

function openCreate() {
  editingMethod.value = null;
  formModalOpen.value = true;
}

function openEdit(m: PaymentMethod) {
  editingMethod.value = m;
  formModalOpen.value = true;
}

function onSaved(m: PaymentMethod) {
  const idx = methods.value.findIndex((x) => x.id === m.id);
  if (idx >= 0) {
    methods.value[idx] = m;
  } else {
    methods.value.push(m);
  }
  formModalOpen.value = false;
}

onMounted(loadMethods);
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6">
    <!-- Page header -->
    <div class="flex items-start justify-between gap-4 flex-wrap">
      <div>
        <h1 class="text-2xl font-bold text-foreground">Métodos de pago</h1>
        <p class="text-sm text-muted-foreground mt-0.5">
          Configurá cómo pueden pagarte tus clientes
        </p>
      </div>
      <button
        class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors shrink-0"
        @click="openCreate"
      >
        <Plus class="w-4 h-4" />
        Nuevo método
      </button>
    </div>

    <!-- KPIs -->
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
      <div class="border bg-card rounded-xl p-4">
        <p class="text-xs text-muted-foreground mb-1">Total</p>
        <p class="text-2xl font-bold text-foreground">{{ totalCount }}</p>
      </div>
      <div class="border bg-card rounded-xl p-4">
        <p class="text-xs text-muted-foreground mb-1">Activos</p>
        <p class="text-2xl font-bold text-emerald-600">{{ activeCount }}</p>
      </div>
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
      Cargando métodos de pago...
    </div>

    <!-- Empty state -->
    <div
      v-else-if="!loading && methods.length === 0"
      class="flex flex-col items-center justify-center py-20 text-center space-y-3"
    >
      <div
        class="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center"
      >
        <CreditCard class="w-8 h-8 text-primary/40" />
      </div>
      <div>
        <p class="font-semibold text-foreground">
          Aún no configuraste métodos de pago
        </p>
        <p class="text-sm text-muted-foreground mt-1 max-w-xs mx-auto">
          Sin esto, tus clientes no van a saber cómo pagarte. Agregá al menos
          uno.
        </p>
      </div>
      <button
        class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors"
        @click="openCreate"
      >
        <Plus class="w-4 h-4" />
        Agregar el primero
      </button>
    </div>

    <!-- Grid of method cards -->
    <div v-else-if="!loading" class="grid grid-cols-1 md:grid-cols-2 gap-3">
      <div
        v-for="m in methods"
        :key="m.id"
        :class="[
          'border bg-card rounded-xl p-4 flex flex-col gap-3 transition-all',
          m.active ? '' : 'opacity-60',
        ]"
      >
        <!-- Card top row: icon + label + badge -->
        <div class="flex items-start gap-3">
          <div
            :class="[
              'w-10 h-10 rounded-lg flex items-center justify-center shrink-0',
              typeBadgeClass(m.type).includes('bg-blue')
                ? 'bg-blue-100'
                : typeBadgeClass(m.type).includes('bg-violet')
                  ? 'bg-violet-100'
                  : typeBadgeClass(m.type).includes('bg-yellow')
                    ? 'bg-yellow-100'
                    : typeBadgeClass(m.type).includes('bg-sky')
                      ? 'bg-sky-100'
                      : typeBadgeClass(m.type).includes('bg-emerald')
                        ? 'bg-emerald-100'
                        : typeBadgeClass(m.type).includes('bg-green')
                          ? 'bg-green-100'
                          : 'bg-muted',
            ]"
          >
            <component
              :is="methodIcon(m.type)"
              :class="[
                'w-5 h-5',
                typeBadgeClass(m.type).split(' ')[1] ?? 'text-muted-foreground',
              ]"
            />
          </div>
          <div class="flex-1 min-w-0">
            <p class="font-semibold text-foreground text-sm truncate">
              {{ effectiveLabel(m) }}
            </p>
            <p class="text-xs text-muted-foreground truncate mt-0.5">
              {{ keyDetail(m) }}
            </p>
          </div>
          <span
            :class="[
              'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium shrink-0',
              typeBadgeClass(m.type),
            ]"
          >
            {{ typeLabel(m.type) }}
          </span>
        </div>

        <!-- Currency + active row -->
        <div class="flex items-center justify-between gap-2">
          <span
            v-if="m.currency"
            class="inline-flex items-center px-2 py-0.5 rounded-md bg-muted text-muted-foreground text-xs font-mono"
          >
            {{ m.currency }}
          </span>
          <span v-else class="text-xs text-muted-foreground/50 italic"
            >Sin moneda</span
          >

          <!-- Active toggle -->
          <button
            type="button"
            :class="[
              'relative inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0',
              m.active ? 'bg-primary' : 'bg-muted-foreground/30',
            ]"
            :title="m.active ? 'Desactivar' : 'Activar'"
            @click="toggleActive(m)"
          >
            <span
              :class="[
                'inline-block h-3.5 w-3.5 rounded-full bg-white shadow transition-transform',
                m.active ? 'translate-x-4' : 'translate-x-1',
              ]"
            />
          </button>
        </div>

        <!-- Actions -->
        <div class="flex items-center gap-2 pt-1 border-t border-border">
          <button
            class="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-lg border border-border hover:bg-muted text-sm text-foreground transition-colors"
            @click="openEdit(m)"
          >
            <Pencil class="w-3.5 h-3.5" />
            Editar
          </button>
          <button
            :disabled="deletingId === m.id"
            class="flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-lg border border-destructive/30 hover:bg-destructive/10 text-sm text-destructive transition-colors disabled:opacity-50"
            @click="deleteMethod(m.id)"
          >
            <Loader2
              v-if="deletingId === m.id"
              class="w-3.5 h-3.5 animate-spin"
            />
            <Trash2 v-else class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <PaymentMethodFormModal
      :open="formModalOpen"
      :method="editingMethod"
      @close="formModalOpen = false"
      @saved="onSaved"
    />
  </div>
</template>
