<script setup lang="ts">
import {
  Loader2,
  AlertTriangle,
  Mail,
  ShoppingBag,
  ArrowRight,
  Clock,
  CreditCard,
  Package,
  CheckCircle2,
  XCircle,
  Truck,
  Search,
  X,
} from "lucide-vue-next";
import { useFormatPrice } from "~/composables/useFormatPrice";

definePageMeta({
  layout: "public",
});

useSeoMeta({
  title: "Mis pedidos",
  description: "Consultá el historial de tus pedidos",
});

const route = useRoute();
const router = useRouter();
const tenantSlug = route.params.tenantSlug as string;

const { format: fmt } = useFormatPrice("VE");

// ─── Types ────────────────────────────────────────────────────────────────────

interface OrderLine {
  id: string;
  name: string;
  quantity: number;
  subtotal: number;
}

interface CustomerOrder {
  id: string;
  customer_email: string;
  total: number;
  status: "pending" | "paid" | "fulfilled" | "delivered" | "cancelled";
  created_at: string;
  lines: OrderLine[];
  access_token: string;
}

// ─── State ────────────────────────────────────────────────────────────────────

const emailInput = ref((route.query.email as string) ?? "");
const submittedEmail = ref((route.query.email as string) ?? "");

const orders = ref<CustomerOrder[]>([]);
const loading = ref(false);
const fetchError = ref("");

// ─── Status display helpers ────────────────────────────────────────────────────

type StatusMeta = {
  label: string;
  badgeClass: string;
  icon: typeof Clock;
};

const STATUS_META: Record<string, StatusMeta> = {
  pending: {
    label: "Pendiente",
    badgeClass: "bg-amber-100 text-amber-800 border border-amber-200",
    icon: Clock,
  },
  paid: {
    label: "Pagado",
    badgeClass: "bg-blue-100 text-blue-800 border border-blue-200",
    icon: CreditCard,
  },
  fulfilled: {
    label: "En camino",
    badgeClass: "bg-purple-100 text-purple-800 border border-purple-200",
    icon: Truck,
  },
  delivered: {
    label: "Entregado",
    badgeClass: "bg-emerald-100 text-emerald-800 border border-emerald-200",
    icon: CheckCircle2,
  },
  cancelled: {
    label: "Cancelado",
    badgeClass: "bg-rose-100 text-rose-800 border border-rose-200",
    icon: XCircle,
  },
};

function statusMeta(status: string): StatusMeta {
  return (
    STATUS_META[status] ?? {
      label: status,
      badgeClass: "bg-muted text-muted-foreground border",
      icon: Package,
    }
  );
}

// ─── Date formatting ──────────────────────────────────────────────────────────

function fmtDate(d: string): string {
  return new Date(d).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}

// ─── Fetch logic ──────────────────────────────────────────────────────────────

const config = useRuntimeConfig();

async function fetchOrders(email: string) {
  loading.value = true;
  fetchError.value = "";
  orders.value = [];
  try {
    const data = await $fetch<CustomerOrder[]>(
      `/t/${tenantSlug}/shop/v1/my-orders`,
      {
        baseURL: config.public.apiBase as string,
        query: { email },
      },
    );
    orders.value = Array.isArray(data) ? data : [];
  } catch (err: unknown) {
    const e = err as {
      status?: number;
      statusCode?: number;
      data?: { message?: string };
    };
    const status = e?.status ?? e?.statusCode;
    if (status === 400) {
      fetchError.value = "El email ingresado no es válido.";
    } else {
      fetchError.value =
        e?.data?.message ??
        "No pudimos consultar tus pedidos. Intentá de nuevo.";
    }
  } finally {
    loading.value = false;
  }
}

// Run on load if email already in URL
onMounted(() => {
  if (submittedEmail.value) {
    fetchOrders(submittedEmail.value);
  }
});

// ─── Form submit ──────────────────────────────────────────────────────────────

const inputError = ref("");

function submit() {
  inputError.value = "";
  const email = emailInput.value.trim();
  if (!email) {
    inputError.value = "Ingresá tu email para continuar.";
    return;
  }
  if (!email.includes("@")) {
    inputError.value = "El email ingresado no parece válido.";
    return;
  }
  submittedEmail.value = email;
  router.replace({ query: { email } });
  fetchOrders(email);
}

function clearEmail() {
  submittedEmail.value = "";
  emailInput.value = "";
  orders.value = [];
  fetchError.value = "";
  router.replace({ query: {} });
}
</script>

<template>
  <div class="bg-background min-h-[60vh]">
    <!-- Page header -->
    <div class="border-b bg-card/50">
      <div class="max-w-3xl mx-auto px-4 sm:px-6 py-8 sm:py-10">
        <div class="flex items-center gap-3 mb-1">
          <div
            class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center shrink-0"
          >
            <ShoppingBag class="w-5 h-5 text-primary" />
          </div>
          <h1
            class="text-2xl sm:text-3xl font-bold font-heading text-foreground"
          >
            Mis pedidos
          </h1>
        </div>
        <p class="text-sm text-muted-foreground mt-2 ml-[3.25rem]">
          Consultá el historial y estado de tus compras
        </p>
      </div>
    </div>

    <div class="max-w-3xl mx-auto px-4 sm:px-6 py-8 space-y-6">
      <!-- ── Email lookup form ─────────────────────────────────────────────── -->
      <div
        v-if="!submittedEmail"
        class="border bg-card rounded-2xl p-6 sm:p-8 shadow-sm"
      >
        <div class="max-w-md mx-auto space-y-5">
          <div class="text-center space-y-1.5">
            <div
              class="inline-flex items-center justify-center w-14 h-14 rounded-full bg-primary/10 text-primary mb-3"
            >
              <Mail class="w-7 h-7" />
            </div>
            <h2 class="text-lg font-bold text-foreground font-heading">
              ¿Cuál es tu email?
            </h2>
            <p class="text-sm text-muted-foreground">
              Ingresá el email con el que hiciste tu pedido para ver tu
              historial
            </p>
          </div>

          <div class="space-y-3">
            <div class="space-y-1.5">
              <label
                for="email-input"
                class="text-xs font-semibold text-foreground"
              >
                Email
              </label>
              <div class="relative">
                <Mail
                  class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground"
                />
                <input
                  id="email-input"
                  v-model="emailInput"
                  type="email"
                  autocomplete="email"
                  placeholder="tu@email.com"
                  :class="[
                    'w-full pl-9 pr-4 py-2.5 rounded-xl border bg-background text-sm outline-none transition-all placeholder:text-muted-foreground/60',
                    inputError
                      ? 'border-destructive focus:ring-2 focus:ring-destructive/30'
                      : 'border-border focus:ring-2 focus:ring-primary/30',
                  ]"
                  @keydown.enter="submit"
                />
              </div>
              <p
                v-if="inputError"
                class="text-xs text-destructive flex items-center gap-1"
              >
                <AlertTriangle class="w-3.5 h-3.5 shrink-0" />
                {{ inputError }}
              </p>
            </div>

            <button
              type="button"
              class="w-full px-4 py-2.5 rounded-xl bg-primary text-primary-foreground text-sm font-semibold hover:opacity-90 transition-opacity cursor-pointer flex items-center justify-center gap-2"
              @click="submit"
            >
              <Search class="w-4 h-4" />
              Ver mis pedidos
            </button>
          </div>
        </div>
      </div>

      <!-- ── Results view ──────────────────────────────────────────────────── -->
      <template v-else>
        <!-- Header with email + change link -->
        <div class="flex items-center justify-between gap-3 flex-wrap">
          <div class="flex items-center gap-2 min-w-0">
            <Mail class="w-4 h-4 text-muted-foreground shrink-0" />
            <p class="text-sm text-foreground font-medium truncate">
              Pedidos de
              <span class="font-semibold text-primary">{{
                submittedEmail
              }}</span>
            </p>
          </div>
          <button
            type="button"
            class="inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer shrink-0"
            @click="clearEmail"
          >
            <X class="w-3.5 h-3.5" />
            Cambiar email
          </button>
        </div>

        <!-- Loading -->
        <div
          v-if="loading"
          class="flex items-center justify-center gap-2 py-16 text-muted-foreground"
        >
          <Loader2 class="w-5 h-5 animate-spin" />
          <span class="text-sm">Buscando tus pedidos...</span>
        </div>

        <!-- Error -->
        <div
          v-else-if="fetchError"
          class="border bg-card rounded-2xl p-8 shadow-sm text-center space-y-4"
        >
          <AlertTriangle class="w-12 h-12 text-muted-foreground/30 mx-auto" />
          <p class="text-sm font-semibold text-foreground">{{ fetchError }}</p>
          <button
            type="button"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-primary-foreground text-sm font-semibold hover:opacity-90 cursor-pointer"
            @click="fetchOrders(submittedEmail)"
          >
            <Loader2 class="w-4 h-4" />
            Reintentar
          </button>
        </div>

        <!-- Empty state -->
        <div
          v-else-if="orders.length === 0"
          class="border bg-card rounded-2xl p-10 shadow-sm text-center space-y-3"
        >
          <ShoppingBag class="w-14 h-14 text-muted-foreground/25 mx-auto" />
          <h3 class="text-base font-semibold text-foreground">
            No encontramos pedidos para este email
          </h3>
          <p class="text-sm text-muted-foreground max-w-xs mx-auto">
            Verificá que el email sea el mismo que usaste al hacer tu pedido.
          </p>
          <button
            type="button"
            class="inline-flex items-center gap-1.5 text-xs text-primary hover:underline cursor-pointer mt-2"
            @click="clearEmail"
          >
            <Mail class="w-3.5 h-3.5" />
            Probar con otro email
          </button>
        </div>

        <!-- Orders list -->
        <ul v-else class="space-y-3">
          <li
            v-for="order in orders"
            :key="order.id"
            class="border bg-card rounded-2xl p-5 shadow-sm hover:shadow-md transition-shadow"
          >
            <div class="flex items-start justify-between gap-3 flex-wrap">
              <!-- Left: order info -->
              <div class="space-y-2 min-w-0 flex-1">
                <!-- Number + date row -->
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="text-sm font-bold text-foreground font-mono">
                    #{{ order.id.slice(0, 8).toUpperCase() }}
                  </span>
                  <span class="text-muted-foreground text-xs">·</span>
                  <span class="text-xs text-muted-foreground">
                    {{ fmtDate(order.created_at) }}
                  </span>
                </div>

                <!-- Status badge -->
                <span
                  :class="[
                    'inline-flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-full',
                    statusMeta(order.status).badgeClass,
                  ]"
                >
                  <component
                    :is="statusMeta(order.status).icon"
                    class="w-3 h-3"
                  />
                  {{ statusMeta(order.status).label }}
                </span>

                <!-- Item count -->
                <p class="text-xs text-muted-foreground">
                  {{
                    (order.lines ?? []).length === 1
                      ? "1 producto"
                      : `${(order.lines ?? []).length} productos`
                  }}
                </p>
              </div>

              <!-- Right: total + CTA -->
              <div class="flex flex-col items-end gap-3 shrink-0">
                <p class="text-base font-bold text-foreground font-mono">
                  {{ fmt(order.total) }}
                </p>
                <NuxtLink
                  :to="`/t/${tenantSlug}/order/${order.id}?access_token=${order.access_token}`"
                  class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-primary text-primary-foreground text-xs font-semibold hover:opacity-90 transition-opacity whitespace-nowrap"
                >
                  Ver detalle
                  <ArrowRight class="w-3.5 h-3.5" />
                </NuxtLink>
              </div>
            </div>
          </li>
        </ul>
      </template>
    </div>
  </div>
</template>
