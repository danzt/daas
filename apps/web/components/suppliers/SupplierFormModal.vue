<script setup lang="ts">
import { X, Save, Building2 } from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";

export interface Supplier {
  id: string;
  name: string;
  rif: string;
  contact_name: string;
  email: string;
  phone: string;
  address: string;
  notes: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

const props = defineProps<{
  open: boolean;
  supplier?: Supplier | null;
}>();

const emit = defineEmits<{
  close: [];
  saved: [supplier: Supplier];
}>();

const saving = ref(false);
const saveError = ref("");

const form = reactive({
  name: "",
  rif: "",
  contact_name: "",
  email: "",
  phone: "",
  address: "",
  notes: "",
  active: true,
});

watch(
  () => props.open,
  (open) => {
    if (!open) return;
    saveError.value = "";
    if (props.supplier) {
      form.name = props.supplier.name;
      form.rif = props.supplier.rif;
      form.contact_name = props.supplier.contact_name;
      form.email = props.supplier.email;
      form.phone = props.supplier.phone;
      form.address = props.supplier.address;
      form.notes = props.supplier.notes;
      form.active = props.supplier.active;
    } else {
      form.name = "";
      form.rif = "";
      form.contact_name = "";
      form.email = "";
      form.phone = "";
      form.address = "";
      form.notes = "";
      form.active = true;
    }
  },
);

const isEdit = computed(() => !!props.supplier);
const title = computed(() =>
  isEdit.value ? "Editar proveedor" : "Nuevo proveedor",
);

async function save() {
  if (!form.name.trim()) {
    saveError.value = "El nombre es obligatorio";
    return;
  }
  saving.value = true;
  saveError.value = "";
  try {
    let result: Supplier;
    if (isEdit.value && props.supplier) {
      result = await useApiFetch<Supplier>(
        `/api/v1/suppliers/${props.supplier.id}`,
        { method: "PUT", body: form },
      );
    } else {
      result = await useApiFetch<Supplier>("/api/v1/suppliers", {
        method: "POST",
        body: {
          name: form.name,
          rif: form.rif,
          contact_name: form.contact_name,
          email: form.email,
          phone: form.phone,
          address: form.address,
          notes: form.notes,
        },
      });
    }
    emit("saved", result);
  } catch (err: unknown) {
    const e = err as { data?: { detail?: string } };
    saveError.value =
      e?.data?.detail ?? "Ocurrió un error al guardar el proveedor";
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
        @click.self="emit('close')"
      >
        <div
          class="border bg-card rounded-xl shadow-2xl w-full max-w-lg flex flex-col max-h-[90vh]"
        >
          <!-- Header -->
          <div
            class="flex items-center justify-between px-6 py-4 border-b border-border"
          >
            <div class="flex items-center gap-3">
              <div class="p-2 bg-primary/10 rounded-lg">
                <Building2 class="w-5 h-5 text-primary" />
              </div>
              <h2 class="text-lg font-semibold text-foreground">{{ title }}</h2>
            </div>
            <button
              class="p-2 rounded-lg hover:bg-muted transition-colors text-muted-foreground"
              @click="emit('close')"
            >
              <X class="w-4 h-4" />
            </button>
          </div>

          <!-- Body -->
          <div class="flex-1 overflow-y-auto px-6 py-5 space-y-4">
            <!-- Error -->
            <div
              v-if="saveError"
              class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3"
            >
              {{ saveError }}
            </div>

            <!-- Name -->
            <div class="space-y-1.5">
              <label class="text-sm font-medium text-foreground"
                >Nombre <span class="text-destructive">*</span></label
              >
              <input
                v-model="form.name"
                class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="Distribuidora Ejemplo C.A."
              />
            </div>

            <!-- RIF -->
            <div class="space-y-1.5">
              <label class="text-sm font-medium text-foreground">RIF</label>
              <input
                v-model="form.rif"
                class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="J-12345678-9"
              />
            </div>

            <!-- Contact -->
            <div class="grid grid-cols-2 gap-3">
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground"
                  >Contacto</label
                >
                <input
                  v-model="form.contact_name"
                  class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="Juan Pérez"
                />
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground"
                  >Teléfono</label
                >
                <input
                  v-model="form.phone"
                  class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="+58 412 000 0000"
                />
              </div>
            </div>

            <!-- Email -->
            <div class="space-y-1.5">
              <label class="text-sm font-medium text-foreground">Email</label>
              <input
                v-model="form.email"
                type="email"
                class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="contacto@proveedor.com"
              />
            </div>

            <!-- Address -->
            <div class="space-y-1.5">
              <label class="text-sm font-medium text-foreground"
                >Dirección</label
              >
              <textarea
                v-model="form.address"
                rows="2"
                class="w-full px-3 py-2 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 resize-none"
                placeholder="Av. Principal, Local 5, Caracas"
              />
            </div>

            <!-- Notes -->
            <div class="space-y-1.5">
              <label class="text-sm font-medium text-foreground"
                >Notas internas</label
              >
              <textarea
                v-model="form.notes"
                rows="2"
                class="w-full px-3 py-2 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 resize-none"
                placeholder="Condiciones de pago, observaciones..."
              />
            </div>

            <!-- Active toggle (edit only) -->
            <div v-if="isEdit" class="flex items-center gap-3 pt-1">
              <button
                :class="[
                  'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
                  form.active ? 'bg-primary' : 'bg-muted-foreground/30',
                ]"
                @click="form.active = !form.active"
              >
                <span
                  :class="[
                    'inline-block h-4 w-4 rounded-full bg-white shadow transition-transform',
                    form.active ? 'translate-x-6' : 'translate-x-1',
                  ]"
                />
              </button>
              <span class="text-sm text-muted-foreground">{{
                form.active ? "Proveedor activo" : "Proveedor inactivo"
              }}</span>
            </div>
          </div>

          <!-- Footer -->
          <div
            class="flex items-center justify-end gap-3 px-6 py-4 border-t border-border"
          >
            <button
              class="px-4 py-2 text-sm rounded-lg border border-border hover:bg-muted transition-colors text-foreground"
              @click="emit('close')"
            >
              Cancelar
            </button>
            <button
              :disabled="saving"
              class="flex items-center gap-2 px-4 py-2 text-sm rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50"
              @click="save"
            >
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              <Save v-else class="w-4 h-4" />
              {{ saving ? "Guardando..." : "Guardar" }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
