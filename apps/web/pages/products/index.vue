<script setup lang="ts">
import {
  Package,
  Search,
  Plus,
  Pencil,
  Trash2,
  AlertTriangle,
  Loader2,
} from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";
import { useApiFetch } from "~/composables/useAuth";
import ProductFormModal from "~/components/products/ProductFormModal.vue";
import type {
  Product,
  Category,
} from "~/components/products/ProductFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

const store = useAuthStore();

// ─── Data ────────────────────────────────────────────────────────────────────
const products = ref<Product[]>([]);
const categories = ref<Category[]>([]);
const loading = ref(false);
const loadError = ref("");

// ─── Filters ─────────────────────────────────────────────────────────────────
const searchQuery = ref("");
const categoryFilter = ref("");
const typeFilter = ref("all"); // "all" | "fiscal" | "interno"
const statusFilter = ref("all"); // "all" | "active" | "inactive"

// ─── Modal state ─────────────────────────────────────────────────────────────
const formModalOpen = ref(false);
const editingProduct = ref<Product | undefined>(undefined);

// ─── Delete state ────────────────────────────────────────────────────────────
const deletingId = ref<string | null>(null);
const confirmDeleteId = ref<string | null>(null);

// ─── Stats ───────────────────────────────────────────────────────────────────
const totalProducts = computed(() => products.value.length);
const fiscalCount = computed(
  () => products.value.filter((p) => p.is_fiscal).length,
);
const internalCount = computed(
  () => products.value.filter((p) => !p.is_fiscal).length,
);
const inactiveCount = computed(
  () => products.value.filter((p) => !p.active).length,
);

// ─── Filtered list ────────────────────────────────────────────────────────────
const filteredProducts = computed(() => {
  let list = products.value;

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        (p.sku ?? "").toLowerCase().includes(q),
    );
  }

  if (categoryFilter.value) {
    list = list.filter((p) => p.category_id === categoryFilter.value);
  }

  if (typeFilter.value === "fiscal") {
    list = list.filter((p) => p.is_fiscal);
  } else if (typeFilter.value === "interno") {
    list = list.filter((p) => !p.is_fiscal);
  }

  if (statusFilter.value === "active") {
    list = list.filter((p) => p.active);
  } else if (statusFilter.value === "inactive") {
    list = list.filter((p) => !p.active);
  }

  return list;
});

const categoryOptions = computed(() =>
  categories.value.map((c) => ({ value: c.id, label: c.name })),
);

// ─── API calls ───────────────────────────────────────────────────────────────
async function fetchProducts() {
  loading.value = true;
  loadError.value = "";
  try {
    const data = await useApiFetch<Product[]>("/api/v1/products");
    products.value = data ?? [];
  } catch {
    loadError.value = "Error al cargar los productos";
  } finally {
    loading.value = false;
  }
}

async function fetchCategories() {
  try {
    const data = await useApiFetch<Category[]>("/api/v1/products/categories");
    categories.value = data ?? [];
  } catch {
    // silently — categories filter just won't populate
  }
}

async function deleteProduct(id: string) {
  deletingId.value = id;
  confirmDeleteId.value = null;
  try {
    await useApiFetch(`/api/v1/products/${id}`, { method: "DELETE" });
    products.value = products.value.filter((p) => p.id !== id);
  } catch {
    // show nothing — product stays in list
  } finally {
    deletingId.value = null;
  }
}

// ─── Modal handlers ───────────────────────────────────────────────────────────
function openCreate() {
  editingProduct.value = undefined;
  formModalOpen.value = true;
}

function openEdit(product: Product) {
  editingProduct.value = product;
  formModalOpen.value = true;
}

function onSaved() {
  fetchProducts();
}

// ─── Helpers ─────────────────────────────────────────────────────────────────
function getCategoryName(categoryId?: string): string {
  if (!categoryId) return "—";
  return categories.value.find((c) => c.id === categoryId)?.name ?? "—";
}

function formatPrice(product: Product): string {
  const price = product.is_fiscal
    ? product.fiscal_price
    : product.internal_price;
  if (price == null) return "—";
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(price);
}

// ─── Lifecycle ────────────────────────────────────────────────────────────────
onMounted(async () => {
  await Promise.all([fetchCategories(), fetchProducts()]);
});
</script>

<template>
  <div class="p-4 sm:p-6 space-y-5">
    <!-- Page header -->
    <div class="flex items-center justify-between gap-4">
      <div class="flex items-center gap-3">
        <div
          class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center"
        >
          <Package class="w-5 h-5 text-primary" />
        </div>
        <div>
          <h1 class="text-2xl font-bold font-heading text-foreground">
            Productos
          </h1>
          <p class="text-muted-foreground text-sm">
            Gestioná tu catálogo de productos
          </p>
        </div>
      </div>
      <button
        v-if="store.isOwner"
        type="button"
        class="flex items-center gap-2 h-10 px-5 bg-primary text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer"
        @click="openCreate"
      >
        <Plus class="w-4 h-4" />
        Nuevo Producto
      </button>
    </div>

    <!-- Stats row -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="rounded-xl border bg-card shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Total
        </p>
        <p class="text-2xl font-bold font-heading text-foreground">
          {{ totalProducts }}
        </p>
      </div>
      <div class="rounded-xl border bg-card shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Fiscales
        </p>
        <p class="text-2xl font-bold font-heading text-primary">
          {{ fiscalCount }}
        </p>
      </div>
      <div class="rounded-xl border bg-card shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Internos
        </p>
        <p class="text-2xl font-bold font-heading text-muted-foreground">
          {{ internalCount }}
        </p>
      </div>
      <div class="rounded-xl border bg-card shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Inactivos
        </p>
        <p class="text-2xl font-bold font-heading text-destructive">
          {{ inactiveCount }}
        </p>
      </div>
    </div>

    <!-- Filter bar -->
    <div class="rounded-xl border bg-card shadow-sm px-4 py-3 sm:px-5 sm:py-4">
      <div class="flex flex-col sm:flex-row gap-3 flex-wrap items-center">
        <!-- Search -->
        <div class="relative flex-1 min-w-48">
          <Search
            class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground"
          />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Buscar por nombre o SKU..."
            class="w-full h-10 pl-9 pr-4 border rounded-lg text-sm transition-all duration-200 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-muted-foreground bg-white"
          />
        </div>

        <!-- Category filter -->
        <div class="w-full sm:w-48">
          <Select
            v-model="categoryFilter"
            :options="categoryOptions"
            placeholder="Todas las categorías"
          />
        </div>

        <!-- Status filter -->
        <div class="w-full sm:w-40">
          <Select
            v-model="statusFilter"
            :options="[
              { value: 'all', label: 'Todos los estados' },
              { value: 'active', label: 'Activos' },
              { value: 'inactive', label: 'Inactivos' },
            ]"
            placeholder="Estado"
          />
        </div>

        <!-- Type tabs -->
        <div class="flex bg-muted rounded-lg p-1 gap-1">
          <button
            v-for="tab in [
              { value: 'all', label: 'Todos' },
              { value: 'fiscal', label: 'Fiscal' },
              { value: 'interno', label: 'Interno' },
            ]"
            :key="tab.value"
            type="button"
            :class="[
              'px-3 py-1.5 text-xs font-semibold rounded-md transition-all duration-200 cursor-pointer',
              typeFilter === tab.value
                ? 'bg-white text-primary shadow-sm'
                : 'text-muted-foreground hover:text-foreground',
            ]"
            @click="typeFilter = tab.value"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>
    </div>

    <!-- Loading skeleton -->
    <div
      v-if="loading"
      class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3 sm:gap-4"
    >
      <div
        v-for="n in 8"
        :key="n"
        class="rounded-xl border bg-card shadow-sm overflow-hidden animate-pulse"
      >
        <div class="aspect-square bg-muted" />
        <div class="p-3 space-y-2">
          <div class="h-3 bg-muted rounded w-3/4" />
          <div class="h-3 bg-muted rounded w-1/2" />
          <div class="h-4 bg-muted rounded w-2/3 mt-3" />
        </div>
      </div>
    </div>

    <!-- Error state -->
    <div
      v-else-if="loadError"
      class="rounded-xl border bg-card shadow-sm px-6 py-12 text-center"
    >
      <AlertTriangle class="w-10 h-10 text-destructive mx-auto mb-3" />
      <p class="text-destructive font-medium">{{ loadError }}</p>
      <button
        type="button"
        class="mt-4 h-9 px-4 border text-sm font-semibold text-muted-foreground rounded-lg hover:bg-muted cursor-pointer transition-all duration-200"
        @click="fetchProducts"
      >
        Reintentar
      </button>
    </div>

    <!-- Empty state — no products -->
    <div
      v-else-if="products.length === 0"
      class="rounded-xl border bg-card shadow-sm px-6 py-16 text-center"
    >
      <Package class="w-12 h-12 text-muted-foreground/20 mx-auto mb-4" />
      <h3 class="text-base font-semibold text-muted-foreground mb-1">
        Sin productos aún
      </h3>
      <p class="text-sm text-muted-foreground mb-5">
        Agregá tu primer producto para comenzar
      </p>
      <button
        v-if="store.isOwner"
        type="button"
        class="inline-flex items-center gap-2 h-10 px-5 bg-primary text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer"
        @click="openCreate"
      >
        <Plus class="w-4 h-4" />
        Nuevo Producto
      </button>
    </div>

    <!-- Empty filtered state -->
    <div
      v-else-if="filteredProducts.length === 0"
      class="rounded-xl border bg-card shadow-sm px-6 py-12 text-center"
    >
      <Search class="w-10 h-10 text-muted-foreground/20 mx-auto mb-3" />
      <p class="text-muted-foreground font-medium">
        No se encontraron productos con los filtros aplicados
      </p>
    </div>

    <!-- Product card grid -->
    <div
      v-else
      class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3 sm:gap-4"
    >
      <div
        v-for="product in filteredProducts"
        :key="product.id"
        class="rounded-xl border bg-card shadow-sm overflow-hidden flex flex-col group hover:shadow-md transition-shadow duration-200"
        :class="!product.active ? 'opacity-60' : ''"
      >
        <!-- Image / icon area -->
        <div class="relative aspect-square overflow-hidden">
          <img
            v-if="product.image_url"
            :src="product.image_url"
            :alt="product.name"
            class="w-full h-full object-cover"
          />
          <div
            v-else
            class="w-full h-full bg-gradient-to-br from-primary/5 via-primary/10 to-violet-100 flex items-center justify-center"
          >
            <Package class="w-10 h-10 sm:w-14 sm:h-14 text-primary/30" />
          </div>
          <!-- Fiscal badge — top-left like Amazon Prime -->
          <span
            v-if="product.is_fiscal"
            class="absolute top-2 left-2 bg-primary text-white text-[10px] font-bold px-1.5 py-0.5 rounded"
          >
            FISCAL
          </span>
          <!-- Inactive ribbon -->
          <span
            v-if="!product.active"
            class="absolute top-2 right-2 bg-muted text-muted-foreground text-[10px] font-semibold px-1.5 py-0.5 rounded"
          >
            INACTIVO
          </span>
        </div>

        <!-- Info area -->
        <div class="flex flex-col flex-1 p-2.5 sm:p-3 gap-1">
          <!-- Name -->
          <p
            class="text-xs sm:text-sm font-semibold text-foreground line-clamp-2 leading-snug"
          >
            {{ product.name }}
          </p>
          <!-- SKU -->
          <p
            v-if="product.sku"
            class="text-[10px] text-muted-foreground font-mono truncate"
          >
            {{ product.sku }}
          </p>
          <!-- Category -->
          <p class="text-[10px] text-muted-foreground truncate">
            {{ getCategoryName(product.category_id) }}
          </p>

          <!-- Price block — prominent like Amazon -->
          <div class="mt-auto pt-2">
            <p class="text-sm sm:text-base font-bold text-foreground font-mono">
              {{ formatPrice(product) }}
            </p>
            <p
              v-if="product.is_fiscal && product.tax_rate != null"
              class="text-[10px] text-muted-foreground"
            >
              + {{ product.tax_rate }}% IVA
            </p>
          </div>

          <!-- Actions row -->
          <div
            class="flex items-center justify-between pt-2 mt-1 border-t border-border"
          >
            <!-- Status pill -->
            <span
              :class="[
                'text-[10px] font-semibold px-1.5 py-0.5 rounded',
                product.active
                  ? 'bg-emerald-100 text-emerald-700'
                  : 'bg-muted text-muted-foreground',
              ]"
            >
              {{ product.active ? "Activo" : "Inactivo" }}
            </span>

            <!-- Buttons -->
            <div class="flex items-center gap-0.5">
              <button
                type="button"
                title="Editar"
                class="w-7 h-7 flex items-center justify-center rounded-lg text-muted-foreground hover:text-primary hover:bg-primary/10 transition-colors cursor-pointer"
                @click="openEdit(product)"
              >
                <Pencil class="w-3.5 h-3.5" />
              </button>

              <template v-if="store.isOwner">
                <template v-if="confirmDeleteId === product.id">
                  <button
                    type="button"
                    :disabled="deletingId === product.id"
                    class="h-7 px-2 bg-red-500 text-white text-[10px] font-bold rounded-md hover:bg-red-600 cursor-pointer disabled:opacity-50 transition-colors"
                    @click="deleteProduct(product.id)"
                  >
                    <Loader2
                      v-if="deletingId === product.id"
                      class="w-3 h-3 animate-spin"
                    />
                    <span v-else>Sí</span>
                  </button>
                  <button
                    type="button"
                    class="h-7 px-2 border text-[10px] font-bold text-muted-foreground rounded-md hover:bg-muted cursor-pointer transition-colors"
                    @click="confirmDeleteId = null"
                  >
                    No
                  </button>
                </template>
                <button
                  v-else
                  type="button"
                  title="Eliminar"
                  class="w-7 h-7 flex items-center justify-center rounded-lg text-muted-foreground hover:text-red-500 hover:bg-red-50 transition-colors cursor-pointer"
                  @click="confirmDeleteId = product.id"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Product form modal -->
    <ProductFormModal
      v-model="formModalOpen"
      :product="editingProduct"
      :categories="categories"
      @saved="onSaved"
    />
  </div>
</template>
