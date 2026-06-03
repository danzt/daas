<script setup lang="ts">
import {
  CheckCircle2,
  Clock,
  CreditCard,
  Package,
  Truck,
  Home,
  ShoppingCart,
  Mail,
  Phone,
  MapPin,
  AlertTriangle,
  Loader2,
  Share2,
  XCircle,
  PartyPopper,
  ArrowLeft,
  Copy,
  Check,
} from "lucide-vue-next";
import { usePublicFetch } from "~/composables/usePublicFetch";
import { useFormatPrice } from "~/composables/useFormatPrice";

definePageMeta({
  layout: "public",
});

const route = useRoute();
const router = useRouter();
const tenantSlug = route.params.tenantSlug as string;
const orderId = route.params.id as string;
const accessToken = (route.query.access_token as string) ?? "";
const isNewOrder = route.query.new === "1";

const { format: fmt } = useFormatPrice("VE");

interface OrderLine {
  id: string;
  product_id: string;
  name: string;
  unit_price: number;
  quantity: number;
  subtotal: number;
  is_fiscal: boolean;
  sort_order: number;
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
  updated_at: string;
  lines: OrderLine[];
}

const order = ref<ShopOrder | null>(null);
const loading = ref(true);
const loadError = ref("");
const cancelling = ref(false);
const cancelError = ref("");
const showConfirmCancel = ref(false);
const copied = ref(false);

useSeoMeta({
  title: () =>
    isNewOrder ? "¡Gracias por tu compra!" : `Pedido #${orderId.slice(0, 8)}`,
  description: "Seguimiento de tu pedido",
});

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    order.value = await usePublicFetch<ShopOrder>(
      `/orders/${orderId}?access_token=${accessToken}`,
    );
  } catch (err: unknown) {
    const e = err as { status?: number; statusCode?: number };
    if (e?.status === 404 || e?.statusCode === 404) {
      loadError.value = "Pedido no encontrado o link inválido";
    } else {
      loadError.value = "No pudimos cargar el pedido";
    }
  } finally {
    loading.value = false;
  }
}

async function cancelOrder() {
  if (!order.value) return;
  cancelling.value = true;
  cancelError.value = "";
  try {
    await usePublicFetch(
      `/orders/${orderId}/cancel?access_token=${accessToken}`,
      { method: "POST" },
    );
    await load();
    showConfirmCancel.value = false;
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } };
    cancelError.value = e?.data?.message ?? "No pudimos cancelar el pedido";
  } finally {
    cancelling.value = false;
  }
}

function copyLink() {
  if (!import.meta.client) return;
  navigator.clipboard.writeText(window.location.href);
  copied.value = true;
  setTimeout(() => (copied.value = false), 2000);
}

function fmtDate(d: string | null): string {
  if (!d) return "";
  return new Date(d).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const statusSteps = [
  { key: "pending", label: "Pedido recibido", icon: ShoppingCart },
  { key: "paid", label: "Pago confirmado", icon: CreditCard },
  { key: "fulfilled", label: "Preparado y despachado", icon: Package },
  { key: "delivered", label: "Entregado", icon: CheckCircle2 },
];

function stepIndex(status: string): number {
  return statusSteps.findIndex((s) => s.key === status);
}

function stepTimestamp(stepKey: string): string {
  if (!order.value) return "";
  const map: Record<string, string | null> = {
    pending: order.value.created_at,
    paid: order.value.paid_at,
    fulfilled: order.value.fulfilled_at,
    delivered: order.value.delivered_at,
  };
  return fmtDate(map[stepKey]);
}

const statusInfo = computed(() => {
  if (!order.value) return null;
  const map: Record<
    string,
    { label: string; color: string; bg: string; icon: typeof Clock }
  > = {
    pending: {
      label: "Pendiente de pago",
      color: "text-amber-700",
      bg: "bg-amber-50 border-amber-200",
      icon: Clock,
    },
    paid: {
      label: "Pagado · Preparando envío",
      color: "text-blue-700",
      bg: "bg-blue-50 border-blue-200",
      icon: CreditCard,
    },
    fulfilled: {
      label: "En camino",
      color: "text-purple-700",
      bg: "bg-purple-50 border-purple-200",
      icon: Truck,
    },
    delivered: {
      label: "Entregado",
      color: "text-emerald-700",
      bg: "bg-emerald-50 border-emerald-200",
      icon: CheckCircle2,
    },
    cancelled: {
      label: "Cancelado",
      color: "text-rose-700",
      bg: "bg-rose-50 border-rose-200",
      icon: XCircle,
    },
  };
  return map[order.value.status];
});

onMounted(load);
</script>

<template>
  <div class="bg-background min-h-[60vh]">
    <!-- Loading -->
    <div
      v-if="loading"
      class="max-w-3xl mx-auto px-4 sm:px-6 py-20 flex items-center justify-center text-muted-foreground"
    >
      <Loader2 class="w-6 h-6 animate-spin mr-2" />
      Cargando tu pedido...
    </div>

    <!-- Error -->
    <div
      v-else-if="loadError"
      class="max-w-3xl mx-auto px-4 sm:px-6 py-20 text-center"
    >
      <AlertTriangle class="w-16 h-16 text-muted-foreground/30 mx-auto mb-4" />
      <h1 class="text-2xl font-bold text-foreground mb-2">{{ loadError }}</h1>
      <p class="text-sm text-muted-foreground mb-6">
        Si recibiste un link por email, asegurate de abrirlo completo.
      </p>
      <NuxtLink
        :to="`/t/${tenantSlug}`"
        class="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-primary text-white font-semibold hover:opacity-90"
      >
        <Home class="w-4 h-4" />
        Volver al inicio
      </NuxtLink>
    </div>

    <!-- Order detail -->
    <template v-else-if="order">
      <!-- Hero -->
      <div
        :class="[
          'border-b',
          isNewOrder
            ? 'bg-gradient-to-br from-emerald-50 via-emerald-100/40 to-primary/5'
            : 'bg-card/50',
        ]"
      >
        <div class="max-w-3xl mx-auto px-4 sm:px-6 py-8 sm:py-10 text-center">
          <div
            v-if="isNewOrder"
            class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-emerald-500 text-white mb-4 shadow-lg shadow-emerald-500/30"
          >
            <PartyPopper class="w-8 h-8" />
          </div>
          <h1
            class="text-2xl sm:text-3xl font-bold font-heading text-foreground mb-2"
          >
            {{ isNewOrder ? "¡Gracias por tu compra!" : "Tu pedido" }}
          </h1>
          <p class="text-sm text-muted-foreground">
            {{
              isNewOrder
                ? "Recibimos tu pedido. Te enviamos los detalles por email."
                : "Acá podés seguir el estado de tu pedido"
            }}
          </p>
          <div
            class="mt-4 flex items-center justify-center gap-3 flex-wrap text-xs"
          >
            <span class="text-muted-foreground">
              Pedido
              <span class="font-mono font-semibold text-foreground"
                >#{{ order.id.slice(0, 8).toUpperCase() }}</span
              >
            </span>
            <span class="text-muted-foreground">·</span>
            <span class="text-muted-foreground">{{
              fmtDate(order.created_at)
            }}</span>
            <button
              type="button"
              class="ml-2 inline-flex items-center gap-1 px-2.5 py-1 rounded-md border bg-card hover:bg-muted transition-colors text-xs cursor-pointer"
              @click="copyLink"
            >
              <Check v-if="copied" class="w-3 h-3 text-emerald-600" />
              <Copy v-else class="w-3 h-3" />
              {{ copied ? "Copiado" : "Copiar link" }}
            </button>
          </div>
        </div>
      </div>

      <div class="max-w-3xl mx-auto px-4 sm:px-6 py-6 space-y-5">
        <!-- Status banner -->
        <div
          v-if="statusInfo"
          :class="[
            'rounded-xl border p-4 flex items-center gap-3',
            statusInfo.bg,
          ]"
        >
          <component
            :is="statusInfo.icon"
            :class="['w-6 h-6 shrink-0', statusInfo.color]"
          />
          <div class="flex-1">
            <p :class="['text-sm font-bold', statusInfo.color]">
              {{ statusInfo.label }}
            </p>
            <p
              v-if="order.status === 'pending'"
              class="text-xs text-muted-foreground mt-0.5"
            >
              Esperamos la confirmación de tu pago para procesar el pedido.
            </p>
            <p
              v-else-if="order.status === 'cancelled'"
              class="text-xs text-muted-foreground mt-0.5"
            >
              {{ fmtDate(order.cancelled_at) }}
            </p>
          </div>
        </div>

        <!-- Timeline -->
        <div
          v-if="order.status !== 'cancelled'"
          class="border bg-card rounded-xl p-5 shadow-sm"
        >
          <h2 class="text-base font-bold text-foreground mb-5">
            Seguimiento del pedido
          </h2>
          <ol class="space-y-4">
            <li
              v-for="(step, idx) in statusSteps"
              :key="step.key"
              class="flex items-start gap-3"
            >
              <div
                :class="[
                  'w-9 h-9 rounded-full flex items-center justify-center shrink-0 transition-colors',
                  stepIndex(order.status) >= idx
                    ? 'bg-primary text-white'
                    : 'bg-muted text-muted-foreground',
                ]"
              >
                <component :is="step.icon" class="w-4 h-4" />
              </div>
              <div class="flex-1 min-w-0 pt-1.5">
                <p
                  :class="[
                    'text-sm font-semibold',
                    stepIndex(order.status) >= idx
                      ? 'text-foreground'
                      : 'text-muted-foreground',
                  ]"
                >
                  {{ step.label }}
                </p>
                <p
                  v-if="
                    stepIndex(order.status) >= idx && stepTimestamp(step.key)
                  "
                  class="text-xs text-muted-foreground mt-0.5"
                >
                  {{ stepTimestamp(step.key) }}
                </p>
              </div>
            </li>
          </ol>
        </div>

        <!-- Line items -->
        <div class="border bg-card rounded-xl p-5 shadow-sm">
          <h2 class="text-base font-bold text-foreground mb-4">
            Productos ({{ order.lines.length }})
          </h2>
          <ul class="divide-y divide-border">
            <li
              v-for="line in order.lines"
              :key="line.id"
              class="py-3 first:pt-0 last:pb-0 flex items-start gap-3"
            >
              <div
                class="w-10 h-10 rounded-md bg-primary/10 flex items-center justify-center shrink-0 text-primary text-xs font-bold"
              >
                {{ line.quantity }}
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-foreground line-clamp-2">
                  {{ line.name }}
                </p>
                <p
                  v-if="line.is_fiscal"
                  class="text-[10px] text-muted-foreground"
                >
                  Fiscal · IVA incluido
                </p>
              </div>
              <p
                class="text-sm font-semibold font-mono text-foreground whitespace-nowrap"
              >
                {{ fmt(line.subtotal) }}
              </p>
            </li>
          </ul>

          <dl class="space-y-2 text-sm pt-4 border-t mt-3">
            <div class="flex items-center justify-between">
              <dt class="text-muted-foreground">Subtotal</dt>
              <dd class="font-semibold text-foreground font-mono">
                {{ fmt(order.subtotal) }}
              </dd>
            </div>
            <div class="flex items-center justify-between">
              <dt class="text-muted-foreground">Envío</dt>
              <dd
                :class="[
                  'font-semibold font-mono',
                  order.shipping_cost === 0
                    ? 'text-emerald-600'
                    : 'text-foreground',
                ]"
              >
                {{
                  order.shipping_cost === 0
                    ? "Gratis"
                    : fmt(order.shipping_cost)
                }}
              </dd>
            </div>
            <div class="h-px bg-border my-2" />
            <div class="flex items-center justify-between text-base">
              <dt class="font-bold text-foreground">Total</dt>
              <dd class="font-bold text-foreground font-mono text-xl">
                {{ fmt(order.total) }}
              </dd>
            </div>
          </dl>
        </div>

        <!-- Customer + shipping -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div class="border bg-card rounded-xl p-5 shadow-sm space-y-2">
            <h3
              class="text-sm font-bold text-foreground flex items-center gap-2 mb-2"
            >
              <Mail class="w-4 h-4 text-primary" />
              Contacto
            </h3>
            <p class="text-sm text-foreground font-medium">
              {{ order.customer_name }}
            </p>
            <p class="text-xs text-muted-foreground flex items-center gap-1.5">
              <Mail class="w-3 h-3" />
              {{ order.customer_email }}
            </p>
            <p
              v-if="order.customer_phone"
              class="text-xs text-muted-foreground flex items-center gap-1.5"
            >
              <Phone class="w-3 h-3" />
              {{ order.customer_phone }}
            </p>
          </div>

          <div class="border bg-card rounded-xl p-5 shadow-sm space-y-1">
            <h3
              class="text-sm font-bold text-foreground flex items-center gap-2 mb-2"
            >
              <MapPin class="w-4 h-4 text-primary" />
              Envío a
            </h3>
            <p class="text-sm text-foreground">
              {{ order.shipping_address }}
            </p>
            <p v-if="order.shipping_city" class="text-xs text-muted-foreground">
              {{ order.shipping_city }}
            </p>
            <p
              v-if="order.shipping_notes"
              class="text-xs text-muted-foreground mt-2 italic"
            >
              "{{ order.shipping_notes }}"
            </p>
          </div>
        </div>

        <!-- Notes -->
        <div
          v-if="order.notes"
          class="border bg-card rounded-xl p-4 text-sm text-muted-foreground"
        >
          <p class="text-xs font-semibold text-foreground mb-1">Tus notas:</p>
          {{ order.notes }}
        </div>

        <!-- Cancel order (pending only) -->
        <div
          v-if="order.status === 'pending'"
          class="border bg-card rounded-xl p-5 shadow-sm"
        >
          <div
            v-if="!showConfirmCancel"
            class="flex items-center justify-between gap-3 flex-wrap"
          >
            <div class="flex-1 min-w-0">
              <p class="text-sm font-semibold text-foreground">
                ¿Cambiaste de opinión?
              </p>
              <p class="text-xs text-muted-foreground">
                Mientras el pedido esté pendiente podés cancelarlo sin costo.
              </p>
            </div>
            <button
              type="button"
              class="px-4 py-2 rounded-lg border border-destructive/40 text-destructive text-sm font-semibold hover:bg-destructive/10 transition-colors cursor-pointer"
              @click="showConfirmCancel = true"
            >
              Cancelar pedido
            </button>
          </div>

          <div v-else class="space-y-3">
            <div class="flex items-start gap-2 text-sm">
              <AlertTriangle class="w-5 h-5 text-amber-500 shrink-0 mt-0.5" />
              <p class="text-foreground">
                ¿Seguro que querés cancelar este pedido? Esta acción no se puede
                deshacer.
              </p>
            </div>
            <p v-if="cancelError" class="text-xs text-destructive">
              {{ cancelError }}
            </p>
            <div class="flex items-center gap-2">
              <button
                type="button"
                :disabled="cancelling"
                class="px-4 py-2 rounded-lg bg-destructive text-destructive-foreground text-sm font-semibold hover:opacity-90 disabled:opacity-50 cursor-pointer flex items-center gap-2"
                @click="cancelOrder"
              >
                <Loader2 v-if="cancelling" class="w-4 h-4 animate-spin" />
                Sí, cancelar
              </button>
              <button
                type="button"
                class="px-4 py-2 rounded-lg border text-sm font-semibold hover:bg-muted cursor-pointer"
                @click="showConfirmCancel = false"
              >
                No, mantener
              </button>
            </div>
          </div>
        </div>

        <!-- Footer actions -->
        <div class="flex items-center justify-between gap-3 pt-3 flex-wrap">
          <NuxtLink
            :to="`/t/${tenantSlug}`"
            class="inline-flex items-center gap-2 text-sm text-primary hover:underline font-semibold"
          >
            <ArrowLeft class="w-4 h-4" />
            Seguir comprando
          </NuxtLink>
          <button
            type="button"
            class="inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
            @click="copyLink"
          >
            <Share2 class="w-3.5 h-3.5" />
            Compartir mi pedido
          </button>
        </div>
      </div>
    </template>
  </div>
</template>
