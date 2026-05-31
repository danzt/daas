<script setup lang="ts">
import {
  Eye,
  EyeOff,
  Loader2,
  ChevronLeft,
  ChevronRight,
  CheckCircle2,
} from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";

definePageMeta({
  layout: "auth",
  middleware: "auth",
});

const store = useAuthStore();
const router = useRouter();

const step = ref(1);
const loading = ref(false);
const errorMessage = ref("");

// Step 1 fields
const businessName = ref("");
const countryCode = ref("");
const ownerEmail = ref("");
const password = ref("");
const confirmPassword = ref("");
const showPassword = ref(false);
const showConfirmPassword = ref(false);

// Step 2 fields
const fiscalId = ref("");

// Errors
const errors = ref<Record<string, string>>({});

const countryOptions = [
  { value: "VE", label: "Venezuela" },
  { value: "DO", label: "Republica Dominicana" },
  { value: "US", label: "Estados Unidos" },
];

function validateEmail(val: string): boolean {
  const re = /^[^\s@]+@[^\s@][^\s.@]*\.[^\s@]+$/;
  return re.test(val);
}

function validateStep1(): boolean {
  const e: Record<string, string> = {};

  if (!businessName.value.trim())
    e.businessName = "El nombre de la empresa es requerido";
  if (!countryCode.value) e.countryCode = "Seleccioná un país";
  if (!ownerEmail.value) e.ownerEmail = "El email es requerido";
  else if (!validateEmail(ownerEmail.value))
    e.ownerEmail = "Ingresá un email válido";
  if (!password.value) e.password = "La contraseña es requerida";
  else if (password.value.length < 8) e.password = "Mínimo 8 caracteres";
  if (!confirmPassword.value) e.confirmPassword = "Confirmá la contraseña";
  else if (password.value !== confirmPassword.value)
    e.confirmPassword = "Las contraseñas no coinciden";

  errors.value = e;
  return Object.keys(e).length === 0;
}

function goNext() {
  if (step.value === 1 && validateStep1()) {
    step.value = 2;
  }
}

function goBack() {
  if (step.value > 1) {
    step.value--;
    errors.value = {};
  }
}

async function handleSubmit() {
  loading.value = true;
  errorMessage.value = "";

  try {
    await store.register({
      name: businessName.value,
      email: ownerEmail.value,
      password: password.value,
      country_code: countryCode.value,
      fiscal_id: fiscalId.value || undefined,
    });
    router.push("/dashboard");
  } catch (err: unknown) {
    const apiError = err as { data?: { detail?: string }; message?: string };
    errorMessage.value =
      apiError?.data?.detail ??
      apiError?.message ??
      "Error al crear la cuenta. Intentá nuevamente.";
  } finally {
    loading.value = false;
  }
}

const selectedCountryLabel = computed(
  () => countryOptions.find((c) => c.value === countryCode.value)?.label ?? "",
);
</script>

<template>
  <div class="w-full max-w-md">
    <!-- Brand -->
    <div class="text-center mb-8">
      <h1 class="text-4xl font-bold font-heading text-primary">DaaS</h1>
      <p class="text-sm text-muted-foreground mt-1 font-body">
        Registrá tu empresa
      </p>
    </div>

    <!-- Card -->
    <div class="rounded-xl border bg-card border-border shadow-md">
      <!-- Header with step indicator -->
      <div class="px-6 pt-6 pb-4 border-b border-border">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-xl font-bold font-heading text-foreground">
            {{ step === 1 ? "Datos de la empresa" : "Detalles de cuenta" }}
          </h2>
          <span
            class="text-xs font-semibold text-muted-foreground bg-gray-100 px-2.5 py-1 rounded-full"
          >
            {{ step }} de 2
          </span>
        </div>
        <!-- Step dots -->
        <div class="flex gap-2">
          <div
            v-for="n in 2"
            :key="n"
            :class="[
              'h-1.5 flex-1 rounded-full transition-colors duration-200',
              n <= step ? 'bg-primary' : 'bg-gray-200',
            ]"
          />
        </div>
      </div>

      <!-- Step 1 -->
      <form
        v-if="step === 1"
        class="px-6 py-5 space-y-4"
        novalidate
        @submit.prevent="goNext"
      >
        <!-- Business name -->
        <div>
          <label
            for="businessName"
            class="block text-sm font-semibold text-foreground mb-1.5"
          >
            Nombre de la empresa
          </label>
          <input
            id="businessName"
            v-model="businessName"
            type="text"
            placeholder="Mi Empresa S.A."
            autocomplete="organization"
            :class="[
              'w-full h-11 px-4 border rounded-lg text-base transition-all duration-200',
              'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-muted-foreground',
              errors.businessName ? 'border-red-400' : 'border-input',
            ]"
          />
          <p v-if="errors.businessName" class="text-xs text-red-500 mt-1">
            {{ errors.businessName }}
          </p>
        </div>

        <!-- Country -->
        <div>
          <label
            for="country"
            class="block text-sm font-semibold text-foreground mb-1.5"
          >
            País
          </label>
          <select
            id="country"
            v-model="countryCode"
            :class="[
              'w-full h-11 px-4 border rounded-lg text-base transition-all duration-200 bg-white cursor-pointer',
              'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20',
              errors.countryCode ? 'border-red-400' : 'border-input',
            ]"
          >
            <option value="" disabled>Seleccioná un país...</option>
            <option
              v-for="opt in countryOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </option>
          </select>
          <p v-if="errors.countryCode" class="text-xs text-red-500 mt-1">
            {{ errors.countryCode }}
          </p>
        </div>

        <!-- Owner email -->
        <div>
          <label
            for="ownerEmail"
            class="block text-sm font-semibold text-foreground mb-1.5"
          >
            Email del propietario
          </label>
          <input
            id="ownerEmail"
            v-model="ownerEmail"
            type="email"
            placeholder="propietario@empresa.com"
            autocomplete="email"
            :class="[
              'w-full h-11 px-4 border rounded-lg text-base transition-all duration-200',
              'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-muted-foreground',
              errors.ownerEmail ? 'border-red-400' : 'border-input',
            ]"
          />
          <p v-if="errors.ownerEmail" class="text-xs text-red-500 mt-1">
            {{ errors.ownerEmail }}
          </p>
        </div>

        <!-- Password -->
        <div>
          <label
            for="reg-password"
            class="block text-sm font-semibold text-foreground mb-1.5"
          >
            Contraseña
          </label>
          <div class="relative">
            <input
              id="reg-password"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="Mínimo 8 caracteres"
              autocomplete="new-password"
              :class="[
                'w-full h-11 px-4 pr-11 border rounded-lg text-base transition-all duration-200',
                'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-muted-foreground',
                errors.password ? 'border-red-400' : 'border-input',
              ]"
            />
            <button
              type="button"
              class="absolute inset-y-0 right-0 flex items-center px-3 text-muted-foreground hover:text-muted-foreground transition-colors duration-200 cursor-pointer"
              @click="showPassword = !showPassword"
            >
              <component :is="showPassword ? EyeOff : Eye" class="w-4 h-4" />
            </button>
          </div>
          <p v-if="errors.password" class="text-xs text-red-500 mt-1">
            {{ errors.password }}
          </p>
        </div>

        <!-- Confirm password -->
        <div>
          <label
            for="confirmPassword"
            class="block text-sm font-semibold text-foreground mb-1.5"
          >
            Confirmar contraseña
          </label>
          <div class="relative">
            <input
              id="confirmPassword"
              v-model="confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              placeholder="Repetí tu contraseña"
              autocomplete="new-password"
              :class="[
                'w-full h-11 px-4 pr-11 border rounded-lg text-base transition-all duration-200',
                'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-muted-foreground',
                errors.confirmPassword ? 'border-red-400' : 'border-input',
              ]"
            />
            <button
              type="button"
              class="absolute inset-y-0 right-0 flex items-center px-3 text-muted-foreground hover:text-muted-foreground transition-colors duration-200 cursor-pointer"
              @click="showConfirmPassword = !showConfirmPassword"
            >
              <component
                :is="showConfirmPassword ? EyeOff : Eye"
                class="w-4 h-4"
              />
            </button>
          </div>
          <p v-if="errors.confirmPassword" class="text-xs text-red-500 mt-1">
            {{ errors.confirmPassword }}
          </p>
        </div>

        <button
          type="submit"
          class="w-full h-12 flex items-center justify-center gap-2 bg-primary text-white rounded-lg font-semibold text-base transition-all duration-200 cursor-pointer hover:opacity-90 focus-visible:outline-2 focus-visible:outline-primary"
        >
          <span>Siguiente</span>
          <ChevronRight class="w-5 h-5" />
        </button>
      </form>

      <!-- Step 2 -->
      <form
        v-if="step === 2"
        class="px-6 py-5 space-y-4"
        novalidate
        @submit.prevent="handleSubmit"
      >
        <!-- Summary card -->
        <div
          class="rounded-lg bg-background border border-primary/20 p-4 space-y-2"
        >
          <p
            class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-2"
          >
            Resumen
          </p>
          <div class="flex items-start gap-2">
            <CheckCircle2 class="w-4 h-4 text-primary mt-0.5 flex-shrink-0" />
            <div>
              <p class="text-sm font-semibold text-foreground">
                {{ businessName }}
              </p>
              <p class="text-xs text-muted-foreground">
                {{ selectedCountryLabel }}
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <CheckCircle2 class="w-4 h-4 text-primary flex-shrink-0" />
            <p class="text-sm text-foreground">
              {{ ownerEmail }}
            </p>
          </div>
        </div>

        <!-- Fiscal ID -->
        <div>
          <label
            for="fiscalId"
            class="block text-sm font-semibold text-foreground mb-1.5"
          >
            <span>ID Fiscal</span>
            <span class="text-muted-foreground font-normal ml-1">
              ({{ countryCode === "VE" ? "RIF — requerido" : "opcional" }})
            </span>
          </label>
          <input
            id="fiscalId"
            v-model="fiscalId"
            type="text"
            :placeholder="
              countryCode === 'VE' ? 'J-12345678-9' : 'Identificación fiscal'
            "
            :class="[
              'w-full h-11 px-4 border rounded-lg text-base transition-all duration-200',
              'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 placeholder:text-muted-foreground',
              'border-input',
            ]"
          />
          <p class="text-xs text-muted-foreground mt-1">
            Podrás modificarlo luego en Configuración.
          </p>
        </div>

        <!-- Error -->
        <div
          v-if="errorMessage"
          class="rounded-lg bg-red-50 border border-red-200 px-4 py-3"
        >
          <p class="text-sm text-red-600">
            {{ errorMessage }}
          </p>
        </div>

        <!-- Buttons -->
        <div class="flex gap-3">
          <button
            type="button"
            :disabled="loading"
            class="flex-1 h-12 flex items-center justify-center gap-2 border-2 border-primary text-primary rounded-lg font-semibold text-base transition-all duration-200 cursor-pointer hover:bg-primary/5 disabled:opacity-50 disabled:cursor-not-allowed"
            @click="goBack"
          >
            <ChevronLeft class="w-5 h-5" />
            <span>Atrás</span>
          </button>
          <button
            type="submit"
            :disabled="loading"
            class="flex-1 h-12 flex items-center justify-center gap-2 bg-primary text-white rounded-lg font-semibold text-base transition-all duration-200 cursor-pointer hover:opacity-90 focus-visible:outline-2 focus-visible:outline-primary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
            <span>{{ loading ? "Creando cuenta..." : "Crear cuenta" }}</span>
          </button>
        </div>
      </form>

      <!-- Login link -->
      <div class="px-6 pb-5 text-center">
        <p class="text-sm text-muted-foreground">
          Ya tenés cuenta?
          <NuxtLink
            to="/auth/login"
            class="text-primary font-semibold hover:underline cursor-pointer transition-all duration-200"
          >
            Iniciá sesión
          </NuxtLink>
        </p>
      </div>
    </div>
  </div>
</template>
