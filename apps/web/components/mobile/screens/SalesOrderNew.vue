<script setup lang="ts">
import {
  Package,
  Minus,
  Plus,
  Trash2,
  ShoppingCart,
  Loader2,
  Check,
  User,
  ChevronDown,
} from "lucide-vue-next";
import type { Product } from "~/components/products/ProductFormModal.vue";

interface CartLine {
  productId: string;
  description: string;
  unitPrice: number;
  quantity: number;
}

interface CustomerData {
  name: string;
  id_type: string;
  id_number: string;
  notes: string;
}

const props = defineProps<{
  products: Product[];
  loadingProducts: boolean;
  saving: boolean;
  saveError: string;
}>();

const emit = defineEmits<{
  submit: [{ lines: CartLine[]; customer: CustomerData }];
}>();

// ── State ──────────────────────────────────────────────────────────────────────
const activeTab = ref<"products" | "cart">("products");
const searchQuery = ref("");

const cart = ref<CartLine[]>([]);
const customerExpanded = ref(false);
const customer = reactive<CustomerData>({
  name: "",
  id_type: "anonymous",
  id_number: "",
  notes: "",
});

// ── Computed ───────────────────────────────────────────────────────────────────
const filteredProducts = computed(() => {
  const q = searchQuery.value.toLowerCase().trim();
  if (!q) return props.products;
  return props.products.filter(
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

// ── Cart actions ───────────────────────────────────────────────────────────────
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

// ── Submit ─────────────────────────────────────────────────────────────────────
function handleSubmit() {
  emit("submit", { lines: [...cart.value], customer: { ...customer } });
}

// ── Helpers ────────────────────────────────────────────────────────────────────
function fmtCurrency(n: number) {
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
}

const tabOptions = computed(() => [
  { key: "products", label: "Productos" },
  { key: "cart", label: "Carrito", count: cartCount.value || undefined },
]);
</script>

<template>
  <MobileScreen title="Nueva venta" back subtitle="POS">
    <!-- Tab selector -->
    <div class="sticky top-[3.75rem] z-10 border-b border-border bg-background">
      <MobileSegment v-model="activeTab" :options="tabOptions" />
    </div>

    <!-- ── Products tab ────────────────────────────────────────────────────────── -->
    <div v-show="activeTab === 'products'">
      <MobileSearchBar
        v-model="searchQuery"
        placeholder="Buscar por nombre o SKU…"
      />

      <!-- Loading -->
      <div v-if="loadingProducts" class="flex justify-center py-20">
        <Loader2 class="size-6 animate-spin text-muted-foreground" />
      </div>

      <!-- Empty -->
      <div
        v-else-if="filteredProducts.length === 0"
        class="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3"
      >
        <Package class="size-10 opacity-30" />
        <p class="text-sm">
          {{ searchQuery ? "Sin resultados" : "No hay productos activos" }}
        </p>
      </div>

      <!-- Grid -->
      <div v-else class="grid grid-cols-2 gap-3 p-4 pb-8">
        <button
          v-for="product in filteredProducts"
          :key="product.id"
          type="button"
          class="relative text-left rounded-2xl border border-border bg-white p-3 active:scale-95 transition-transform duration-100"
          @click="addToCart(product)"
        >
          <!-- Product image / placeholder -->
          <div
            class="w-full aspect-square rounded-xl mb-3 overflow-hidden bg-gradient-to-br from-primary/10 to-primary/5 flex items-center justify-center"
          >
            <img
              v-if="product.image_url"
              :src="product.image_url"
              :alt="product.name"
              class="w-full h-full object-cover"
            />
            <Package v-else class="size-8 text-primary/40" />
          </div>

          <!-- In-cart badge -->
          <div
            v-if="cartQtyFor(product.id) > 0"
            class="absolute top-2 right-2 flex size-6 items-center justify-center rounded-full bg-primary text-xs font-bold text-white shadow"
          >
            {{ cartQtyFor(product.id) }}
          </div>

          <p
            class="line-clamp-2 text-sm font-semibold leading-snug text-foreground mb-1"
          >
            {{ product.name }}
          </p>
          <p class="font-mono text-xs text-muted-foreground">
            ${{
              fmtCurrency(product.internal_price ?? product.fiscal_price ?? 0)
            }}
          </p>
        </button>
      </div>

      <!-- Go to cart CTA -->
      <div
        v-if="cartCount > 0"
        class="fixed inset-x-0 bottom-0 z-40 border-t border-border bg-white px-4 pb-[calc(0.75rem+env(safe-area-inset-bottom))] pt-3"
      >
        <button
          type="button"
          class="no-min-tap flex h-14 w-full items-center justify-between rounded-2xl bg-primary px-5 text-white"
          @click="activeTab = 'cart'"
        >
          <div class="flex items-center gap-2 text-sm font-bold">
            <ShoppingCart class="size-5" />
            Ver carrito
          </div>
          <div class="text-right">
            <span class="text-lg font-black tabular-nums">
              ${{ fmtCurrency(cartTotal) }}
            </span>
            <span class="ml-2 text-xs opacity-80">
              {{ cartCount }} ítem{{ cartCount !== 1 ? "s" : "" }}
            </span>
          </div>
        </button>
      </div>

      <div v-if="cartCount > 0" class="h-20" />
    </div>

    <!-- ── Cart tab ──────────────────────────────────────────────────────────────── -->
    <div v-show="activeTab === 'cart'">
      <!-- Empty cart -->
      <div
        v-if="cart.length === 0"
        class="flex flex-col items-center justify-center py-24 gap-3 text-muted-foreground"
      >
        <ShoppingCart class="size-12 opacity-20" />
        <p class="text-sm text-center px-8">
          Agregá productos desde la pestaña Productos
        </p>
        <button
          type="button"
          class="mt-2 h-10 rounded-xl bg-primary px-6 text-sm font-bold text-white active:opacity-90"
          @click="activeTab = 'products'"
        >
          Ir a productos
        </button>
      </div>

      <template v-else>
        <!-- Lines -->
        <MobileSectionHeader title="Productos" class="pt-4" />
        <div class="divide-y divide-border border-y border-border bg-white">
          <div
            v-for="(line, idx) in cart"
            :key="line.productId"
            class="flex items-center gap-3 px-4 py-3"
          >
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium text-foreground">
                {{ line.description }}
              </p>
              <p class="font-mono text-[11px] text-muted-foreground">
                ${{ fmtCurrency(line.unitPrice) }} c/u
              </p>
            </div>

            <!-- Qty controls -->
            <div class="flex items-center gap-1 shrink-0">
              <button
                type="button"
                class="no-min-tap flex size-8 items-center justify-center rounded-xl border border-border text-muted-foreground active:bg-accent"
                @click="decrement(idx)"
              >
                <Minus class="size-3.5" />
              </button>
              <span
                class="w-8 text-center text-sm font-bold tabular-nums text-foreground"
              >
                {{ line.quantity }}
              </span>
              <button
                type="button"
                class="no-min-tap flex size-8 items-center justify-center rounded-xl border border-border text-muted-foreground active:bg-accent"
                @click="increment(idx)"
              >
                <Plus class="size-3.5" />
              </button>
            </div>

            <span
              class="w-16 text-right font-mono text-sm font-bold tabular-nums text-foreground shrink-0"
            >
              ${{ fmtCurrency(line.quantity * line.unitPrice) }}
            </span>

            <button
              type="button"
              class="no-min-tap shrink-0 p-1 text-muted-foreground active:text-red-500"
              @click="removeLine(idx)"
            >
              <Trash2 class="size-4" />
            </button>
          </div>

          <div class="flex items-center justify-between px-4 py-3.5">
            <span class="text-sm font-bold text-foreground">Total</span>
            <span
              class="font-mono text-xl font-black tabular-nums text-primary"
            >
              ${{ fmtCurrency(cartTotal) }}
            </span>
          </div>
        </div>

        <!-- Customer -->
        <MobileSectionHeader title="Cliente" class="pt-5" />
        <div class="border-y border-border bg-white">
          <button
            type="button"
            class="no-min-tap flex w-full items-center justify-between px-4 py-3.5 text-left active:bg-accent"
            @click="customerExpanded = !customerExpanded"
          >
            <div class="flex items-center gap-2">
              <User class="size-4 text-muted-foreground" />
              <span class="text-sm font-semibold text-foreground">
                {{ customer.name || "Consumidor anónimo" }}
              </span>
            </div>
            <ChevronDown
              class="size-4 text-muted-foreground transition-transform duration-200"
              :class="customerExpanded ? 'rotate-180' : ''"
            />
          </button>

          <div
            v-if="customerExpanded"
            class="border-t border-border px-4 pb-4 pt-3 space-y-3"
          >
            <div class="space-y-1">
              <label class="text-xs font-semibold text-muted-foreground"
                >Nombre</label
              >
              <input
                v-model="customer.name"
                class="h-10 w-full rounded-xl border border-border px-3 text-sm bg-white focus:border-primary focus:outline-none"
                placeholder="Nombre del cliente…"
              />
            </div>

            <div class="grid grid-cols-2 gap-2">
              <div class="space-y-1">
                <label class="text-xs font-semibold text-muted-foreground"
                  >Tipo doc.</label
                >
                <select
                  v-model="customer.id_type"
                  class="h-10 w-full rounded-xl border border-border px-3 text-sm bg-white focus:border-primary focus:outline-none"
                >
                  <option value="anonymous">Anónimo</option>
                  <option value="cedula">Cédula</option>
                  <option value="rif">RIF</option>
                  <option value="passport">Pasaporte</option>
                </select>
              </div>
              <div v-if="customer.id_type !== 'anonymous'" class="space-y-1">
                <label class="text-xs font-semibold text-muted-foreground"
                  >Número</label
                >
                <input
                  v-model="customer.id_number"
                  class="h-10 w-full rounded-xl border border-border px-3 text-sm bg-white focus:border-primary focus:outline-none"
                  placeholder="V-12345678"
                />
              </div>
            </div>

            <div class="space-y-1">
              <label class="text-xs font-semibold text-muted-foreground"
                >Notas</label
              >
              <textarea
                v-model="customer.notes"
                rows="2"
                class="w-full rounded-xl border border-border px-3 py-2 text-sm bg-white resize-none focus:border-primary focus:outline-none"
                placeholder="Observaciones…"
              />
            </div>
          </div>
        </div>

        <!-- Save error -->
        <div v-if="saveError" class="px-4 pt-3">
          <p
            class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
          >
            {{ saveError }}
          </p>
        </div>

        <!-- Spacer for action bar -->
        <div class="h-24" />
      </template>

      <!-- Action bar (cart tab, cart has items) -->
      <div
        v-if="cart.length > 0"
        class="fixed inset-x-0 bottom-0 z-40 border-t border-border bg-white px-4 pb-[calc(0.75rem+env(safe-area-inset-bottom))] pt-3"
      >
        <button
          type="button"
          :disabled="saving"
          class="no-min-tap flex h-14 w-full items-center justify-center gap-2 rounded-2xl bg-primary text-sm font-bold text-white active:opacity-90 disabled:opacity-40"
          @click="handleSubmit"
        >
          <Loader2 v-if="saving" class="size-5 animate-spin" />
          <Check v-else class="size-5" />
          {{
            saving
              ? "Creando orden…"
              : `Crear orden · $${fmtCurrency(cartTotal)}`
          }}
        </button>
      </div>
    </div>
  </MobileScreen>
</template>
