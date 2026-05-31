<script setup lang="ts">
import { Loader2, X, Plus, Trash2 } from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";
import type { Product } from "~/components/products/ProductFormModal.vue";

export interface InvoiceLine {
  id: string;
  invoice_id: string;
  product_id: string;
  description: string;
  quantity: number;
  unit_price: number;
  subtotal: number;
  sort_order: number;
}

export interface Invoice {
  id: string;
  tenant_id: string;
  correlative: string;
  customer_name: string;
  customer_id_type: string;
  customer_id_number: string;
  status: "draft" | "issued" | "cancelled";
  subtotal: number;
  total: number;
  notes: string;
  issued_at: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
  lines: InvoiceLine[];
}

interface LineForm {
  _key: number;
  product_id: string;
  description: string;
  quantity: string;
  unit_price: string;
}

interface Props {
  modelValue: boolean;
  products: Product[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
  saved: [];
}>();

const CUSTOMER_ID_TYPES = [
  { value: "anonymous", label: "Anónimo" },
  { value: "cedula", label: "Cédula" },
  { value: "rif", label: "RIF" },
  { value: "passport", label: "Pasaporte" },
];

const customerName = ref("");
const customerIdType = ref("anonymous");
const customerIdNumber = ref("");
const notes = ref("");
const lines = ref<LineForm[]>([]);
const errors = ref<Record<string, string>>({});
const saving = ref(false);
const serverError = ref("");
let lineKey = 0;

const internalProducts = computed(() =>
  props.products.filter((p) => !p.is_fiscal && p.active),
);

const productOptions = computed(() =>
  internalProducts.value.map((p) => ({
    value: p.id,
    label: p.name + (p.sku ? ` (${p.sku})` : ""),
  })),
);

const total = computed(() =>
  lines.value.reduce((sum, l) => {
    const qty = parseFloat(l.quantity) || 0;
    const price = parseFloat(l.unit_price) || 0;
    return sum + qty * price;
  }, 0),
);

function addLine() {
  lines.value.push({
    _key: lineKey++,
    product_id: "",
    description: "",
    quantity: "1",
    unit_price: "",
  });
}

function removeLine(key: number) {
  lines.value = lines.value.filter((l) => l._key !== key);
}

function onProductChange(line: LineForm) {
  const product = internalProducts.value.find((p) => p.id === line.product_id);
  if (product) {
    line.description = product.name;
    line.unit_price = product.internal_price?.toString() ?? "";
  }
}

function resetForm() {
  customerName.value = "";
  customerIdType.value = "anonymous";
  customerIdNumber.value = "";
  notes.value = "";
  lines.value = [];
  errors.value = {};
  serverError.value = "";
  lineKey = 0;
  addLine();
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) resetForm();
  },
);

function validate(): boolean {
  const errs: Record<string, string> = {};

  if (lines.value.length === 0) {
    errs.lines = "Agregá al menos una línea";
  }

  lines.value.forEach((l, i) => {
    if (!l.product_id) errs[`line_${i}_product`] = "Seleccioná un producto";
    const qty = parseFloat(l.quantity);
    if (isNaN(qty) || qty <= 0) errs[`line_${i}_qty`] = "Cantidad inválida";
    const price = parseFloat(l.unit_price);
    if (isNaN(price) || price <= 0) errs[`line_${i}_price`] = "Precio inválido";
  });

  errors.value = errs;
  return Object.keys(errs).length === 0;
}

async function handleSave() {
  if (!validate()) return;

  saving.value = true;
  serverError.value = "";

  try {
    await useApiFetch("/api/v1/invoices/internal", {
      method: "POST",
      body: {
        customer_name: customerName.value.trim() || "Consumidor Final",
        customer_id_type: customerIdType.value,
        customer_id_number: customerIdNumber.value.trim(),
        notes: notes.value.trim(),
        lines: lines.value.map((l) => ({
          product_id: l.product_id,
          description: l.description.trim(),
          quantity: parseFloat(l.quantity),
          unit_price: parseFloat(l.unit_price),
        })),
      },
    });

    emit("update:modelValue", false);
    emit("saved");
  } catch (err: unknown) {
    const apiError = err as { data?: { detail?: string; message?: string } };
    serverError.value =
      apiError?.data?.detail ??
      apiError?.data?.message ??
      "Error al crear la factura";
  } finally {
    saving.value = false;
  }
}

function handleClose() {
  emit("update:modelValue", false);
}
</script>

<template>
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
        v-if="modelValue"
        class="fixed inset-0 z-50 flex items-start justify-center p-4 pt-10 overflow-y-auto"
        role="dialog"
        aria-modal="true"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/50" @click="handleClose" />

        <!-- Modal -->
        <div
          class="relative z-10 w-full max-w-2xl bg-white rounded-xl shadow-xl mb-10"
        >
          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-5 border-b">
            <h2 class="text-xl font-bold font-heading text-foreground">
              Nueva Factura Interna
            </h2>
            <button
              type="button"
              class="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-all duration-200 cursor-pointer"
              @click="handleClose"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Body -->
          <div class="px-6 py-5 space-y-6">
            <!-- Server error -->
            <div
              v-if="serverError"
              class="rounded-lg bg-red-50 border border-red-200 px-4 py-3"
            >
              <p class="text-sm text-red-600">{{ serverError }}</p>
            </div>

            <!-- Customer section -->
            <div>
              <p class="text-sm font-bold text-foreground mb-3">Cliente</p>
              <div class="grid grid-cols-3 gap-4">
                <!-- ID type -->
                <div>
                  <label
                    class="block text-xs font-semibold text-muted-foreground mb-1.5"
                  >
                    Tipo ID
                  </label>
                  <Select
                    v-model="customerIdType"
                    :options="CUSTOMER_ID_TYPES"
                    placeholder="Tipo..."
                  />
                </div>
                <!-- ID number -->
                <div>
                  <label
                    class="block text-xs font-semibold text-muted-foreground mb-1.5"
                  >
                    Número
                  </label>
                  <input
                    v-model="customerIdNumber"
                    type="text"
                    placeholder="V-12345678"
                    :disabled="customerIdType === 'anonymous'"
                    class="w-full h-10 px-3 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:border-primary focus:ring-primary/20 placeholder:text-gray-400 disabled:bg-muted disabled:text-muted-foreground"
                  />
                </div>
                <!-- Name -->
                <div>
                  <label
                    class="block text-xs font-semibold text-muted-foreground mb-1.5"
                  >
                    Nombre
                  </label>
                  <input
                    v-model="customerName"
                    type="text"
                    placeholder="Consumidor Final"
                    class="w-full h-10 px-3 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:border-primary focus:ring-primary/20 placeholder:text-gray-400"
                  />
                </div>
              </div>
            </div>

            <!-- Lines section -->
            <div>
              <div class="flex items-center justify-between mb-3">
                <p class="text-sm font-bold text-foreground">Líneas</p>
                <button
                  type="button"
                  class="flex items-center gap-1.5 text-xs font-semibold text-primary hover:text-primary/80 transition-colors cursor-pointer"
                  @click="addLine"
                >
                  <Plus class="w-3.5 h-3.5" />
                  Agregar línea
                </button>
              </div>

              <p v-if="errors.lines" class="text-xs text-red-500 mb-2">
                {{ errors.lines }}
              </p>

              <div class="space-y-3">
                <div
                  v-for="(line, i) in lines"
                  :key="line._key"
                  class="grid grid-cols-[1fr_auto_auto_auto] gap-2 items-start"
                >
                  <!-- Product -->
                  <div>
                    <Select
                      v-model="line.product_id"
                      :options="productOptions"
                      placeholder="Producto..."
                      @update:model-value="onProductChange(line)"
                    />
                    <p
                      v-if="errors[`line_${i}_product`]"
                      class="mt-0.5 text-xs text-red-500"
                    >
                      {{ errors[`line_${i}_product`] }}
                    </p>
                  </div>

                  <!-- Qty -->
                  <div class="w-20">
                    <input
                      v-model="line.quantity"
                      type="number"
                      min="0.001"
                      step="0.001"
                      placeholder="Cant."
                      :class="[
                        'w-full h-10 px-3 border rounded-lg text-sm text-right focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-gray-400',
                        errors[`line_${i}_qty`]
                          ? 'border-red-400'
                          : 'focus:border-primary',
                      ]"
                    />
                  </div>

                  <!-- Price -->
                  <div class="w-28">
                    <input
                      v-model="line.unit_price"
                      type="number"
                      min="0.01"
                      step="0.01"
                      placeholder="Precio"
                      :class="[
                        'w-full h-10 px-3 border rounded-lg text-sm text-right focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-gray-400',
                        errors[`line_${i}_price`]
                          ? 'border-red-400'
                          : 'focus:border-primary',
                      ]"
                    />
                  </div>

                  <!-- Subtotal + remove -->
                  <div class="flex items-center gap-2 pt-1">
                    <span
                      class="w-24 text-sm font-semibold text-right text-foreground tabular-nums"
                    >
                      {{
                        (
                          (parseFloat(line.quantity) || 0) *
                          (parseFloat(line.unit_price) || 0)
                        ).toFixed(2)
                      }}
                    </span>
                    <button
                      type="button"
                      class="p-1.5 text-muted-foreground hover:text-red-500 hover:bg-red-50 rounded-lg transition-all duration-200 cursor-pointer"
                      @click="removeLine(line._key)"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- Total -->
              <div
                v-if="lines.length > 0"
                class="mt-4 pt-3 border-t flex justify-end"
              >
                <div class="text-right">
                  <p class="text-xs text-muted-foreground">Total</p>
                  <p
                    class="text-xl font-bold font-heading text-foreground tabular-nums"
                  >
                    {{ total.toFixed(2) }}
                  </p>
                </div>
              </div>
            </div>

            <!-- Notes -->
            <div>
              <label class="block text-sm font-bold text-foreground mb-1.5">
                Notas
                <span class="font-normal text-muted-foreground"
                  >(opcional)</span
                >
              </label>
              <textarea
                v-model="notes"
                rows="2"
                placeholder="Observaciones para esta factura..."
                class="w-full px-4 py-3 border rounded-lg text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-gray-400 resize-none"
              />
            </div>
          </div>

          <!-- Footer -->
          <div class="px-6 py-4 border-t flex items-center justify-end gap-3">
            <button
              type="button"
              class="h-10 px-5 border text-muted-foreground text-sm font-semibold rounded-lg hover:bg-muted transition-all duration-200 cursor-pointer"
              @click="handleClose"
            >
              Cancelar
            </button>
            <button
              type="button"
              :disabled="saving"
              class="flex items-center gap-2 h-10 px-6 bg-primary text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer disabled:opacity-50"
              @click="handleSave"
            >
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              <span>{{ saving ? "Creando..." : "Crear borrador" }}</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
