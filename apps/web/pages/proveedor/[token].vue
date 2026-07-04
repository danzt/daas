<script setup lang="ts">
import {
  Loader2,
  Package,
  Plus,
  Trash2,
  Store,
  AlertTriangle,
} from "lucide-vue-next";

// Página pública — sin login, sin layout del panel.
definePageMeta({ layout: false });

interface CatalogItem {
  id: string;
  name: string;
  sku: string;
  cost: number;
  unit: string;
  barcode: string;
  description: string;
}
interface PortalInfo {
  supplier_name: string;
  store_name: string;
  catalog: CatalogItem[];
}

const route = useRoute();
const token = route.params.token as string;
const config = useRuntimeConfig();
const base = `${config.public.apiBase}/api/v1/supplier-portal/${token}`;

const info = ref<PortalInfo | null>(null);
const loading = ref(true);
const loadError = ref(false);
const saving = ref(false);
const formError = ref("");

const form = reactive({
  name: "",
  sku: "",
  cost: 0,
  unit: "",
  barcode: "",
  description: "",
});

async function load() {
  loading.value = true;
  loadError.value = false;
  try {
    info.value = await $fetch<PortalInfo>(base);
  } catch {
    loadError.value = true;
  } finally {
    loading.value = false;
  }
}

async function addItem() {
  if (!form.name.trim()) {
    formError.value = "El nombre del producto es obligatorio";
    return;
  }
  saving.value = true;
  formError.value = "";
  try {
    const item = await $fetch<CatalogItem>(`${base}/catalog`, {
      method: "POST",
      body: {
        name: form.name.trim(),
        sku: form.sku,
        cost: Number(form.cost) || 0,
        unit: form.unit,
        barcode: form.barcode,
        description: form.description,
      },
    });
    info.value?.catalog.unshift(item);
    form.name = "";
    form.sku = "";
    form.cost = 0;
    form.unit = "";
    form.barcode = "";
    form.description = "";
  } catch {
    formError.value = "No se pudo agregar el producto. Intentá de nuevo.";
  } finally {
    saving.value = false;
  }
}

async function removeItem(id: string) {
  try {
    await $fetch(`${base}/catalog/${id}`, { method: "DELETE" });
    if (info.value) {
      info.value.catalog = info.value.catalog.filter((i) => i.id !== id);
    }
  } catch {
    // no-op — el ítem sigue visible si falla
  }
}

function fmt(n: number) {
  return new Intl.NumberFormat("es-VE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
}

onMounted(load);
</script>

<template>
  <div class="min-h-screen bg-[#FAF5FF] py-8 px-4">
    <div class="mx-auto max-w-2xl">
      <!-- Loading -->
      <div v-if="loading" class="flex justify-center py-24">
        <Loader2 class="size-8 animate-spin text-primary" />
      </div>

      <!-- Link inválido -->
      <div
        v-else-if="loadError"
        class="rounded-2xl border border-border bg-white p-8 text-center shadow-sm"
      >
        <AlertTriangle class="mx-auto mb-3 size-12 text-red-500" />
        <h1 class="text-lg font-bold text-foreground">Link no válido</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Este enlace no existe o fue revocado. Pedile a tu contacto un link
          nuevo.
        </p>
      </div>

      <!-- Portal -->
      <template v-else-if="info">
        <!-- Header -->
        <div class="mb-6 flex items-center gap-3">
          <div
            class="flex size-12 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary"
          >
            <Store class="size-6" />
          </div>
          <div>
            <h1 class="text-xl font-bold text-foreground">
              Catálogo para {{ info.store_name }}
            </h1>
            <p class="text-sm text-muted-foreground">
              Hola {{ info.supplier_name }} — cargá acá los productos que le
              vendés.
            </p>
          </div>
        </div>

        <!-- Alta de producto -->
        <div class="rounded-2xl border border-border bg-white p-5 shadow-sm">
          <h2 class="mb-4 text-sm font-bold text-foreground">
            Agregar producto
          </h2>

          <div
            v-if="formError"
            class="mb-3 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600"
          >
            {{ formError }}
          </div>

          <div class="space-y-3">
            <div>
              <label
                class="mb-1 block text-xs font-semibold text-muted-foreground"
                >Nombre del producto *</label
              >
              <input
                v-model="form.name"
                placeholder="Ej. Labial mate rojo"
                class="h-11 w-full rounded-xl border border-border bg-white px-3 text-sm focus:border-primary focus:outline-none"
              />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label
                  class="mb-1 block text-xs font-semibold text-muted-foreground"
                  >Código / SKU</label
                >
                <input
                  v-model="form.sku"
                  placeholder="LAB-001"
                  class="h-11 w-full rounded-xl border border-border bg-white px-3 text-sm focus:border-primary focus:outline-none"
                />
              </div>
              <div>
                <label
                  class="mb-1 block text-xs font-semibold text-muted-foreground"
                  >Costo</label
                >
                <input
                  v-model="form.cost"
                  type="number"
                  inputmode="decimal"
                  min="0"
                  step="any"
                  placeholder="0.00"
                  class="h-11 w-full rounded-xl border border-border bg-white px-3 font-mono text-sm focus:border-primary focus:outline-none"
                />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label
                  class="mb-1 block text-xs font-semibold text-muted-foreground"
                  >Unidad</label
                >
                <input
                  v-model="form.unit"
                  placeholder="unidad / caja"
                  class="h-11 w-full rounded-xl border border-border bg-white px-3 text-sm focus:border-primary focus:outline-none"
                />
              </div>
              <div>
                <label
                  class="mb-1 block text-xs font-semibold text-muted-foreground"
                  >Código de barras</label
                >
                <input
                  v-model="form.barcode"
                  placeholder="759..."
                  class="h-11 w-full rounded-xl border border-border bg-white px-3 font-mono text-sm focus:border-primary focus:outline-none"
                />
              </div>
            </div>
            <button
              type="button"
              :disabled="saving"
              class="flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-primary text-sm font-bold text-white hover:opacity-90 disabled:opacity-50"
              @click="addItem"
            >
              <Loader2 v-if="saving" class="size-5 animate-spin" />
              <Plus v-else class="size-5" />
              Agregar al catálogo
            </button>
          </div>
        </div>

        <!-- Catálogo cargado -->
        <div class="mt-6">
          <h2 class="mb-3 text-sm font-bold text-foreground">
            Tu catálogo ({{ info.catalog.length }})
          </h2>
          <div
            v-if="info.catalog.length === 0"
            class="rounded-2xl border border-dashed border-border bg-white/50 py-10 text-center text-sm text-muted-foreground"
          >
            <Package class="mx-auto mb-2 size-8 opacity-30" />
            Todavía no cargaste productos.
          </div>
          <div
            v-else
            class="divide-y divide-border overflow-hidden rounded-2xl border border-border bg-white"
          >
            <div
              v-for="item in info.catalog"
              :key="item.id"
              class="flex items-center gap-3 px-4 py-3"
            >
              <div
                class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-muted text-muted-foreground"
              >
                <Package class="size-4" />
              </div>
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-semibold text-foreground">
                  {{ item.name }}
                </p>
                <p class="text-[11px] text-muted-foreground">
                  <span v-if="item.sku" class="font-mono">{{ item.sku }}</span>
                  <span v-if="item.unit"> · {{ item.unit }}</span>
                </p>
              </div>
              <span
                v-if="item.cost"
                class="font-mono text-sm font-bold text-foreground"
                >{{ fmt(item.cost) }}</span
              >
              <button
                type="button"
                aria-label="Eliminar"
                class="flex size-8 items-center justify-center rounded-lg text-muted-foreground hover:bg-red-50 hover:text-red-500"
                @click="removeItem(item.id)"
              >
                <Trash2 class="size-4" />
              </button>
            </div>
          </div>
        </div>

        <p class="mt-8 text-center text-xs text-muted-foreground">
          Portal de proveedores · DaaS
        </p>
      </template>
    </div>
  </div>
</template>
