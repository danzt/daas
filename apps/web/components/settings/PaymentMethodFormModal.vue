<script setup lang="ts">
import {
  X,
  Save,
  Loader2,
  Smartphone,
  Building2,
  DollarSign,
  Bitcoin,
  Banknote,
  MoreHorizontal,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";

export interface PaymentMethod {
  id: string;
  tenant_id: string;
  type: PaymentMethodType;
  label: string;
  details: Record<string, string>;
  currency: string;
  active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

type PaymentMethodType =
  | "pago_movil"
  | "transfer_bank"
  | "zelle"
  | "paypal"
  | "usdt"
  | "cash"
  | "other";

const VENEZUELAN_BANKS = [
  "Banesco",
  "Banco de Venezuela",
  "Mercantil",
  "Provincial (BBVA)",
  "Banco del Tesoro",
  "BNC",
  "Bicentenario",
  "Venezolano de Crédito",
  "Fondo Común",
  "Exterior",
  "Activo",
  "Sofitasa",
  "Banplus",
  "Caroní",
  "Bancrecer",
  "Del Sur",
  "Otro",
];

const ACCOUNT_TYPES = [
  { value: "corriente", label: "Corriente" },
  { value: "ahorro", label: "Ahorro" },
];

const USDT_NETWORKS = [
  { value: "TRC20", label: "TRC20 (Tron)" },
  { value: "ERC20", label: "ERC20 (Ethereum)" },
  { value: "BEP20", label: "BEP20 (BSC)" },
];

interface MethodOption {
  type: PaymentMethodType;
  label: string;
  icon: Component;
  defaultCurrency: string;
}

const METHOD_OPTIONS: MethodOption[] = [
  {
    type: "pago_movil",
    label: "Pago Móvil",
    icon: Smartphone,
    defaultCurrency: "VES",
  },
  {
    type: "transfer_bank",
    label: "Transferencia bancaria",
    icon: Building2,
    defaultCurrency: "VES",
  },
  {
    type: "zelle",
    label: "Zelle",
    icon: DollarSign,
    defaultCurrency: "USD",
  },
  {
    type: "paypal",
    label: "PayPal",
    icon: DollarSign,
    defaultCurrency: "USD",
  },
  {
    type: "usdt",
    label: "USDT (Crypto)",
    icon: Bitcoin,
    defaultCurrency: "USDT",
  },
  {
    type: "cash",
    label: "Efectivo",
    icon: Banknote,
    defaultCurrency: "",
  },
  {
    type: "other",
    label: "Otro",
    icon: MoreHorizontal,
    defaultCurrency: "",
  },
];

const CURRENCY_OPTIONS = [
  { value: "VES", label: "VES (Bolívares)" },
  { value: "USD", label: "USD (Dólares)" },
  { value: "USDT", label: "USDT (Tether)" },
  { value: "EUR", label: "EUR (Euros)" },
  { value: "", label: "Sin especificar" },
];

// ─── Props & emits ────────────────────────────────────────────────────────────

const props = defineProps<{
  open: boolean;
  method?: PaymentMethod | null;
}>();

const emit = defineEmits<{
  close: [];
  saved: [method: PaymentMethod];
}>();

// ─── State ────────────────────────────────────────────────────────────────────

const saving = ref(false);
const saveError = ref("");
const selectedType = ref<PaymentMethodType>("pago_movil");

// Generic form state
const formLabel = ref("");
const formCurrency = ref("VES");
const formActive = ref(true);
const formSortOrder = ref(0);

// Type-specific detail fields
const d = reactive({
  // pago_movil
  pm_bank: "",
  pm_document: "",
  pm_phone: "",
  pm_holder: "",
  // transfer_bank
  tb_bank: "",
  tb_account_type: "corriente",
  tb_account_number: "",
  tb_document: "",
  tb_holder: "",
  // zelle
  z_email: "",
  z_holder: "",
  z_bank: "",
  // paypal
  pp_email: "",
  // usdt
  usdt_network: "TRC20",
  usdt_wallet: "",
  // cash
  cash_notes: "",
  // other
  other_notes: "",
});

const isEdit = computed(() => !!props.method);
const title = computed(() =>
  isEdit.value ? "Editar método de pago" : "Nuevo método de pago",
);

// ─── Watch open ──────────────────────────────────────────────────────────────

watch(
  () => props.open,
  (open) => {
    if (!open) return;
    saveError.value = "";
    saving.value = false;

    if (props.method) {
      selectedType.value = props.method.type;
      formLabel.value = props.method.label;
      formCurrency.value = props.method.currency || "";
      formActive.value = props.method.active;
      formSortOrder.value = props.method.sort_order;
      loadDetails(props.method.type, props.method.details);
    } else {
      selectedType.value = "pago_movil";
      formLabel.value = "";
      formCurrency.value = "VES";
      formActive.value = true;
      formSortOrder.value = 0;
      resetDetails();
    }
  },
);

// When user picks a new type (create mode), reset currency to default
watch(selectedType, (type) => {
  if (!isEdit.value) {
    const opt = METHOD_OPTIONS.find((m) => m.type === type);
    formCurrency.value = opt?.defaultCurrency ?? "";
  }
});

function loadDetails(type: PaymentMethodType, det: Record<string, string>) {
  resetDetails();
  if (type === "pago_movil") {
    d.pm_bank = det.bank ?? "";
    d.pm_document = det.document_number ?? "";
    d.pm_phone = det.phone ?? "";
    d.pm_holder = det.holder ?? "";
  } else if (type === "transfer_bank") {
    d.tb_bank = det.bank ?? "";
    d.tb_account_type = det.account_type ?? "corriente";
    d.tb_account_number = det.account_number ?? "";
    d.tb_document = det.document_number ?? "";
    d.tb_holder = det.account_holder ?? "";
  } else if (type === "zelle") {
    d.z_email = det.email ?? "";
    d.z_holder = det.account_holder ?? "";
    d.z_bank = det.bank ?? "";
  } else if (type === "paypal") {
    d.pp_email = det.email ?? "";
  } else if (type === "usdt") {
    d.usdt_network = det.network ?? "TRC20";
    d.usdt_wallet = det.wallet_address ?? "";
  } else if (type === "cash") {
    d.cash_notes = det.notes ?? "";
  } else if (type === "other") {
    d.other_notes = det.notes ?? "";
  }
}

function resetDetails() {
  d.pm_bank = "";
  d.pm_document = "";
  d.pm_phone = "";
  d.pm_holder = "";
  d.tb_bank = "";
  d.tb_account_type = "corriente";
  d.tb_account_number = "";
  d.tb_document = "";
  d.tb_holder = "";
  d.z_email = "";
  d.z_holder = "";
  d.z_bank = "";
  d.pp_email = "";
  d.usdt_network = "TRC20";
  d.usdt_wallet = "";
  d.cash_notes = "";
  d.other_notes = "";
}

function buildDetails(): Record<string, string> {
  const type = selectedType.value;
  if (type === "pago_movil") {
    return {
      bank: d.pm_bank,
      document_number: d.pm_document,
      phone: d.pm_phone,
      ...(d.pm_holder ? { holder: d.pm_holder } : {}),
    };
  } else if (type === "transfer_bank") {
    return {
      bank: d.tb_bank,
      account_type: d.tb_account_type,
      account_number: d.tb_account_number,
      document_number: d.tb_document,
      account_holder: d.tb_holder,
    };
  } else if (type === "zelle") {
    return {
      email: d.z_email,
      account_holder: d.z_holder,
      ...(d.z_bank ? { bank: d.z_bank } : {}),
    };
  } else if (type === "paypal") {
    return { email: d.pp_email };
  } else if (type === "usdt") {
    return { network: d.usdt_network, wallet_address: d.usdt_wallet };
  } else if (type === "cash") {
    return d.cash_notes ? { notes: d.cash_notes } : {};
  } else {
    return { notes: d.other_notes };
  }
}

// ─── Save ─────────────────────────────────────────────────────────────────────

async function save() {
  saving.value = true;
  saveError.value = "";
  try {
    const details = buildDetails();
    let result: PaymentMethod;
    if (isEdit.value && props.method) {
      result = await useApiFetch<PaymentMethod>(
        `/api/v1/payment-methods/${props.method.id}`,
        {
          method: "PUT",
          body: {
            label: formLabel.value || null,
            details,
            currency: formCurrency.value || null,
            active: formActive.value,
            sort_order: formSortOrder.value,
          },
        },
      );
    } else {
      result = await useApiFetch<PaymentMethod>("/api/v1/payment-methods", {
        method: "POST",
        body: {
          type: selectedType.value,
          label: formLabel.value,
          details,
          currency: formCurrency.value,
          sort_order: formSortOrder.value,
        },
      });
    }
    emit("saved", result);
  } catch (err: unknown) {
    const e = err as {
      data?: { detail?: string; message?: string; field?: string };
    };
    const field = e?.data?.field ? ` (campo: ${e.data.field})` : "";
    saveError.value =
      e?.data?.message ??
      e?.data?.detail ??
      `Ocurrió un error al guardar${field}`;
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      @click.self="emit('close')"
    >
      <div
        class="border bg-card rounded-xl shadow-2xl w-full max-w-xl flex flex-col max-h-[90vh]"
      >
        <!-- Header -->
        <div
          class="flex items-center justify-between px-6 py-4 border-b border-border shrink-0"
        >
          <h2 class="text-lg font-semibold text-foreground">{{ title }}</h2>
          <button
            class="p-2 rounded-lg hover:bg-muted transition-colors text-muted-foreground"
            @click="emit('close')"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Body -->
        <div class="flex-1 overflow-y-auto px-6 py-5 space-y-5">
          <!-- Error -->
          <div
            v-if="saveError"
            class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3"
          >
            {{ saveError }}
          </div>

          <!-- Type selector (create only) -->
          <div v-if="!isEdit" class="space-y-2">
            <p class="text-sm font-medium text-foreground">
              Tipo de método <span class="text-destructive">*</span>
            </p>
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-2">
              <button
                v-for="opt in METHOD_OPTIONS"
                :key="opt.type"
                type="button"
                :class="[
                  'flex items-center gap-2 px-3 py-2.5 rounded-lg border text-sm font-medium transition-all',
                  selectedType === opt.type
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-border text-foreground hover:bg-muted',
                ]"
                @click="selectedType = opt.type"
              >
                <component :is="opt.icon" class="w-4 h-4 shrink-0" />
                <span class="truncate">{{ opt.label }}</span>
              </button>
            </div>
          </div>

          <!-- Dynamic fields per type -->

          <!-- Pago Móvil -->
          <template v-if="selectedType === 'pago_movil'">
            <div class="space-y-3">
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground"
                  >Banco <span class="text-destructive">*</span></label
                >
                <select
                  v-model="d.pm_bank"
                  class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                >
                  <option value="">Seleccioná un banco</option>
                  <option v-for="b in VENEZUELAN_BANKS" :key="b" :value="b">
                    {{ b }}
                  </option>
                </select>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div class="space-y-1.5">
                  <label class="text-sm font-medium text-foreground"
                    >Cédula / RIF <span class="text-destructive">*</span></label
                  >
                  <input
                    v-model="d.pm_document"
                    class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                    placeholder="V-12345678"
                  />
                </div>
                <div class="space-y-1.5">
                  <label class="text-sm font-medium text-foreground"
                    >Teléfono <span class="text-destructive">*</span></label
                  >
                  <input
                    v-model="d.pm_phone"
                    class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                    placeholder="0412-1234567"
                  />
                </div>
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground"
                  >Titular (opcional)</label
                >
                <input
                  v-model="d.pm_holder"
                  class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="Nombre del titular"
                />
              </div>
            </div>
          </template>

          <!-- Transferencia bancaria -->
          <template v-else-if="selectedType === 'transfer_bank'">
            <div class="space-y-3">
              <div class="grid grid-cols-2 gap-3">
                <div class="space-y-1.5">
                  <label class="text-sm font-medium text-foreground"
                    >Banco <span class="text-destructive">*</span></label
                  >
                  <select
                    v-model="d.tb_bank"
                    class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    <option value="">Seleccioná un banco</option>
                    <option v-for="b in VENEZUELAN_BANKS" :key="b" :value="b">
                      {{ b }}
                    </option>
                  </select>
                </div>
                <div class="space-y-1.5">
                  <label class="text-sm font-medium text-foreground"
                    >Tipo de cuenta</label
                  >
                  <select
                    v-model="d.tb_account_type"
                    class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    <option
                      v-for="t in ACCOUNT_TYPES"
                      :key="t.value"
                      :value="t.value"
                    >
                      {{ t.label }}
                    </option>
                  </select>
                </div>
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground"
                  >Número de cuenta
                  <span class="text-destructive">*</span></label
                >
                <input
                  v-model="d.tb_account_number"
                  class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="01050000000000000000"
                  maxlength="20"
                />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div class="space-y-1.5">
                  <label class="text-sm font-medium text-foreground"
                    >Cédula / RIF</label
                  >
                  <input
                    v-model="d.tb_document"
                    class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                    placeholder="V-12345678"
                  />
                </div>
                <div class="space-y-1.5">
                  <label class="text-sm font-medium text-foreground"
                    >Titular <span class="text-destructive">*</span></label
                  >
                  <input
                    v-model="d.tb_holder"
                    class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                    placeholder="Empresa XYZ C.A."
                  />
                </div>
              </div>
            </div>
          </template>

          <!-- Zelle -->
          <template v-else-if="selectedType === 'zelle'">
            <div class="space-y-3">
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground"
                  >Email <span class="text-destructive">*</span></label
                >
                <input
                  v-model="d.z_email"
                  type="email"
                  class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="payments@example.com"
                />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div class="space-y-1.5">
                  <label class="text-sm font-medium text-foreground"
                    >Titular <span class="text-destructive">*</span></label
                  >
                  <input
                    v-model="d.z_holder"
                    class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                    placeholder="John Doe"
                  />
                </div>
                <div class="space-y-1.5">
                  <label class="text-sm font-medium text-foreground"
                    >Banco (opcional)</label
                  >
                  <input
                    v-model="d.z_bank"
                    class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                    placeholder="Bank of America"
                  />
                </div>
              </div>
            </div>
          </template>

          <!-- PayPal -->
          <template v-else-if="selectedType === 'paypal'">
            <div class="space-y-1.5">
              <label class="text-sm font-medium text-foreground"
                >Email de PayPal <span class="text-destructive">*</span></label
              >
              <input
                v-model="d.pp_email"
                type="email"
                class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="tu@paypal.com"
              />
            </div>
          </template>

          <!-- USDT -->
          <template v-else-if="selectedType === 'usdt'">
            <div class="space-y-3">
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground"
                  >Red <span class="text-destructive">*</span></label
                >
                <div class="flex gap-2">
                  <button
                    v-for="net in USDT_NETWORKS"
                    :key="net.value"
                    type="button"
                    :class="[
                      'flex-1 px-3 py-2 rounded-lg border text-sm font-medium transition-all',
                      d.usdt_network === net.value
                        ? 'border-primary bg-primary/10 text-primary'
                        : 'border-border text-foreground hover:bg-muted',
                    ]"
                    @click="d.usdt_network = net.value"
                  >
                    {{ net.label }}
                  </button>
                </div>
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground"
                  >Dirección de wallet
                  <span class="text-destructive">*</span></label
                >
                <input
                  v-model="d.usdt_wallet"
                  class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="TXYZ..."
                />
              </div>
            </div>
          </template>

          <!-- Efectivo -->
          <template v-else-if="selectedType === 'cash'">
            <div class="space-y-1.5">
              <label class="text-sm font-medium text-foreground"
                >Notas (opcional)</label
              >
              <textarea
                v-model="d.cash_notes"
                rows="3"
                class="w-full px-3 py-2 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 resize-none"
                placeholder="Pago en efectivo al recibir el pedido. Tener cambio exacto."
              />
            </div>
          </template>

          <!-- Otro -->
          <template v-else-if="selectedType === 'other'">
            <div class="space-y-3">
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground"
                  >Nombre del método
                  <span class="text-destructive">*</span></label
                >
                <input
                  v-model="formLabel"
                  class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="Ej: Pago por Instagram, Binance P2P..."
                />
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-foreground"
                  >Instrucciones <span class="text-destructive">*</span></label
                >
                <textarea
                  v-model="d.other_notes"
                  rows="3"
                  class="w-full px-3 py-2 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 resize-none"
                  placeholder="Describí cómo debe pagar el cliente..."
                />
              </div>
            </div>
          </template>

          <!-- Shared: label -->
          <div v-if="selectedType !== 'other'" class="space-y-1.5">
            <label class="text-sm font-medium text-foreground"
              >Nombre amigable (opcional)</label
            >
            <input
              v-model="formLabel"
              class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
              placeholder='Ej: "Banesco · Cuenta Corriente"'
            />
          </div>

          <!-- Currency -->
          <div class="space-y-1.5">
            <label class="text-sm font-medium text-foreground">Moneda</label>
            <select
              v-model="formCurrency"
              class="w-full h-10 px-3 rounded-lg border border-input bg-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
            >
              <option
                v-for="c in CURRENCY_OPTIONS"
                :key="c.value"
                :value="c.value"
              >
                {{ c.label }}
              </option>
            </select>
          </div>

          <!-- Active toggle -->
          <div class="flex items-center gap-3 pt-1">
            <button
              type="button"
              :class="[
                'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
                formActive ? 'bg-primary' : 'bg-muted-foreground/30',
              ]"
              @click="formActive = !formActive"
            >
              <span
                :class="[
                  'inline-block h-4 w-4 rounded-full bg-white shadow transition-transform',
                  formActive ? 'translate-x-6' : 'translate-x-1',
                ]"
              />
            </button>
            <span class="text-sm text-muted-foreground">{{
              formActive ? "Activo (visible en tienda)" : "Inactivo (oculto)"
            }}</span>
          </div>
        </div>

        <!-- Footer -->
        <div
          class="flex items-center justify-end gap-3 px-6 py-4 border-t border-border shrink-0"
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
  </Teleport>
</template>
