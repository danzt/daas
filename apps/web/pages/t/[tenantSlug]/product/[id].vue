<script setup lang="ts">
import {
  ArrowLeft,
  ShoppingCart,
  Loader2,
  ChevronRight,
  Star,
  Truck,
  ShieldCheck,
  RotateCcw,
  Sparkles,
  Plus,
  Minus,
  Heart,
  Share2,
  Package,
} from "lucide-vue-next";
import { usePublicFetch } from "~/composables/usePublicFetch";
import { useFormatPrice } from "~/composables/useFormatPrice";

definePageMeta({
  layout: "public",
});

const route = useRoute();
const productId = route.params.id as string;
const tenantSlug = route.params.tenantSlug as string;

interface ShopProduct {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  price: number;
  fiscal_price?: number;
  category?: string;
  stock_qty: number;
  is_fiscal: boolean;
}

interface ListResponse {
  data: ShopProduct[];
}

const product = ref<ShopProduct | null>(null);
const related = ref<ShopProduct[]>([]);
const loading = ref(false);
const notFound = ref(false);
const qty = ref(1);
const selectedThumb = ref(0);
const activeTab = ref<"description" | "specs" | "reviews">("description");

const { format: fmt } = useFormatPrice("VE");

function effectivePrice(p: ShopProduct): number {
  return p.is_fiscal && p.fiscal_price ? p.fiscal_price : p.price;
}

function productGradient(id: string): string {
  const palettes = [
    "from-purple-200 via-violet-100 to-fuchsia-100",
    "from-blue-200 via-sky-100 to-cyan-100",
    "from-emerald-200 via-teal-100 to-cyan-100",
    "from-amber-200 via-orange-100 to-rose-100",
    "from-rose-200 via-pink-100 to-purple-100",
    "from-indigo-200 via-blue-100 to-cyan-100",
    "from-lime-200 via-emerald-100 to-teal-100",
    "from-fuchsia-200 via-rose-100 to-amber-100",
  ];
  const hash = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  return palettes[hash % palettes.length];
}

function productRating(id: string): { stars: number; count: number } {
  const hash = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  const stars = 3.8 + (hash % 12) / 10;
  const count = 47 + (hash % 400);
  return { stars: Math.round(stars * 10) / 10, count };
}

function productSoldCount(id: string): number {
  const hash = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  return 12 + (hash % 80);
}

// Mock rating distribution
function ratingBars(id: string): { star: number; pct: number }[] {
  const total = productRating(id).count;
  const hash = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  // weighted toward 4/5 stars
  return [
    { star: 5, pct: 50 + (hash % 25) },
    { star: 4, pct: 20 + (hash % 15) },
    { star: 3, pct: 8 + (hash % 8) },
    { star: 2, pct: 3 + (hash % 4) },
    { star: 1, pct: 1 + (hash % 3) },
  ];
}

// Mock reviews
function mockReviews(
  id: string,
): {
  author: string;
  stars: number;
  date: string;
  title: string;
  body: string;
}[] {
  const hash = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  const names = ["María G.", "Carlos R.", "Ana López", "Juan P.", "Sofía M."];
  const titles = [
    "Excelente producto",
    "Muy buena calidad",
    "Lo recomiendo",
    "Llegó rápido",
    "Como esperaba",
  ];
  const bodies = [
    "Súper contenta con la compra, llegó antes de lo esperado y la calidad es muy buena.",
    "Cumple con lo prometido. Buen precio y empaque cuidado.",
    "Ya lo había comprado antes, ahora lo repito. Sigue siendo confiable.",
    "Entrega rápida y el producto coincide con la descripción.",
  ];
  return Array.from({ length: 3 }, (_, i) => ({
    author: names[(hash + i) % names.length],
    stars: 4 + ((hash + i) % 2),
    date: `${((hash + i) % 28) + 1}/05/2026`,
    title: titles[(hash + i) % titles.length],
    body: bodies[(hash + i) % bodies.length],
  }));
}

async function load() {
  loading.value = true;
  try {
    product.value = await usePublicFetch<ShopProduct>(`/products/${productId}`);
  } catch (err: unknown) {
    const e = err as { status?: number; statusCode?: number };
    if (e?.status === 404 || e?.statusCode === 404) {
      notFound.value = true;
    }
  }
}

async function loadRelated() {
  try {
    const resp = await usePublicFetch<ListResponse>(`/products?per_page=20`);
    const all = resp.data ?? [];
    related.value = all.filter((p) => p.id !== productId).slice(0, 4);
  } catch {
    // non-fatal
  } finally {
    loading.value = false;
  }
}

await load();
if (product.value) {
  await loadRelated();
}

useSeoMeta({
  title: () => (product.value ? `${product.value.name} - DaaS` : "Producto"),
  description: () =>
    product.value?.description?.slice(0, 160) ??
    `Comprá ${product.value?.name ?? "este producto"} con envío rápido y pago seguro.`,
  ogTitle: () => product.value?.name ?? "Producto",
  ogDescription: () => product.value?.description?.slice(0, 160) ?? "",
  ogImage: "/og-placeholder.png",
  ogType: "website",
});

function incQty() {
  if (product.value && qty.value < product.value.stock_qty) qty.value++;
}
function decQty() {
  if (qty.value > 1) qty.value--;
}
</script>

<template>
  <div class="bg-background pb-24 lg:pb-0">
    <!-- Loading -->
    <div
      v-if="loading && !product"
      class="max-w-7xl mx-auto px-4 sm:px-6 py-20 flex items-center justify-center text-muted-foreground"
    >
      <Loader2 class="w-6 h-6 animate-spin mr-2" />
      Cargando...
    </div>

    <!-- Not found -->
    <div
      v-else-if="notFound"
      class="max-w-7xl mx-auto px-4 sm:px-6 py-20 text-center"
    >
      <Package class="w-16 h-16 text-muted-foreground/30 mx-auto mb-4" />
      <h1 class="text-2xl font-bold text-foreground mb-2">
        Producto no encontrado
      </h1>
      <p class="text-muted-foreground mb-6">
        El producto que buscás no está disponible.
      </p>
      <NuxtLink
        :to="`/t/${tenantSlug}`"
        class="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-primary text-white font-semibold hover:opacity-90"
      >
        Volver al catálogo
      </NuxtLink>
    </div>

    <!-- Product detail -->
    <template v-else-if="product">
      <!-- Breadcrumb -->
      <div class="border-b bg-card/50">
        <div
          class="max-w-7xl mx-auto px-4 sm:px-6 py-3 flex items-center gap-2 text-xs sm:text-sm text-muted-foreground overflow-x-auto whitespace-nowrap"
        >
          <NuxtLink
            :to="`/t/${tenantSlug}`"
            class="hover:text-primary transition-colors"
          >
            Inicio
          </NuxtLink>
          <ChevronRight class="w-3 h-3 shrink-0" />
          <span class="hover:text-primary cursor-pointer">{{
            product.category ?? "Productos"
          }}</span>
          <ChevronRight class="w-3 h-3 shrink-0" />
          <span class="text-foreground font-medium truncate">{{
            product.name
          }}</span>
        </div>
      </div>

      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-6 lg:py-8">
        <NuxtLink
          :to="`/t/${tenantSlug}`"
          class="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-primary transition-colors mb-4 lg:hidden"
        >
          <ArrowLeft class="w-4 h-4" />
          Volver al catálogo
        </NuxtLink>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 lg:gap-10">
          <!-- ─── Gallery ──────────────────────────────────────────── -->
          <div class="space-y-3">
            <div
              :class="[
                'aspect-square rounded-2xl overflow-hidden border bg-gradient-to-br relative',
                productGradient(product.id),
              ]"
            >
              <div class="absolute inset-0 flex items-center justify-center">
                <ShoppingCart class="w-32 h-32 text-foreground/15" />
              </div>

              <!-- Badges -->
              <div class="absolute top-4 left-4 flex flex-col gap-2">
                <span
                  v-if="product.is_fiscal"
                  class="inline-flex items-center px-3 py-1 rounded-full bg-primary text-white text-xs font-bold shadow"
                >
                  FISCAL
                </span>
                <span
                  v-if="product.stock_qty > 0 && product.stock_qty <= 5"
                  class="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-amber-500 text-white text-xs font-bold shadow"
                >
                  <Sparkles class="w-3 h-3" />
                  ¡Pocas unidades!
                </span>
              </div>

              <!-- Wishlist / share -->
              <div class="absolute top-4 right-4 flex flex-col gap-2">
                <button
                  type="button"
                  aria-label="Guardar en favoritos"
                  class="w-9 h-9 rounded-full bg-white/90 backdrop-blur shadow flex items-center justify-center hover:bg-white text-foreground transition-colors cursor-pointer"
                >
                  <Heart class="w-4 h-4" />
                </button>
                <button
                  type="button"
                  aria-label="Compartir"
                  class="w-9 h-9 rounded-full bg-white/90 backdrop-blur shadow flex items-center justify-center hover:bg-white text-foreground transition-colors cursor-pointer"
                >
                  <Share2 class="w-4 h-4" />
                </button>
              </div>
            </div>

            <!-- Thumbnails (mock — same gradient, different deco) -->
            <div class="grid grid-cols-4 gap-2 sm:gap-3">
              <button
                v-for="i in 4"
                :key="i"
                type="button"
                :class="[
                  'aspect-square rounded-lg border-2 overflow-hidden bg-gradient-to-br relative transition-all cursor-pointer',
                  productGradient(product.id),
                  selectedThumb === i - 1
                    ? 'border-primary ring-2 ring-primary/20'
                    : 'border-border hover:border-primary/40',
                ]"
                @click="selectedThumb = i - 1"
              >
                <div class="absolute inset-0 flex items-center justify-center">
                  <ShoppingCart class="w-6 h-6 text-foreground/20" />
                </div>
              </button>
            </div>
          </div>

          <!-- ─── Info ─────────────────────────────────────────────── -->
          <div class="space-y-4">
            <div>
              <p
                v-if="product.category"
                class="text-xs uppercase tracking-wider text-muted-foreground mb-1 font-semibold"
              >
                {{ product.category }}
              </p>
              <h1
                class="text-2xl sm:text-3xl lg:text-4xl font-bold font-heading text-foreground leading-tight"
              >
                {{ product.name }}
              </h1>
            </div>

            <!-- Rating -->
            <div class="flex items-center gap-3 flex-wrap">
              <div class="flex items-center gap-1.5">
                <div class="flex">
                  <Star
                    v-for="i in 5"
                    :key="i"
                    :class="[
                      'w-4 h-4',
                      i <= Math.round(productRating(product.id).stars)
                        ? 'fill-amber-400 text-amber-400'
                        : 'text-muted-foreground/30',
                    ]"
                  />
                </div>
                <span class="text-sm font-semibold text-foreground">
                  {{ productRating(product.id).stars }}
                </span>
              </div>
              <span class="text-sm text-muted-foreground">
                · {{ productRating(product.id).count }} reseñas
              </span>
              <span class="text-sm text-muted-foreground">
                · {{ productSoldCount(product.id) }} vendidos este mes
              </span>
            </div>

            <!-- Price -->
            <div class="border-y py-4 my-4">
              <div class="flex items-baseline gap-3 flex-wrap">
                <span
                  class="text-3xl sm:text-4xl font-bold font-mono text-foreground"
                >
                  {{ fmt(effectivePrice(product)) }}
                </span>
                <span
                  v-if="product.is_fiscal"
                  class="text-sm text-muted-foreground"
                  >+ IVA incluido</span
                >
              </div>
              <p
                v-if="product.stock_qty > 5"
                class="text-sm text-emerald-600 mt-2 flex items-center gap-1.5 font-medium"
              >
                <ShieldCheck class="w-4 h-4" />
                Disponible — Envío rápido
              </p>
              <p
                v-else-if="product.stock_qty > 0"
                class="text-sm text-amber-600 mt-2 flex items-center gap-1.5 font-medium"
              >
                <Sparkles class="w-4 h-4" />
                ¡Solo quedan {{ product.stock_qty }} unidades!
              </p>
              <p v-else class="text-sm text-rose-600 mt-2 font-medium">
                Agotado temporalmente
              </p>
            </div>

            <!-- Qty + Add to cart -->
            <div class="space-y-3">
              <div class="flex items-center gap-3">
                <span class="text-sm font-medium text-foreground"
                  >Cantidad:</span
                >
                <div class="inline-flex items-center border rounded-lg">
                  <button
                    type="button"
                    aria-label="Disminuir cantidad"
                    :disabled="qty <= 1"
                    class="w-9 h-9 flex items-center justify-center hover:bg-muted disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
                    @click="decQty"
                  >
                    <Minus class="w-3.5 h-3.5" />
                  </button>
                  <span class="w-10 text-center text-sm font-semibold">{{
                    qty
                  }}</span>
                  <button
                    type="button"
                    aria-label="Aumentar cantidad"
                    :disabled="qty >= product.stock_qty"
                    class="w-9 h-9 flex items-center justify-center hover:bg-muted disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
                    @click="incQty"
                  >
                    <Plus class="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>

              <button
                type="button"
                :disabled="product.stock_qty === 0"
                class="w-full px-5 py-3.5 rounded-xl bg-primary text-white text-base font-bold flex items-center justify-center gap-2 hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-opacity cursor-pointer shadow-lg shadow-primary/20"
                title="Carrito disponible próximamente"
              >
                <ShoppingCart class="w-5 h-5" />
                Agregar al carrito
              </button>

              <button
                type="button"
                :disabled="product.stock_qty === 0"
                class="w-full px-5 py-3.5 rounded-xl bg-amber-500 text-white text-base font-bold hover:bg-amber-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors cursor-pointer"
                title="Checkout disponible próximamente"
              >
                Comprar ahora
              </button>
            </div>

            <!-- Trust badges -->
            <div
              class="grid grid-cols-1 sm:grid-cols-3 gap-2 sm:gap-3 pt-4 border-t"
            >
              <div
                class="flex items-center gap-2 text-xs text-muted-foreground"
              >
                <div
                  class="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center shrink-0"
                >
                  <Truck class="w-4 h-4 text-primary" />
                </div>
                <div>
                  <p class="font-semibold text-foreground">Envío rápido</p>
                  <p>1-3 días hábiles</p>
                </div>
              </div>
              <div
                class="flex items-center gap-2 text-xs text-muted-foreground"
              >
                <div
                  class="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center shrink-0"
                >
                  <ShieldCheck class="w-4 h-4 text-primary" />
                </div>
                <div>
                  <p class="font-semibold text-foreground">Pago seguro</p>
                  <p>Compra protegida</p>
                </div>
              </div>
              <div
                class="flex items-center gap-2 text-xs text-muted-foreground"
              >
                <div
                  class="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center shrink-0"
                >
                  <RotateCcw class="w-4 h-4 text-primary" />
                </div>
                <div>
                  <p class="font-semibold text-foreground">Devolución gratis</p>
                  <p>Hasta 7 días</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- ─── Tabs: Descripción / Specs / Reviews ───────────────── -->
        <div class="mt-10 lg:mt-14 border rounded-2xl bg-card overflow-hidden">
          <div class="flex border-b">
            <button
              v-for="tab in [
                { id: 'description', label: 'Descripción' },
                { id: 'specs', label: 'Especificaciones' },
                {
                  id: 'reviews',
                  label: `Reseñas (${productRating(product.id).count})`,
                },
              ]"
              :key="tab.id"
              type="button"
              :class="[
                'px-5 py-3 text-sm font-semibold transition-colors cursor-pointer relative',
                activeTab === tab.id
                  ? 'text-primary'
                  : 'text-muted-foreground hover:text-foreground',
              ]"
              @click="activeTab = tab.id as typeof activeTab"
            >
              {{ tab.label }}
              <span
                v-if="activeTab === tab.id"
                class="absolute inset-x-3 -bottom-px h-0.5 bg-primary rounded-full"
              />
            </button>
          </div>

          <div class="p-5 sm:p-6">
            <!-- Description -->
            <div
              v-if="activeTab === 'description'"
              class="prose prose-sm max-w-none"
            >
              <p
                class="text-sm text-foreground/80 leading-relaxed whitespace-pre-line"
              >
                {{
                  product.description ||
                  `${product.name} — un producto de calidad seleccionado por nuestro equipo. Diseño cuidado, materiales premium y la confianza de una tienda con miles de clientes satisfechos.\n\nIdeal para tu día a día o como regalo para alguien especial.`
                }}
              </p>
            </div>

            <!-- Specs -->
            <div v-else-if="activeTab === 'specs'">
              <dl
                class="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-3 text-sm"
              >
                <div class="flex justify-between border-b pb-2">
                  <dt class="text-muted-foreground">Categoría</dt>
                  <dd class="font-medium text-foreground">
                    {{ product.category ?? "—" }}
                  </dd>
                </div>
                <div class="flex justify-between border-b pb-2">
                  <dt class="text-muted-foreground">Tipo</dt>
                  <dd class="font-medium text-foreground">
                    {{ product.is_fiscal ? "Fiscal (con IVA)" : "Interno" }}
                  </dd>
                </div>
                <div class="flex justify-between border-b pb-2">
                  <dt class="text-muted-foreground">Stock disponible</dt>
                  <dd class="font-medium text-foreground">
                    {{ product.stock_qty }} unidades
                  </dd>
                </div>
                <div class="flex justify-between border-b pb-2">
                  <dt class="text-muted-foreground">SKU</dt>
                  <dd class="font-mono text-xs text-foreground">
                    {{ product.id.slice(0, 8) }}...
                  </dd>
                </div>
              </dl>
            </div>

            <!-- Reviews -->
            <div v-else-if="activeTab === 'reviews'" class="space-y-6">
              <!-- Summary -->
              <div
                class="grid grid-cols-1 sm:grid-cols-[200px_1fr] gap-6 pb-5 border-b"
              >
                <div class="text-center sm:text-left">
                  <div class="text-4xl font-bold text-foreground font-mono">
                    {{ productRating(product.id).stars }}
                  </div>
                  <div
                    class="flex items-center justify-center sm:justify-start mt-1"
                  >
                    <Star
                      v-for="i in 5"
                      :key="i"
                      :class="[
                        'w-4 h-4',
                        i <= Math.round(productRating(product.id).stars)
                          ? 'fill-amber-400 text-amber-400'
                          : 'text-muted-foreground/30',
                      ]"
                    />
                  </div>
                  <p class="text-xs text-muted-foreground mt-1">
                    Basado en {{ productRating(product.id).count }} reseñas
                  </p>
                </div>
                <div class="space-y-1.5">
                  <div
                    v-for="bar in ratingBars(product.id)"
                    :key="bar.star"
                    class="flex items-center gap-2 text-xs"
                  >
                    <span class="w-6 text-muted-foreground"
                      >{{ bar.star }}★</span
                    >
                    <div
                      class="flex-1 h-2 rounded-full bg-muted overflow-hidden"
                    >
                      <div
                        class="h-full bg-amber-400"
                        :style="{ width: `${Math.min(bar.pct, 100)}%` }"
                      />
                    </div>
                    <span class="w-10 text-right text-muted-foreground"
                      >{{ Math.min(bar.pct, 100) }}%</span
                    >
                  </div>
                </div>
              </div>

              <!-- Reviews list -->
              <div class="space-y-5">
                <div
                  v-for="(review, idx) in mockReviews(product.id)"
                  :key="idx"
                  class="border-b pb-5 last:border-b-0"
                >
                  <div class="flex items-start gap-3">
                    <div
                      class="w-9 h-9 rounded-full bg-primary/10 flex items-center justify-center shrink-0 text-primary font-bold text-sm"
                    >
                      {{ review.author.charAt(0) }}
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center gap-2 flex-wrap">
                        <p class="text-sm font-semibold text-foreground">
                          {{ review.author }}
                        </p>
                        <span class="text-xs text-muted-foreground"
                          >· {{ review.date }}</span
                        >
                      </div>
                      <div class="flex items-center gap-1 mt-0.5">
                        <Star
                          v-for="i in 5"
                          :key="i"
                          :class="[
                            'w-3.5 h-3.5',
                            i <= review.stars
                              ? 'fill-amber-400 text-amber-400'
                              : 'text-muted-foreground/30',
                          ]"
                        />
                      </div>
                      <p class="text-sm font-medium text-foreground mt-2">
                        {{ review.title }}
                      </p>
                      <p class="text-sm text-muted-foreground mt-1">
                        {{ review.body }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- ─── Related products ─────────────────────────────────── -->
        <section v-if="related.length > 0" class="mt-12">
          <div class="flex items-end justify-between mb-5">
            <div>
              <h2
                class="text-xl sm:text-2xl font-bold font-heading text-foreground"
              >
                También te puede interesar
              </h2>
              <p class="text-sm text-muted-foreground mt-0.5">
                Productos relacionados de la misma tienda
              </p>
            </div>
            <NuxtLink
              :to="`/t/${tenantSlug}`"
              class="text-sm font-semibold text-primary hover:underline hidden sm:inline-flex items-center gap-1"
            >
              Ver todos
              <ChevronRight class="w-4 h-4" />
            </NuxtLink>
          </div>

          <div
            class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3 sm:gap-4"
          >
            <NuxtLink
              v-for="rel in related"
              :key="rel.id"
              :to="`/t/${tenantSlug}/product/${rel.id}`"
              class="group border bg-card rounded-xl overflow-hidden hover:shadow-lg hover:border-primary/30 hover:-translate-y-0.5 transition-all duration-300 cursor-pointer flex flex-col"
            >
              <div
                :class="[
                  'aspect-square bg-gradient-to-br relative',
                  productGradient(rel.id),
                ]"
              >
                <div class="absolute inset-0 flex items-center justify-center">
                  <ShoppingCart
                    class="w-12 h-12 text-foreground/15 group-hover:scale-110 transition-transform"
                  />
                </div>
                <span
                  v-if="rel.is_fiscal"
                  class="absolute top-2 left-2 px-2 py-0.5 rounded-full bg-primary text-white text-[10px] font-bold"
                >
                  FISCAL
                </span>
              </div>
              <div class="p-3 flex flex-col flex-1">
                <h3
                  class="text-sm font-semibold text-foreground line-clamp-2 group-hover:text-primary transition-colors min-h-[2.5rem]"
                >
                  {{ rel.name }}
                </h3>
                <div class="flex items-center gap-1 mt-1.5">
                  <Star
                    v-for="i in 5"
                    :key="i"
                    :class="[
                      'w-3 h-3',
                      i <= Math.round(productRating(rel.id).stars)
                        ? 'fill-amber-400 text-amber-400'
                        : 'text-muted-foreground/30',
                    ]"
                  />
                  <span class="text-[10px] text-muted-foreground ml-0.5"
                    >({{ productRating(rel.id).count }})</span
                  >
                </div>
                <div class="flex-1" />
                <p class="text-base font-bold font-mono text-foreground mt-2">
                  {{ fmt(effectivePrice(rel)) }}
                </p>
              </div>
            </NuxtLink>
          </div>
        </section>
      </div>

      <!-- ─── Sticky mobile CTA ─────────────────────────────────── -->
      <div
        class="fixed bottom-0 inset-x-0 lg:hidden bg-card border-t shadow-lg z-30 px-4 py-3 safe-bottom"
      >
        <div class="flex items-center gap-3">
          <div class="flex-1">
            <p class="text-xs text-muted-foreground">Precio</p>
            <p class="text-base font-bold font-mono text-foreground">
              {{ fmt(effectivePrice(product)) }}
            </p>
          </div>
          <button
            type="button"
            :disabled="product.stock_qty === 0"
            class="flex-1 px-4 py-3 rounded-xl bg-primary text-white text-sm font-bold flex items-center justify-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
          >
            <ShoppingCart class="w-4 h-4" />
            Agregar
          </button>
        </div>
      </div>
    </template>
  </div>
</template>
