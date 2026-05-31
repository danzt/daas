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
  <div>
    <!-- Page header -->
    <div class="mb-8 flex items-center justify-between gap-4">
      <div class="flex items-center gap-3">
        <div
          class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center"
        >
          <Package class="w-5 h-5 text-primary" />
        </div>
        <div>
          <h1 class="text-2xl font-bold font-heading text-text-brand">
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
        class="flex items-center gap-2 h-10 px-5 bg-cta text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer"
        @click="openCreate"
      >
        <Plus class="w-4 h-4" />
        Nuevo Producto
      </button>
    </div>

    <!-- Stats row -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <div class="bg-white rounded-xl border shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Total
        </p>
        <p class="text-2xl font-bold font-heading text-text-brand">
          {{ totalProducts }}
        </p>
      </div>
      <div class="bg-white rounded-xl border shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Fiscales
        </p>
        <p class="text-2xl font-bold font-heading text-primary">
          {{ fiscalCount }}
        </p>
      </div>
      <div class="bg-white rounded-xl border shadow-sm px-5 py-4">
        <p
          class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1"
        >
          Internos
        </p>
        <p class="text-2xl font-bold font-heading text-muted-foreground">
          {{ internalCount }}
        </p>
      </div>
      <div class="bg-white rounded-xl border shadow-sm px-5 py-4">
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
    <div class="bg-white rounded-xl border shadow-sm px-5 py-4 mb-4">
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
                : 'text-muted-foreground hover:text-text-brand',
            ]"
            @click="typeFilter = tab.value"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>
    </div>

    <!-- Table card -->
    <div class="bg-white rounded-xl border shadow-sm">
      <!-- Loading skeleton -->
      <div v-if="loading" class="px-6 py-5 space-y-3">
        <div
          v-for="n in 3"
          :key="n"
          class="h-14 animate-pulse bg-muted rounded-lg"
        />
      </div>

      <!-- Error state -->
      <div v-else-if="loadError" class="px-6 py-12 text-center">
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

      <!-- Empty state (no data at all) -->
      <div v-else-if="products.length === 0" class="px-6 py-16 text-center">
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
          class="inline-flex items-center gap-2 h-10 px-5 bg-cta text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer"
          @click="openCreate"
        >
          <Plus class="w-4 h-4" />
          Nuevo Producto
        </button>
      </div>

      <!-- Empty filtered state -->
      <div
        v-else-if="filteredProducts.length === 0"
        class="px-6 py-12 text-center"
      >
        <Search class="w-10 h-10 text-muted-foreground/20 mx-auto mb-3" />
        <p class="text-muted-foreground font-medium">
          No se encontraron productos con los filtros aplicados
        </p>
      </div>

      <!-- Table -->
      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="border-b">
              <th
                class="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wide"
              >
                Producto
              </th>
              <th
                class="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wide"
              >
                Categoría
              </th>
              <th
                class="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wide"
              >
                Tipo
              </th>
              <th
                class="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wide"
              >
                Precio
              </th>
              <th
                class="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wide"
              >
                Estado
              </th>
              <th
                class="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wide"
              >
                Acciones
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            <tr
              v-for="product in filteredProducts"
              :key="product.id"
              class="hover:bg-muted/50 transition-colors duration-150"
            >
              <!-- Name + SKU -->
              <td class="px-6 py-4">
                <p class="text-sm font-semibold text-text-brand">
                  {{ product.name }}
                </p>
                <p
                  v-if="product.sku"
                  class="text-xs text-muted-foreground mt-0.5 font-mono"
                >
                  {{ product.sku }}
                </p>
              </td>

              <!-- Category -->
              <td class="px-6 py-4 text-sm text-muted-foreground">
                {{ getCategoryName(product.category_id) }}
              </td>

              <!-- Type badge -->
              <td class="px-6 py-4">
                <span
                  :class="[
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold',
                    product.is_fiscal
                      ? 'bg-primary/10 text-primary'
                      : 'bg-muted text-muted-foreground',
                  ]"
                >
                  {{ product.is_fiscal ? "Fiscal" : "Interno" }}
                </span>
              </td>

              <!-- Price -->
              <td class="px-6 py-4 text-sm font-mono text-text-brand">
                {{ formatPrice(product) }}
                <span
                  v-if="product.is_fiscal && product.tax_rate != null"
                  class="text-xs text-muted-foreground ml-1"
                  >+{{ product.tax_rate }}%</span
                >
              </td>

              <!-- Status badge -->
              <td class="px-6 py-4">
                <span
                  :class="[
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold',
                    product.active
                      ? 'bg-green-100 text-green-700'
                      : 'bg-red-100 text-red-600',
                  ]"
                >
                  {{ product.active ? "Activo" : "Inactivo" }}
                </span>
              </td>

              <!-- Actions -->
              <td class="px-6 py-4">
                <div class="flex items-center gap-2">
                  <!-- Edit -->
                  <button
                    type="button"
                    title="Editar"
                    class="w-8 h-8 flex items-center justify-center rounded-lg text-muted-foreground hover:text-primary hover:bg-primary/10 transition-all duration-200 cursor-pointer"
                    @click="openEdit(product)"
                  >
                    <Pencil class="w-4 h-4" />
                  </button>

                  <!-- Delete (owners only) -->
                  <template v-if="store.isOwner">
                    <!-- Confirm step -->
                    <div
                      v-if="confirmDeleteId === product.id"
                      class="flex items-center gap-1.5"
                    >
                      <span class="text-xs text-destructive font-semibold"
                        >¿Confirmás?</span
                      >
                      <button
                        type="button"
                        :disabled="deletingId === product.id"
                        class="h-7 px-2.5 bg-red-500 text-white text-xs font-semibold rounded-md hover:bg-red-600 cursor-pointer disabled:opacity-50 transition-all duration-200"
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
                        class="h-7 px-2.5 border text-muted-foreground text-xs font-semibold rounded-md hover:bg-muted cursor-pointer transition-all duration-200"
                        @click="confirmDeleteId = null"
                      >
                        No
                      </button>
                    </div>

                    <!-- Delete icon -->
                    <button
                      v-else
                      type="button"
                      title="Eliminar"
                      class="w-8 h-8 flex items-center justify-center rounded-lg text-muted-foreground hover:text-red-500 hover:bg-red-50 transition-all duration-200 cursor-pointer"
                      @click="confirmDeleteId = product.id"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
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
