<script setup lang="ts">
import {
  ArrowLeft,
  Search,
  Plus,
  Minus,
  Trash2,
  ShoppingCart,
  Loader2,
  Package,
  Check,
  User,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import type { Product } from "~/components/products/ProductFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

const { isMobile } = useMobileMode();

const router = useRouter();

// ─── Types ────────────────────────────────────────────────────────────────────

interface CartLine {
  productId: string;
  description: string;
  unitPrice: number;
  quantity: number;
}

// ─── State ────────────────────────────────────────────────────────────────────

const products = ref<Product[]>([]);
const loadingProducts = ref(false);
const searchQuery = ref("");

const cart = ref<CartLine[]>([]);
const customerExpanded = ref(false);
const customer = reactive({
  name: "",
  id_type: "anonymous",
  id_number: "",
  notes: "",
});

const saving = ref(false);
const saveError = ref("");

// ─── Computed ─────────────────────────────────────────────────────────────────

const filteredProducts = computed(() => {
  const q = searchQuery.value.toLowerCase().trim();
  if (!q) return products.value;
  return products.value.filter(
    (p) =>
      p.name.toLowerCase().includes(q) ||
      (p.sku ?? "").toLowerCase().includes(q),
  );
});

const cartTotal = computed(() =>
  cart.value.reduce((sum, l) => sum + l.quantity * l.unitPrice, 0),
);

const cartCount = computed(() =>
  cart.value.reduce((sum, l) => sum + l.quantity, 0),
);

function cartQtyFor(productId: string) {
  return cart.value.find((l) => l.productId === productId)?.quantity ?? 0;
}

// ─── Cart actions ─────────────────────────────────────────────────────────────

function addToCart(product: Product) {
  const line = cart.value.find((l) => l.productId === product.id);
  if (line) {
    line.quantity++;
  } else {
    cart.value.push({
      productId: product.id,
      description: product.name,
      unitPrice: product.internal_price ?? product.fiscal_price ?? 0,
      quantity: 1,
    });
  }
}

function increment(idx: number) {
  cart.value[idx].quantity++;
}

function decrement(idx: number) {
  if (cart.value[idx].quantity <= 1) {
    cart.value.splice(idx, 1);
  } else {
    cart.value[idx].quantity--;
  }
}

function removeLine(idx: number) {
  cart.value.splice(idx, 1);
}

// ─── Submit ───────────────────────────────────────────────────────────────────

async function submitOrder() {
  if (cart.value.length === 0) {
    saveError.value = "Agregá al menos un producto al carrito";
    return;
  }
  saving.value = true;
  saveError.value = "";
  try {
    const order = await useApiFetch<{ id: string }>("/api/v1/sales-orders", {
      method: "POST",
      body: {
        customer_name: customer.name.trim(),
        customer_id_type: customer.id_type,
        customer_id_number: customer.id_number.trim(),
        notes: customer.notes.trim(),
        lines: cart.value.map((l) => ({
          product_id: l.productId,
          description: l.description,
          quantity: l.quantity,
          unit_price: l.unitPrice,
        })),
      },
    });
    router.push(`/sales-orders/${order.id}`);
  } catch (err: unknown) {
    const e = err as { data?: { detail?: string } };
    saveError.value = e?.data?.detail ?? "Error al crear la orden";
    saving.value = false;
  }
}

async function handleMobileSubmit({
  lines,
  customer: c,
}: {
  lines: {
    productId: string;
    description: string;
    unitPrice: number;
    quantity: number;
  }[];
  customer: { name: string; id_type: string; id_number: string; notes: string };
}) {
  saving.value = true;
  saveError.value = "";
  try {
    const order = await useApiFetch<{ id: string }>("/api/v1/sales-orders", {
      method: "POST",
      body: {
        customer_name: c.name.trim(),
        customer_id_type: c.id_type,
        customer_id_number: c.id_number.trim(),
        notes: c.notes.trim(),
        lines: lines.map((l) => ({
          product_id: l.productId,
          description: l.description,
          quantity: l.quantity,
          unit_price: l.unitPrice,
        })),
      },
    });
    router.push(`/sales-orders/${order.id}`);
  } catch (err: unknown) {
    const e = err as { data?: { detail?: string } };
    saveError.value = e?.data?.detail ?? "Error al crear la orden";
    saving.value = false;
  }
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function fmtCurrency(n: number) {
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
}

// ─── Load ─────────────────────────────────────────────────────────────────────

onMounted(async () => {
  loadingProducts.value = true;
  try {
    const all = await useApiFetch<Product[]>("/api/v1/products");
    products.value = all.filter((p) => p.active);
  } finally {
    loadingProducts.value = false;
  }
});
</script>

<template>
  <MobileScreensSalesOrderNew
    v-if="isMobile"
    :products="products"
    :loading-products="loadingProducts"
    :saving="saving"
    :save-error="saveError"
    @submit="handleMobileSubmit"
  />
  <div v-else class="flex flex-col h-full min-h-screen bg-background">
    <!-- ─── Top bar ──────────────────────────────────────────────────────────── -->
    <div
      class="flex items-center justify-between px-4 sm:px-6 py-4 border-b bg-card shadow-sm sticky top-0 z-20"
    >
      <div class="flex items-center gap-3">
        <button
          type="button"
          class="p-2 rounded-lg hover:bg-muted transition-colors text-muted-foreground cursor-pointer"
          @click="router.push('/sales-orders')"
        >
          <ArrowLeft class="w-4 h-4" />
        </button>
        <div>
          <h1 class="text-lg font-bold font-heading text-foreground">
            Nueva orden de venta
          </h1>
          <p class="text-xs text-muted-foreground hidden sm:block">
            Seleccioná los productos y completá los datos del cliente
          </p>
        </div>
      </div>

      <!-- Cart chip (mobile) -->
      <div
        v-if="cartCount > 0"
        class="sm:hidden flex items-center gap-2 bg-primary/10 border border-primary/20 text-primary text-sm font-semibold px-3 py-1.5 rounded-full"
      >
        <ShoppingCart class="w-4 h-4" />
        {{ cartCount }}
      </div>
    </div>

    <!-- ─── Body: 2-column split ─────────────────────────────────────────────── -->
    <div class="flex flex-1 overflow-hidden">
      <!-- ── Left: product catalog ─────────────────────────────────────────── -->
      <div
        class="flex-1 flex flex-col overflow-hidden border-r"
        style="max-width: 62%"
      >
        <!-- Search -->
        <div class="px-4 sm:px-6 py-4 border-b bg-card">
          <div class="relative">
            <Search
              class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground"
            />
            <input
              v-model="searchQuery"
              autofocus
              class="w-full h-10 pl-9 pr-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
              placeholder="Buscar producto por nombre o SKU..."
            />
          </div>
        </div>

        <!-- Product grid -->
        <div class="flex-1 overflow-y-auto px-4 sm:px-6 py-4">
          <!-- Loading -->
          <div
            v-if="loadingProducts"
            class="flex items-center justify-center py-20 text-muted-foreground"
          >
            <Loader2 class="w-6 h-6 animate-spin mr-2" />
            Cargando productos...
          </div>

          <!-- Empty search -->
          <div
            v-else-if="filteredProducts.length === 0"
            class="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3"
          >
            <Package class="w-10 h-10 opacity-30" />
            <p class="text-sm">
              {{
                searchQuery
                  ? "Sin resultados para esa búsqueda"
                  : "No hay productos activos"
              }}
            </p>
          </div>

          <!-- Grid -->
          <div
            v-else
            class="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-4 gap-3"
          >
            <button
              v-for="product in filteredProducts"
              :key="product.id"
              type="button"
              class="relative text-left border rounded-xl bg-card p-3 hover:border-primary/50 hover:shadow-md transition-all duration-150 group cursor-pointer"
              @click="addToCart(product)"
            >
              <!-- Product image or placeholder -->
              <div
                class="w-full aspect-square rounded-lg mb-3 overflow-hidden bg-gradient-to-br from-primary/10 to-primary/5 flex items-center justify-center"
              >
                <img
                  v-if="product.image_url"
                  :src="product.image_url"
                  :alt="product.name"
                  class="w-full h-full object-cover"
                />
                <Package v-else class="w-8 h-8 text-primary/40" />
              </div>

              <!-- In-cart badge -->
              <div
                v-if="cartQtyFor(product.id) > 0"
                class="absolute top-2 right-2 w-6 h-6 rounded-full bg-primary text-white text-xs font-bold flex items-center justify-center shadow"
              >
                {{ cartQtyFor(product.id) }}
              </div>

              <!-- Add overlay on hover -->
              <div
                class="absolute inset-0 rounded-xl bg-primary/5 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center"
              >
                <div
                  class="w-9 h-9 rounded-full bg-primary text-white flex items-center justify-center shadow-lg"
                >
                  <Plus class="w-5 h-5" />
                </div>
              </div>

              <p
                class="text-sm font-semibold text-foreground leading-snug line-clamp-2 mb-1"
              >
                {{ product.name }}
              </p>
              <p class="text-xs text-muted-foreground font-mono">
                $
                {{
                  fmtCurrency(
                    product.internal_price ?? product.fiscal_price ?? 0,
                  )
                }}
              </p>
            </button>
          </div>
        </div>
      </div>

      <!-- ── Right: cart + checkout ─────────────────────────────────────────── -->
      <div
        class="w-full flex flex-col bg-card"
        style="min-width: 38%; max-width: 38%"
      >
        <!-- Cart header -->
        <div class="px-5 py-4 border-b flex items-center justify-between">
          <div class="flex items-center gap-2">
            <ShoppingCart class="w-4 h-4 text-primary" />
            <span class="text-sm font-bold text-foreground">Carrito</span>
          </div>
          <span
            v-if="cartCount > 0"
            class="text-xs text-muted-foreground font-medium"
          >
            {{ cartCount }} ítem{{ cartCount !== 1 ? "s" : "" }}
          </span>
        </div>

        <div class="flex-1 overflow-y-auto flex flex-col">
          <!-- Empty cart -->
          <div
            v-if="cart.length === 0"
            class="flex-1 flex flex-col items-center justify-center text-muted-foreground gap-3 px-6 py-10"
          >
            <ShoppingCart class="w-10 h-10 opacity-20" />
            <p class="text-sm text-center">
              Hacé clic en un producto para agregarlo
            </p>
          </div>

          <!-- Cart lines -->
          <div v-else class="flex-1 divide-y divide-border overflow-y-auto">
            <div
              v-for="(line, idx) in cart"
              :key="line.productId"
              class="px-5 py-3 flex items-center gap-3"
            >
              <!-- Name + price -->
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-foreground truncate">
                  {{ line.description }}
                </p>
                <p class="text-xs text-muted-foreground font-mono mt-0.5">
                  ${{ fmtCurrency(line.unitPrice) }} c/u
                </p>
              </div>

              <!-- Qty controls -->
              <div class="flex items-center gap-1 shrink-0">
                <button
                  type="button"
                  class="w-7 h-7 rounded-lg border border-input hover:bg-muted flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                  @click="decrement(idx)"
                >
                  <Minus class="w-3 h-3" />
                </button>
                <span
                  class="w-8 text-center text-sm font-semibold tabular-nums text-foreground"
                >
                  {{ line.quantity }}
                </span>
                <button
                  type="button"
                  class="w-7 h-7 rounded-lg border border-input hover:bg-muted flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                  @click="increment(idx)"
                >
                  <Plus class="w-3 h-3" />
                </button>
              </div>

              <!-- Subtotal -->
              <p
                class="text-sm font-semibold font-mono text-foreground w-20 text-right shrink-0"
              >
                ${{ fmtCurrency(line.quantity * line.unitPrice) }}
              </p>

              <!-- Remove -->
              <button
                type="button"
                class="text-muted-foreground hover:text-red-500 transition-colors p-1 rounded cursor-pointer"
                @click="removeLine(idx)"
              >
                <Trash2 class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>

          <!-- Customer section -->
          <div class="border-t">
            <button
              type="button"
              class="w-full px-5 py-3 flex items-center justify-between text-sm font-semibold text-foreground hover:bg-muted/40 transition-colors cursor-pointer"
              @click="customerExpanded = !customerExpanded"
            >
              <div class="flex items-center gap-2">
                <User class="w-4 h-4 text-muted-foreground" />
                <span>{{
                  customer.name ? customer.name : "Datos del cliente"
                }}</span>
                <span
                  v-if="customer.id_type === 'anonymous' && !customer.name"
                  class="text-xs font-normal text-muted-foreground"
                  >(anónimo)</span
                >
              </div>
              <svg
                :class="[
                  'w-4 h-4 text-muted-foreground transition-transform duration-200',
                  customerExpanded ? 'rotate-180' : '',
                ]"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M19 9l-7 7-7-7"
                />
              </svg>
            </button>

            <div v-if="customerExpanded" class="px-5 pb-4 space-y-3">
              <div class="space-y-1.5">
                <label class="text-xs font-semibold text-muted-foreground"
                  >Nombre</label
                >
                <input
                  v-model="customer.name"
                  class="w-full h-9 px-3 border border-input rounded-lg text-sm bg-white focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="Nombre del cliente..."
                />
              </div>

              <div class="grid grid-cols-2 gap-2">
                <div class="space-y-1.5">
                  <label class="text-xs font-semibold text-muted-foreground"
                    >Tipo doc.</label
                  >
                  <select
                    v-model="customer.id_type"
                    class="w-full h-9 px-3 border border-input rounded-lg text-sm bg-white focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
                  >
                    <option value="anonymous">Anónimo</option>
                    <option value="cedula">Cédula</option>
                    <option value="rif">RIF</option>
                    <option value="passport">Pasaporte</option>
                  </select>
                </div>
                <div
                  v-if="customer.id_type !== 'anonymous'"
                  class="space-y-1.5"
                >
                  <label class="text-xs font-semibold text-muted-foreground"
                    >Número</label
                  >
                  <input
                    v-model="customer.id_number"
                    class="w-full h-9 px-3 border border-input rounded-lg text-sm bg-white focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                    placeholder="V-12345678"
                  />
                </div>
              </div>

              <div class="space-y-1.5">
                <label class="text-xs font-semibold text-muted-foreground"
                  >Notas</label
                >
                <textarea
                  v-model="customer.notes"
                  rows="2"
                  class="w-full px-3 py-2 border border-input rounded-lg text-sm bg-white focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 resize-none"
                  placeholder="Observaciones..."
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Footer: total + submit -->
        <div class="border-t bg-card px-5 py-4 space-y-3">
          <!-- Error -->
          <div
            v-if="saveError"
            class="bg-red-50 border border-red-200 text-red-700 text-xs rounded-lg px-3 py-2"
          >
            {{ saveError }}
          </div>

          <!-- Total -->
          <div class="flex items-center justify-between">
            <span class="text-sm font-semibold text-muted-foreground"
              >Total</span
            >
            <span class="text-2xl font-bold font-heading text-foreground">
              ${{ fmtCurrency(cartTotal) }}
            </span>
          </div>

          <!-- Submit -->
          <button
            type="button"
            :disabled="saving || cart.length === 0"
            class="w-full flex items-center justify-center gap-2 h-12 rounded-xl bg-primary text-white text-sm font-bold hover:opacity-90 transition-all duration-200 disabled:opacity-40 cursor-pointer"
            @click="submitOrder"
          >
            <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
            <Check v-else class="w-4 h-4" />
            {{ saving ? "Creando orden..." : "Crear orden" }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
