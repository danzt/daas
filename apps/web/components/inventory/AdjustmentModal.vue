<script setup lang="ts">
import {
  X,
  Loader2,
  AlertTriangle,
  TrendingUp,
  TrendingDown,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";

interface Props {
  open: boolean;
  productId: string;
  productName: string;
  currentStock: number;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  "update:open": [value: boolean];
  adjusted: [];
}>();

const delta = ref<number | null>(null);
const notes = ref("");
const saving = ref(false);
const deltaError = ref("");
const notesError = ref("");
const serverError = ref("");

const newStock = computed(() => {
  if (delta.value === null || isNaN(delta.value)) return props.currentStock;
  return props.currentStock + delta.value;
});

const isIncrease = computed(() => (delta.value ?? 0) > 0);
const isDecrease = computed(() => (delta.value ?? 0) < 0);
const wouldGoNegative = computed(() => newStock.value < 0);

function close() {
  emit("update:open", false);
  reset();
}

function reset() {
  delta.value = null;
  notes.value = "";
  deltaError.value = "";
  notesError.value = "";
  serverError.value = "";
}

async function submit() {
  deltaError.value = "";
  notesError.value = "";
  serverError.value = "";

  let valid = true;
  if (delta.value === null || delta.value === 0) {
    deltaError.value = "El delta no puede ser cero";
    valid = false;
  }
  if (!notes.value.trim()) {
    notesError.value = "La razón del ajuste es obligatoria";
    valid = false;
  }
  if (!valid) return;

  if (wouldGoNegative.value) {
    deltaError.value = `El stock quedaría en ${newStock.value.toFixed(2)}, lo cual es negativo`;
    return;
  }

  saving.value = true;
  try {
    await useApiFetch("/api/v1/inventory/adjustments", {
      method: "POST",
      body: {
        product_id: props.productId,
        delta: delta.value,
        notes: notes.value.trim(),
      },
    });
    emit("adjusted");
    close();
  } catch (err: unknown) {
    const e = err as { data?: { detail?: string; message?: string } };
    serverError.value =
      e?.data?.detail ?? e?.data?.message ?? "Error al registrar el ajuste";
  } finally {
    saving.value = false;
  }
}

watch(
  () => props.open,
  (v) => {
    if (!v) reset();
  },
);
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
        v-if="open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        role="dialog"
        aria-modal="true"
      >
        <div class="absolute inset-0 bg-black/50" @click="close" />
        <div
          class="relative z-10 w-full max-w-md bg-white rounded-xl shadow-xl"
        >
          <!-- Header -->
          <div
            class="flex items-center justify-between px-6 py-5 border-b border-gray-100"
          >
            <h2 class="text-xl font-bold font-heading text-text-brand">
              Ajuste de Stock
            </h2>
            <button
              type="button"
              class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-all duration-200 cursor-pointer"
              @click="close"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Body -->
          <div class="px-6 py-5 space-y-5">
            <!-- Product info -->
            <div class="rounded-lg bg-gray-50 px-4 py-3">
              <p
                class="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-0.5"
              >
                Producto
              </p>
              <p class="text-sm font-semibold text-text-brand">
                {{ productName }}
              </p>
              <p class="text-xs text-gray-500 mt-0.5">
                Stock actual:
                <span class="font-semibold text-text-brand">{{
                  currentStock
                }}</span>
              </p>
            </div>

            <!-- Server error -->
            <div
              v-if="serverError"
              class="rounded-lg bg-red-50 border border-red-200 px-4 py-3 flex items-start gap-2"
            >
              <AlertTriangle
                class="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5"
              />
              <p class="text-sm text-red-600">{{ serverError }}</p>
            </div>

            <!-- Delta input -->
            <div>
              <label
                for="adj-delta"
                class="block text-sm font-semibold text-text-brand mb-1.5"
              >
                Delta <span class="text-red-500">*</span>
                <span class="text-gray-400 font-normal ml-1">
                  (positivo = entrada, negativo = salida)
                </span>
              </label>
              <input
                id="adj-delta"
                v-model.number="delta"
                type="number"
                step="0.001"
                placeholder="Ej: 10 o -5"
                :class="[
                  'w-full h-11 px-4 border rounded-lg text-base transition-all duration-200',
                  'focus:outline-none focus:ring-2 placeholder:text-gray-400',
                  deltaError
                    ? 'border-red-400 focus:border-red-400 focus:ring-red-200'
                    : 'border-gray-200 focus:border-primary focus:ring-primary/20',
                ]"
              />
              <p v-if="deltaError" class="mt-1 text-xs text-red-500">
                {{ deltaError }}
              </p>
            </div>

            <!-- Impact preview -->
            <div
              v-if="delta !== null && delta !== 0 && !wouldGoNegative"
              class="rounded-lg border px-4 py-3 flex items-center gap-3"
              :class="
                isIncrease
                  ? 'border-green-200 bg-green-50'
                  : 'border-orange-200 bg-orange-50'
              "
            >
              <component
                :is="isIncrease ? TrendingUp : TrendingDown"
                class="w-5 h-5 flex-shrink-0"
                :class="isIncrease ? 'text-green-600' : 'text-orange-600'"
              />
              <div class="text-sm">
                <span
                  class="font-semibold"
                  :class="isIncrease ? 'text-green-700' : 'text-orange-700'"
                >
                  {{ currentStock }} → {{ newStock.toFixed(3) }}
                </span>
                <span class="text-gray-500 ml-1">
                  ({{ isIncrease ? "+" : "" }}{{ delta }})
                </span>
              </div>
            </div>

            <!-- Negative warning -->
            <div
              v-if="wouldGoNegative && delta !== null && delta !== 0"
              class="rounded-lg bg-red-50 border border-red-200 px-4 py-3 flex items-center gap-2"
            >
              <AlertTriangle class="w-4 h-4 text-red-500 flex-shrink-0" />
              <p class="text-sm text-red-600">
                El stock no puede ser negativo. Stock resultante:
                <strong>{{ newStock.toFixed(3) }}</strong>
              </p>
            </div>

            <!-- Notes -->
            <div>
              <label
                for="adj-notes"
                class="block text-sm font-semibold text-text-brand mb-1.5"
              >
                Razón del ajuste <span class="text-red-500">*</span>
              </label>
              <textarea
                id="adj-notes"
                v-model="notes"
                rows="3"
                placeholder="Ej: Conteo físico, merma, devolución de proveedor..."
                :class="[
                  'w-full px-4 py-3 border rounded-lg text-base transition-all duration-200 resize-none',
                  'focus:outline-none focus:ring-2 placeholder:text-gray-400',
                  notesError
                    ? 'border-red-400 focus:border-red-400 focus:ring-red-200'
                    : 'border-gray-200 focus:border-primary focus:ring-primary/20',
                ]"
              />
              <p v-if="notesError" class="mt-1 text-xs text-red-500">
                {{ notesError }}
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
              @click="close"
            >
              Cancelar
            </button>
            <button
              type="button"
              :disabled="saving || wouldGoNegative"
              class="flex items-center gap-2 h-10 px-6 bg-cta text-white text-sm font-semibold rounded-lg hover:opacity-90 transition-all duration-200 cursor-pointer disabled:opacity-50"
              @click="submit"
            >
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              <span>{{ saving ? "Guardando..." : "Registrar ajuste" }}</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
