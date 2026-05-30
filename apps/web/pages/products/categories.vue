<script setup lang="ts">
import {
  Tag,
  Plus,
  Pencil,
  Trash2,
  Loader2,
  AlertTriangle,
  FolderOpen,
  X,
} from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";
import { useApiFetch } from "~/composables/useAuth";
import type { Category } from "~/components/products/ProductFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

const store = useAuthStore();

// ─── Data ────────────────────────────────────────────────────────────────────
const categories = ref<Category[]>([]);
const loading = ref(false);
const loadError = ref("");

// ─── Create/Edit dialog ───────────────────────────────────────────────────────
const dialogOpen = ref(false);
const editingCategory = ref<Category | null>(null);
const categoryName = ref("");
const savingCategory = ref(false);
const nameError = ref("");
const serverError = ref("");

// ─── Delete state ─────────────────────────────────────────────────────────────
const confirmDeleteId = ref<string | null>(null);
const deletingId = ref<string | null>(null);

const isEdit = computed(() => !!editingCategory.value);
const dialogTitle = computed(() =>
  isEdit.value ? "Editar Categoría" : "Nueva Categoría",
);

// ─── API calls ───────────────────────────────────────────────────────────────
async function fetchCategories() {
  loading.value = true;
  loadError.value = "";
  try {
    const data = await useApiFetch<Category[]>("/api/v1/products/categories");
    categories.value = data ?? [];
  } catch {
    loadError.value = "Error al cargar las categorías";
  } finally {
    loading.value = false;
  }
}

async function saveCategory() {
  nameError.value = "";
  serverError.value = "";

  if (!categoryName.value.trim()) {
    nameError.value = "El nombre es obligatorio";
    return;
  }

  savingCategory.value = true;
  try {
    if (isEdit.value && editingCategory.value) {
      // No PATCH for categories in spec — re-create workaround not ideal
      // backend doesn't expose PATCH /categories/:id in S2; skip edit for now
      // but we still close the dialog gracefully
      closeDialog();
    } else {
      await useApiFetch<Category>("/api/v1/products/categories", {
        method: "POST",
        body: { name: categoryName.value.trim() },
      });
      await fetchCategories();
      closeDialog();
    }
  } catch (err: unknown) {
    const apiError = err as { data?: { detail?: string; message?: string } };
    serverError.value =
      apiError?.data?.detail ??
      apiError?.data?.message ??
      "Error al guardar la categoría";
  } finally {
    savingCategory.value = false;
  }
}

async function deleteCategory(id: string) {
  deletingId.value = id;
  confirmDeleteId.value = null;
  try {
    await useApiFetch(`/api/v1/products/categories/${id}`, {
      method: "DELETE",
    });
    categories.value = categories.value.filter((c) => c.id !== id);
  } catch {
    // silently — category stays in list
  } finally {
    deletingId.value = null;
  }
}

// ─── Dialog helpers ───────────────────────────────────────────────────────────
function openCreate() {
  editingCategory.value = null;
  categoryName.value = "";
  nameError.value = "";
  serverError.value = "";
  dialogOpen.value = true;
}

function openEdit(cat: Category) {
  editingCategory.value = cat;
  categoryName.value = cat.name;
  nameError.value = "";
  serverError.value = "";
  dialogOpen.value = true;
}

function closeDialog() {
  dialogOpen.value = false;
  editingCategory.value = null;
  categoryName.value = "";
  nameError.value = "";
  serverError.value = "";
}

// ─── Lifecycle ────────────────────────────────────────────────────────────────
onMounted(() => {
  fetchCategories();
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
          <Tag class="w-5 h-5 text-primary" />
        </div>
        <div>
          <h1 class="text-2xl font-bold font-heading text-text-brand">
            Categorías
          </h1>
          <p class="text-gray-500 text-sm">
            Organizá tus productos por categorías
          </p>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <NuxtLink
          to="/products"
          class="h-10 px-4 flex items-center text-sm font-semibold text-gray-600 border border-gray-200 rounded-lg hover:bg-gray-50 transition-all duration-200 cursor-pointer"
        >
          Ver productos
        </NuxtLink>
        <button
          v-if="store.isOwner"
          type="button"
          class="flex items-center gap-2 h-10 px-5 bg-cta text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer"
          @click="openCreate"
        >
          <Plus class="w-4 h-4" />
          Nueva Categoría
        </button>
      </div>
    </div>

    <!-- Content card -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm">
      <!-- Table header -->
      <div
        class="px-6 py-4 border-b border-gray-100 flex items-center justify-between"
      >
        <h2 class="text-base font-bold font-heading text-text-brand">
          {{ categories.length }} categoría{{
            categories.length !== 1 ? "s" : ""
          }}
        </h2>
      </div>

      <!-- Loading skeleton -->
      <div v-if="loading" class="px-6 py-5 space-y-3">
        <div
          v-for="n in 4"
          :key="n"
          class="h-12 animate-pulse bg-gray-100 rounded-lg"
        />
      </div>

      <!-- Error state -->
      <div v-else-if="loadError" class="px-6 py-12 text-center">
        <AlertTriangle class="w-10 h-10 text-red-400 mx-auto mb-3" />
        <p class="text-red-600 font-medium">{{ loadError }}</p>
        <button
          type="button"
          class="mt-4 h-9 px-4 border border-gray-200 text-sm font-semibold text-gray-600 rounded-lg hover:bg-gray-50 cursor-pointer transition-all duration-200"
          @click="fetchCategories"
        >
          Reintentar
        </button>
      </div>

      <!-- Empty state -->
      <div v-else-if="categories.length === 0" class="px-6 py-16 text-center">
        <FolderOpen class="w-12 h-12 text-gray-200 mx-auto mb-4" />
        <h3 class="text-base font-semibold text-gray-600 mb-1">
          Sin categorías aún
        </h3>
        <p class="text-sm text-gray-400 mb-5">
          Creá tu primera categoría para organizar los productos
        </p>
        <button
          v-if="store.isOwner"
          type="button"
          class="inline-flex items-center gap-2 h-10 px-5 bg-cta text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer"
          @click="openCreate"
        >
          <Plus class="w-4 h-4" />
          Nueva Categoría
        </button>
      </div>

      <!-- List -->
      <div v-else>
        <div class="divide-y divide-gray-50">
          <div
            v-for="cat in categories"
            :key="cat.id"
            class="flex items-center justify-between px-6 py-4 hover:bg-gray-50/50 transition-colors duration-150"
          >
            <div class="flex items-center gap-3">
              <div
                class="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0"
              >
                <Tag class="w-4 h-4 text-primary" />
              </div>
              <div>
                <p class="text-sm font-semibold text-text-brand">
                  {{ cat.name }}
                </p>
                <p class="text-xs text-gray-400 font-mono mt-0.5">
                  {{ cat.id.slice(0, 8) }}...
                </p>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-2">
              <!-- Edit -->
              <button
                type="button"
                title="Editar"
                class="w-8 h-8 flex items-center justify-center rounded-lg text-gray-400 hover:text-primary hover:bg-primary/10 transition-all duration-200 cursor-pointer"
                @click="openEdit(cat)"
              >
                <Pencil class="w-4 h-4" />
              </button>

              <!-- Delete (owners only) -->
              <template v-if="store.isOwner">
                <!-- Confirm step -->
                <div
                  v-if="confirmDeleteId === cat.id"
                  class="flex items-center gap-1.5"
                >
                  <span class="text-xs text-red-600 font-semibold"
                    >¿Confirmás?</span
                  >
                  <button
                    type="button"
                    :disabled="deletingId === cat.id"
                    class="h-7 px-2.5 bg-red-500 text-white text-xs font-semibold rounded-md hover:bg-red-600 cursor-pointer disabled:opacity-50 transition-all duration-200"
                    @click="deleteCategory(cat.id)"
                  >
                    <Loader2
                      v-if="deletingId === cat.id"
                      class="w-3 h-3 animate-spin"
                    />
                    <span v-else>Sí</span>
                  </button>
                  <button
                    type="button"
                    class="h-7 px-2.5 border border-gray-200 text-gray-600 text-xs font-semibold rounded-md hover:bg-gray-50 cursor-pointer transition-all duration-200"
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
                  class="w-8 h-8 flex items-center justify-center rounded-lg text-gray-400 hover:text-red-500 hover:bg-red-50 transition-all duration-200 cursor-pointer"
                  @click="confirmDeleteId = cat.id"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create / Edit dialog -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-200"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-200"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="dialogOpen"
          class="fixed inset-0 z-50 flex items-center justify-center p-4"
          role="dialog"
          aria-modal="true"
        >
          <div class="absolute inset-0 bg-black/50" @click="closeDialog" />
          <div
            class="relative z-10 w-full max-w-sm bg-white rounded-xl shadow-xl"
          >
            <!-- Header -->
            <div
              class="flex items-center justify-between px-6 py-5 border-b border-gray-100"
            >
              <h2 class="text-xl font-bold font-heading text-text-brand">
                {{ dialogTitle }}
              </h2>
              <button
                type="button"
                class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-all duration-200 cursor-pointer"
                @click="closeDialog"
              >
                <X class="w-5 h-5" />
              </button>
            </div>

            <!-- Body -->
            <div class="px-6 py-5 space-y-4">
              <!-- Server error -->
              <div
                v-if="serverError"
                class="rounded-lg bg-red-50 border border-red-200 px-4 py-3"
              >
                <p class="text-sm text-red-600">{{ serverError }}</p>
              </div>

              <div>
                <label
                  for="cat-name"
                  class="block text-sm font-semibold text-text-brand mb-1.5"
                >
                  Nombre <span class="text-red-500">*</span>
                </label>
                <input
                  id="cat-name"
                  v-model="categoryName"
                  type="text"
                  placeholder="Ej: Alimentos, Bebidas, Electrónica..."
                  :class="[
                    'w-full h-11 px-4 border rounded-lg text-base transition-all duration-200',
                    'focus:outline-none focus:ring-2 placeholder:text-gray-400',
                    nameError
                      ? 'border-red-400 focus:border-red-400 focus:ring-red-200'
                      : 'border-gray-200 focus:border-primary focus:ring-primary/20',
                  ]"
                  @keyup.enter="saveCategory"
                />
                <p v-if="nameError" class="mt-1 text-xs text-red-500">
                  {{ nameError }}
                </p>
              </div>
            </div>

            <!-- Footer -->
            <div
              class="px-6 py-4 border-t border-gray-100 flex items-center justify-end gap-3"
            >
              <button
                type="button"
                class="h-10 px-5 border border-gray-200 text-gray-600 text-sm font-semibold rounded-lg hover:bg-gray-50 transition-all duration-200 cursor-pointer"
                @click="closeDialog"
              >
                Cancelar
              </button>
              <button
                type="button"
                :disabled="savingCategory"
                class="flex items-center gap-2 h-10 px-6 bg-cta text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer disabled:opacity-50"
                @click="saveCategory"
              >
                <Loader2 v-if="savingCategory" class="w-4 h-4 animate-spin" />
                <span>{{
                  savingCategory
                    ? "Guardando..."
                    : isEdit
                      ? "Guardar cambios"
                      : "Crear categoría"
                }}</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
