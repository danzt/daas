<script setup lang="ts">
import {
  ShoppingCart,
  Clock,
  CreditCard,
  Package,
  CheckCircle2,
  XCircle,
  Mail,
  Phone,
  MapPin,
  Paperclip,
  FileText,
  Download,
  Loader2,
  AlertTriangle,
  Truck,
  Send,
} from "lucide-vue-next";

interface OrderLine {
  id: string;
  name: string;
  unit_price: number;
  quantity: number;
  subtotal: number;
  is_fiscal: boolean;
}

interface PaymentMethod {
  id: string;
  type: string;
  label: string;
  currency: string;
}

interface ShopOrder {
  id: string;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  shipping_address: string;
  shipping_city: string;
  shipping_notes: string;
  subtotal: number;
  shipping_cost: number;
  total: number;
  status: "pending" | "paid" | "fulfilled" | "delivered" | "cancelled";
  notes: string;
  paid_at: string | null;
  fulfilled_at: string | null;
  delivered_at: string | null;
  cancelled_at: string | null;
  created_at: string;
  lines: OrderLine[];
  payment_proof_url?: string;
  payment_proof_filename?: string;
  payment_proof_uploaded_at?: string;
  payment_method_id?: string;
  payment_reference?: string;
}

const props = defineProps<{
  order: ShopOrder | null;
  loading: boolean;
  loadError: string;
  actionLoading: string | null;
  actionError: string;
  paymentMethods: PaymentMethod[];
}>();

const emit = defineEmits<{
  markPaid: [];
  markFulfilled: [];
  markDelivered: [];
  cancel: [];
  retry: [];
}>();

const confirmingCancel = ref(false);

const STATUS: Record<string, { label: string; chip: string; icon: Component }> =
  {
    pending: {
      label: "Pendiente de pago",
      chip: "bg-amber-50 text-amber-700",
      icon: Clock,
    },
    paid: {
      label: "Pagada",
      chip: "bg-blue-100 text-blue-700",
      icon: CreditCard,
    },
    fulfilled: {
      label: "Despachada",
      chip: "bg-purple-100 text-purple-700",
      icon: Package,
    },
    delivered: {
      label: "Entregada",
      chip: "bg-emerald-100 text-emerald-700",
      icon: CheckCircle2,
    },
    cancelled: {
      label: "Cancelada",
      chip: "bg-rose-100 text-rose-600",
      icon: XCircle,
    },
  };

function proofIsImage(url?: string, filename?: string) {
  return /\.(jpe?g|png|webp)$/i.test(filename ?? url ?? "");
}

function proofMethodLabel() {
  if (!props.order?.payment_method_id) return "";
  const pm = props.paymentMethods.find(
    (m) => m.id === props.order!.payment_method_id,
  );
  return pm?.label || pm?.type || "";
}

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
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const hasActionBar = computed(
  () => props.order && !["delivered", "cancelled"].includes(props.order.status),
);
</script>

<template>
  <MobileScreen title="Pedido online" back subtitle="B2C">
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

    <template v-else-if="order">
      <!-- Hero -->
      <div class="px-4 pt-4">
        <div class="rounded-2xl border border-border bg-white p-5 space-y-3">
          <div class="flex items-center justify-between gap-2">
            <span
              class="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-bold"
              :class="STATUS[order.status]?.chip"
            >
              <component :is="STATUS[order.status]?.icon" class="size-3.5" />
              {{ STATUS[order.status]?.label }}
            </span>
            <span class="font-mono text-xs text-muted-foreground">
              #{{ order.id.slice(0, 8).toUpperCase() }}
            </span>
          </div>

          <div>
            <p class="text-2xl font-black tabular-nums text-foreground">
              Bs.S {{ fmtCurrency(order.total) }}
            </p>
            <p class="mt-1 text-sm font-semibold text-foreground">
              {{ order.customer_name }}
            </p>
          </div>

          <div class="space-y-1">
            <a
              :href="`mailto:${order.customer_email}`"
              class="flex items-center gap-2 text-xs text-primary"
            >
              <Mail class="size-3.5 shrink-0" />
              {{ order.customer_email }}
            </a>
            <a
              v-if="order.customer_phone"
              :href="`tel:${order.customer_phone}`"
              class="flex items-center gap-2 text-xs text-primary"
            >
              <Phone class="size-3.5 shrink-0" />
              {{ order.customer_phone }}
            </a>
          </div>

          <!-- Timeline row -->
          <div
            class="grid grid-cols-2 gap-2 pt-2 border-t border-border text-xs"
          >
            <div>
              <p class="text-muted-foreground">Recibido</p>
              <p class="font-medium text-foreground">
                {{ fmtDate(order.created_at) }}
              </p>
            </div>
            <div>
              <p class="text-muted-foreground">Pagado</p>
              <p class="font-medium text-foreground">
                {{ fmtDate(order.paid_at) }}
              </p>
            </div>
            <div>
              <p class="text-muted-foreground">Despachado</p>
              <p class="font-medium text-foreground">
                {{ fmtDate(order.fulfilled_at) }}
              </p>
            </div>
            <div>
              <p class="text-muted-foreground">Entregado</p>
              <p class="font-medium text-foreground">
                {{ fmtDate(order.delivered_at) }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Payment proof -->
      <template v-if="order.payment_proof_url || order.status === 'pending'">
        <MobileSectionHeader title="Comprobante de pago" class="pt-5" />
        <div
          class="mx-4 rounded-2xl border bg-white p-4"
          :class="
            order.payment_proof_url
              ? 'border-emerald-200 bg-emerald-50/60'
              : 'border-border'
          "
        >
          <template v-if="order.payment_proof_url">
            <a
              v-if="
                proofIsImage(
                  order.payment_proof_url,
                  order.payment_proof_filename,
                )
              "
              :href="order.payment_proof_url"
              target="_blank"
              rel="noopener noreferrer"
              class="block"
            >
              <img
                :src="order.payment_proof_url"
                :alt="order.payment_proof_filename ?? 'Comprobante'"
                class="w-full max-h-56 rounded-xl object-contain border"
              />
            </a>
            <a
              v-else
              :href="order.payment_proof_url"
              target="_blank"
              rel="noopener noreferrer"
              download
              class="inline-flex items-center gap-2 text-sm text-primary"
            >
              <FileText class="size-4 shrink-0" />
              {{ order.payment_proof_filename ?? "Ver comprobante" }}
              <Download class="size-3.5" />
            </a>

            <div class="mt-3 space-y-1 text-xs text-muted-foreground">
              <div v-if="order.payment_proof_uploaded_at" class="flex gap-2">
                <span class="w-20 shrink-0">Subido</span>
                <span>{{ fmtDate(order.payment_proof_uploaded_at) }}</span>
              </div>
              <div v-if="proofMethodLabel()" class="flex gap-2">
                <span class="w-20 shrink-0">Método</span>
                <span>{{ proofMethodLabel() }}</span>
              </div>
              <div v-if="order.payment_reference" class="flex gap-2">
                <span class="w-20 shrink-0">Referencia</span>
                <span class="font-mono">{{ order.payment_reference }}</span>
              </div>
            </div>
          </template>
          <p v-else class="text-xs text-muted-foreground">
            El cliente aún no subió el comprobante.
          </p>
        </div>
      </template>

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
        :title="`Productos (${order.lines?.length ?? 0})`"
        class="pt-5"
      />
      <div class="divide-y divide-border border-y border-border bg-white">
        <div
          v-for="line in order.lines"
          :key="line.id"
          class="flex items-start gap-3 px-4 py-3"
        >
          <span
            class="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-lg bg-muted text-xs font-bold text-muted-foreground"
          >
            {{ line.quantity }}×
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-foreground">{{ line.name }}</p>
            <div class="flex items-center gap-1.5 mt-0.5">
              <span class="text-[11px] text-muted-foreground">
                Bs.S {{ fmtCurrency(line.unit_price) }} c/u
              </span>
              <span
                v-if="line.is_fiscal"
                class="rounded px-1.5 py-px text-[10px] font-bold bg-primary/10 text-primary"
              >
                FISCAL
              </span>
            </div>
          </div>
          <span
            class="font-mono text-sm font-bold tabular-nums text-foreground"
          >
            Bs.S {{ fmtCurrency(line.subtotal) }}
          </span>
        </div>

        <div class="flex items-center justify-between px-4 py-2.5 text-sm">
          <span class="text-muted-foreground">Subtotal</span>
          <span class="font-mono tabular-nums"
            >Bs.S {{ fmtCurrency(order.subtotal) }}</span
          >
        </div>
        <div class="flex items-center justify-between px-4 py-2.5 text-sm">
          <span class="text-muted-foreground">Envío</span>
          <span
            class="font-mono tabular-nums"
            :class="
              order.shipping_cost === 0 ? 'text-emerald-600' : 'text-foreground'
            "
          >
            {{
              order.shipping_cost === 0
                ? "Gratis"
                : `Bs.S ${fmtCurrency(order.shipping_cost)}`
            }}
          </span>
        </div>
        <div class="flex items-center justify-between px-4 py-3.5">
          <span class="text-sm font-bold text-foreground">Total</span>
          <span class="font-mono text-lg font-black tabular-nums text-primary">
            Bs.S {{ fmtCurrency(order.total) }}
          </span>
        </div>
      </div>

      <!-- Shipping -->
      <MobileSectionHeader title="Envío" class="pt-5" />
      <div class="divide-y divide-border border-y border-border bg-white">
        <div class="flex items-start gap-3 px-4 py-3">
          <Truck class="mt-0.5 size-5 shrink-0 text-muted-foreground" />
          <div>
            <p class="text-sm text-foreground">{{ order.shipping_address }}</p>
            <p v-if="order.shipping_city" class="text-xs text-muted-foreground">
              {{ order.shipping_city }}
            </p>
          </div>
        </div>
        <div
          v-if="order.shipping_notes"
          class="flex items-start gap-3 px-4 py-3"
        >
          <MapPin class="mt-0.5 size-5 shrink-0 text-muted-foreground" />
          <p class="text-xs italic text-muted-foreground">
            "{{ order.shipping_notes }}"
          </p>
        </div>
      </div>

      <!-- Customer notes -->
      <div v-if="order.notes" class="px-4 pt-4">
        <p
          class="rounded-2xl bg-amber-50 border border-amber-200 p-3.5 text-sm text-amber-800"
        >
          {{ order.notes }}
        </p>
      </div>

      <!-- Terminal state -->
      <div v-if="order.status === 'delivered'" class="px-4 pt-4">
        <p class="flex items-center gap-2 text-sm text-emerald-600">
          <CheckCircle2 class="size-4" />
          Entregada el {{ fmtDate(order.delivered_at) }}
        </p>
      </div>
      <div v-if="order.status === 'cancelled'" class="px-4 pt-4">
        <p class="flex items-center gap-2 text-sm text-rose-600">
          <XCircle class="size-4" />
          Cancelada el {{ fmtDate(order.cancelled_at) }}
        </p>
      </div>

      <!-- Spacer for action bar -->
      <div v-if="hasActionBar" class="h-24" />
    </template>

    <!-- Action bar -->
    <template v-if="hasActionBar && order">
      <div
        class="fixed inset-x-0 bottom-0 z-40 border-t border-border bg-white px-4 pb-[calc(0.75rem+env(safe-area-inset-bottom))] pt-3"
      >
        <!-- Confirming cancel -->
        <div v-if="confirmingCancel" class="flex gap-2">
          <button
            type="button"
            :disabled="actionLoading === 'cancel'"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-red-500 text-sm font-bold text-white active:bg-red-600 disabled:opacity-50"
            @click="emit('cancel')"
          >
            <Loader2
              v-if="actionLoading === 'cancel'"
              class="mx-auto size-5 animate-spin"
            />
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
          <!-- Cancel icon (pending / paid only) -->
          <button
            v-if="order.status === 'pending' || order.status === 'paid'"
            type="button"
            class="no-min-tap h-12 w-12 shrink-0 rounded-2xl border border-rose-200 text-rose-500 active:bg-rose-50"
            aria-label="Cancelar pedido"
            @click="confirmingCancel = true"
          >
            <XCircle class="mx-auto size-5" />
          </button>

          <!-- pending → mark paid -->
          <button
            v-if="order.status === 'pending'"
            type="button"
            :disabled="actionLoading !== null"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-blue-600 text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('markPaid')"
          >
            <Loader2
              v-if="actionLoading === 'mark-paid'"
              class="mx-auto size-5 animate-spin"
            />
            <span v-else>Confirmar pago</span>
          </button>

          <!-- paid → mark fulfilled -->
          <button
            v-else-if="order.status === 'paid'"
            type="button"
            :disabled="actionLoading !== null"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-purple-600 text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('markFulfilled')"
          >
            <Loader2
              v-if="actionLoading === 'mark-fulfilled'"
              class="mx-auto size-5 animate-spin"
            />
            <span v-else class="flex items-center justify-center gap-2">
              <Send class="size-4" />
              Marcar despachada
            </span>
          </button>

          <!-- fulfilled → mark delivered -->
          <button
            v-else-if="order.status === 'fulfilled'"
            type="button"
            :disabled="actionLoading !== null"
            class="no-min-tap h-12 flex-1 rounded-2xl bg-emerald-600 text-sm font-bold text-white active:opacity-90 disabled:opacity-50"
            @click="emit('markDelivered')"
          >
            <Loader2
              v-if="actionLoading === 'mark-delivered'"
              class="mx-auto size-5 animate-spin"
            />
            <span v-else class="flex items-center justify-center gap-2">
              <CheckCircle2 class="size-4" />
              Marcar entregada
            </span>
          </button>
        </div>
      </div>
    </template>
  </MobileScreen>
</template>
