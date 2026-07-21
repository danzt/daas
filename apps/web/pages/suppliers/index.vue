<script setup lang="ts">
import {
  Building2,
  Plus,
  Search,
  Phone,
  Mail,
  Pencil,
  ChevronRight,
  Loader2,
  FolderOpen,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import SupplierFormModal from "~/components/suppliers/SupplierFormModal.vue";
import type { Supplier } from "~/components/suppliers/SupplierFormModal.vue";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

// ─── State ───────────────────────────────────────────────────────────────────
const { isMobile } = useMobileMode();

const suppliers = ref<Supplier[]>([]);
const loading = ref(false);
const loadError = ref("");
const searchQuery = ref("");
const showInactive = ref(false);
const formModalOpen = ref(false);
const editingSupplier = ref<Supplier | null>(null);

// ─── Computed ────────────────────────────────────────────────────────────────
const activeCount = computed(
  () => suppliers.value.filter((s) => s.active).length,
);

const filteredSuppliers = computed(() => {
  let list = showInactive.value
    ? suppliers.value
    : suppliers.value.filter((s) => s.active);

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        s.rif?.toLowerCase().includes(q) ||
        s.contact_name?.toLowerCase().includes(q),
    );
  }

  return list;
});

// ─── API ─────────────────────────────────────────────────────────────────────
async function loadSuppliers() {
  loading.value = true;
  loadError.value = "";
  try {
    const url = showInactive.value
      ? "/api/v1/suppliers?active=false"
      : "/api/v1/suppliers";
    suppliers.value = await useApiFetch<Supplier[]>(url);
  } catch {
    loadError.value = "No se pudieron cargar los proveedores";
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingSupplier.value = null;
  formModalOpen.value = true;
}

function openEdit(supplier: Supplier) {
  editingSupplier.value = supplier;
  formModalOpen.value = true;
}

function onSaved(supplier: Supplier) {
  const idx = suppliers.value.findIndex((s) => s.id === supplier.id);
  if (idx >= 0) {
    suppliers.value[idx] = supplier;
  } else {
    suppliers.value.unshift(supplier);
  }
  formModalOpen.value = false;
}

onMounted(loadSuppliers);
watch(showInactive, loadSuppliers);
</script>

<template>
  <!-- MOBILE -->
  <MobileScreensSuppliers
    v-if="isMobile"
    :suppliers="suppliers"
    :loading="loading"
    :load-error="loadError"
    :is-owner="true"
    @create="openCreate"
    @edit="openEdit"
    @retry="loadSuppliers"
  />

  <!-- WEB -->
  <div v-else class="p-4 sm:p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-foreground">Proveedores</h1>
        <p class="text-sm text-muted-foreground mt-0.5">
          {{ activeCount }} proveedor{{
            activeCount !== 1 ? "es" : ""
          }}
          activo{{ activeCount !== 1 ? "s" : "" }}
        </p>
      </div>
      <button
        class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors"
        @click="openCreate"
      >
        <Plus class="w-4 h-4" />
        Nuevo proveedor
      </button>
    </div>

    <!-- Filters -->
    <div class="flex items-center gap-3 flex-wrap">
      <div class="relative flex-1 min-w-48">
        <Search
          class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground"
        />
        <input
          v-model="searchQuery"
          class="w-full h-10 pl-9 pr-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
          placeholder="Buscar por nombre, RIF o contacto..."
        />
      </div>
      <label
        class="flex items-center gap-2 text-sm text-muted-foreground cursor-pointer select-none"
      >
        <input
          v-model="showInactive"
          type="checkbox"
          class="rounded border-input"
        />
        Mostrar inactivos
      </label>
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
      Cargando proveedores...
    </div>

    <!-- Empty -->
    <div
      v-else-if="!loading && filteredSuppliers.length === 0"
      class="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3"
    >
      <FolderOpen class="w-12 h-12 opacity-30" />
      <p class="text-sm">
        {{
          searchQuery
            ? "No hay proveedores que coincidan con la búsqueda"
            : "Aún no hay proveedores registrados"
        }}
      </p>
      <button
        v-if="!searchQuery"
        class="text-sm text-primary hover:underline"
        @click="openCreate"
      >
        Agregar el primero
      </button>
    </div>

    <!-- Table -->
    <div v-else-if="!loading" class="border bg-card rounded-xl overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border bg-muted/30">
              <th class="px-4 py-3 text-left font-medium text-muted-foreground">
                Proveedor
              </th>
              <th
                class="px-4 py-3 text-left font-medium text-muted-foreground hidden md:table-cell"
              >
                RIF
              </th>
              <th
                class="px-4 py-3 text-left font-medium text-muted-foreground hidden lg:table-cell"
              >
                Contacto
              </th>
              <th class="px-4 py-3 text-left font-medium text-muted-foreground">
                Estado
              </th>
              <th class="px-4 py-3" />
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            <tr
              v-for="sup in filteredSuppliers"
              :key="sup.id"
              class="hover:bg-muted/20 transition-colors group"
            >
              <td class="px-4 py-3">
                <div class="flex items-center gap-3">
                  <div
                    class="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center shrink-0"
                  >
                    <Building2 class="w-4 h-4 text-primary" />
                  </div>
                  <div>
                    <p class="font-medium text-foreground">{{ sup.name }}</p>
                    <p
                      v-if="sup.email"
                      class="text-xs text-muted-foreground flex items-center gap-1 mt-0.5"
                    >
                      <Mail class="w-3 h-3" />{{ sup.email }}
                    </p>
                  </div>
                </div>
              </td>
              <td
                class="px-4 py-3 text-muted-foreground hidden md:table-cell font-mono text-xs"
              >
                {{ sup.rif || "—" }}
              </td>
              <td class="px-4 py-3 hidden lg:table-cell">
                <div v-if="sup.contact_name || sup.phone">
                  <p class="text-foreground">{{ sup.contact_name || "—" }}</p>
                  <p
                    v-if="sup.phone"
                    class="text-xs text-muted-foreground flex items-center gap-1 mt-0.5"
                  >
                    <Phone class="w-3 h-3" />{{ sup.phone }}
                  </p>
                </div>
                <span v-else class="text-muted-foreground">—</span>
              </td>
              <td class="px-4 py-3">
                <span
                  :class="[
                    'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                    sup.active
                      ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300'
                      : 'bg-muted text-muted-foreground',
                  ]"
                >
                  {{ sup.active ? "Activo" : "Inactivo" }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div
                  class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity justify-end"
                >
                  <button
                    class="p-1.5 rounded-lg hover:bg-muted transition-colors text-muted-foreground"
                    title="Editar"
                    @click="openEdit(sup)"
                  >
                    <Pencil class="w-4 h-4" />
                  </button>
                  <NuxtLink
                    :to="`/suppliers/${sup.id}`"
                    class="p-1.5 rounded-lg hover:bg-muted transition-colors text-muted-foreground"
                    title="Ver detalle"
                  >
                    <ChevronRight class="w-4 h-4" />
                  </NuxtLink>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <SupplierFormModal
      :open="formModalOpen"
      :supplier="editingSupplier"
      @close="formModalOpen = false"
      @saved="onSaved"
    />
  </div>
</template>
