<script setup lang="ts">
import {
  Eye,
  EyeOff,
  Loader2,
  CheckCircle2,
  AlertTriangle,
} from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";

definePageMeta({
  layout: "auth",
  middleware: "auth",
});

const store = useAuthStore();

const accessToken = ref("");
const linkError = ref("");
const password = ref("");
const confirm = ref("");
const showPassword = ref(false);
const loading = ref(false);
const done = ref(false);
const passwordError = ref("");
const errorMessage = ref("");

// Supabase redirects here with the recovery token in the URL hash fragment:
// #access_token=...&type=recovery&... — the fragment is never sent to the
// server, so this must run on the client after mount.
onMounted(() => {
  const hash = window.location.hash.replace(/^#/, "");
  const params = new URLSearchParams(hash);

  const err = params.get("error_description") || params.get("error");
  if (err) {
    linkError.value =
      "El link no es válido o ya expiró. Pedí uno nuevo desde “Recuperar contraseña”.";
    return;
  }

  const token = params.get("access_token");
  const type = params.get("type");
  if (!token || type !== "recovery") {
    linkError.value =
      "Link inválido. Abrí el enlace del email de recuperación tal cual llegó.";
    return;
  }
  accessToken.value = token;
});

async function handleSubmit() {
  passwordError.value = "";
  errorMessage.value = "";

  if (password.value.length < 8) {
    passwordError.value = "La contraseña debe tener al menos 8 caracteres";
    return;
  }
  if (password.value !== confirm.value) {
    passwordError.value = "Las contraseñas no coinciden";
    return;
  }

  loading.value = true;
  try {
    await store.updatePassword(accessToken.value, password.value);
    done.value = true;
  } catch (err: unknown) {
    const apiError = err as { data?: { detail?: string }; message?: string };
    errorMessage.value =
      apiError?.data?.detail ??
      apiError?.message ??
      "No pudimos actualizar la contraseña. El link puede haber expirado.";
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="w-full max-w-md">
    <!-- Brand -->
    <div class="text-center mb-8">
      <h1 class="text-4xl font-bold font-heading text-primary">DaaS</h1>
      <p class="text-sm text-muted-foreground mt-1 font-body">
        Gestión de inventario empresarial
      </p>
    </div>

    <div class="rounded-xl border bg-card border-border shadow-md">
      <!-- Success -->
      <div v-if="done" class="px-6 py-10 text-center">
        <CheckCircle2 class="mx-auto mb-4 size-12 text-primary" />
        <h2 class="text-xl font-bold font-heading text-foreground">
          Contraseña actualizada
        </h2>
        <p class="text-sm text-muted-foreground mt-2">
          Ya podés iniciar sesión con tu nueva contraseña.
        </p>
        <NuxtLink
          to="/auth/login"
          class="inline-block mt-6 h-11 px-6 leading-[2.75rem] bg-primary text-white rounded-lg font-semibold hover:opacity-90"
        >
          Ir a iniciar sesión
        </NuxtLink>
      </div>

      <!-- Invalid / expired link -->
      <div v-else-if="linkError" class="px-6 py-10 text-center">
        <AlertTriangle class="mx-auto mb-4 size-12 text-red-500" />
        <h2 class="text-xl font-bold font-heading text-foreground">
          Link no válido
        </h2>
        <p class="text-sm text-muted-foreground mt-2">{{ linkError }}</p>
        <NuxtLink
          to="/auth/forgot-password"
          class="inline-block mt-6 text-primary font-semibold hover:underline"
        >
          Pedir un nuevo link
        </NuxtLink>
      </div>

      <!-- Form -->
      <template v-else>
        <div class="px-6 pt-6 pb-2">
          <h2 class="text-2xl font-bold font-heading text-foreground">
            Nueva contraseña
          </h2>
          <p class="text-sm text-muted-foreground mt-1">
            Elegí una contraseña segura para tu cuenta
          </p>
        </div>

        <form
          class="px-6 pb-6 pt-4 space-y-4"
          novalidate
          @submit.prevent="handleSubmit"
        >
          <div>
            <label
              for="password"
              class="block text-sm font-semibold text-foreground mb-1.5"
            >
              Nueva contraseña
            </label>
            <div class="relative">
              <input
                id="password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="Mínimo 8 caracteres"
                autocomplete="new-password"
                :disabled="loading"
                :class="[
                  'w-full h-11 px-4 pr-11 border rounded-lg text-base transition-all duration-200',
                  'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20',
                  'placeholder:text-muted-foreground',
                  passwordError ? 'border-red-400' : 'border-input',
                  loading && 'opacity-50 cursor-not-allowed',
                ]"
              />
              <button
                type="button"
                class="absolute inset-y-0 right-0 flex items-center px-3 text-muted-foreground transition-colors duration-200 cursor-pointer"
                :aria-label="
                  showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'
                "
                @click="showPassword = !showPassword"
              >
                <component :is="showPassword ? EyeOff : Eye" class="w-4 h-4" />
              </button>
            </div>
          </div>

          <div>
            <label
              for="confirm"
              class="block text-sm font-semibold text-foreground mb-1.5"
            >
              Confirmar contraseña
            </label>
            <input
              id="confirm"
              v-model="confirm"
              :type="showPassword ? 'text' : 'password'"
              placeholder="Repetí la contraseña"
              autocomplete="new-password"
              :disabled="loading"
              :class="[
                'w-full h-11 px-4 border rounded-lg text-base transition-all duration-200',
                'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20',
                'placeholder:text-muted-foreground',
                passwordError ? 'border-red-400' : 'border-input',
                loading && 'opacity-50 cursor-not-allowed',
              ]"
            />
            <p v-if="passwordError" class="text-xs text-red-500 mt-1">
              {{ passwordError }}
            </p>
          </div>

          <div
            v-if="errorMessage"
            class="rounded-lg bg-red-50 border border-red-200 px-4 py-3"
          >
            <p class="text-sm text-red-600">{{ errorMessage }}</p>
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full h-12 flex items-center justify-center gap-2 bg-primary text-white rounded-lg font-semibold text-base transition-all duration-200 cursor-pointer hover:opacity-90 focus-visible:outline-2 focus-visible:outline-primary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
            <span>{{ loading ? "Guardando..." : "Guardar contraseña" }}</span>
          </button>
        </form>
      </template>
    </div>
  </div>
</template>
