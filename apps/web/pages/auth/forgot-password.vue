<script setup lang="ts">
import { Loader2, MailCheck } from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";

definePageMeta({
  layout: "auth",
  middleware: "auth",
});

const store = useAuthStore();

const email = ref("");
const loading = ref(false);
const sent = ref(false);
const emailError = ref("");
const errorMessage = ref("");

function validateEmail(val: string): boolean {
  const re = /^[^\s@]+@[^\s@][^\s.@]*\.[^\s@]+$/;
  return re.test(val);
}

async function handleSubmit() {
  emailError.value = "";
  errorMessage.value = "";

  if (!email.value) {
    emailError.value = "El email es requerido";
    return;
  }
  if (!validateEmail(email.value)) {
    emailError.value = "Ingresá un email válido";
    return;
  }

  loading.value = true;
  try {
    await store.requestPasswordReset(email.value);
    // Always show success — the API never reveals whether the email exists.
    sent.value = true;
  } catch (err: unknown) {
    const apiError = err as { data?: { detail?: string }; message?: string };
    errorMessage.value =
      apiError?.data?.detail ??
      apiError?.message ??
      "No pudimos enviar el email. Intentá nuevamente.";
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

    <!-- Card -->
    <div class="rounded-xl border bg-card border-border shadow-md">
      <!-- Success state -->
      <div v-if="sent" class="px-6 py-10 text-center">
        <MailCheck class="mx-auto mb-4 size-12 text-primary" />
        <h2 class="text-xl font-bold font-heading text-foreground">
          Revisá tu email
        </h2>
        <p class="text-sm text-muted-foreground mt-2">
          Si <span class="font-semibold text-foreground">{{ email }}</span> está
          registrado, te enviamos un link para restablecer tu contraseña.
        </p>
        <NuxtLink
          to="/auth/login"
          class="inline-block mt-6 text-primary font-semibold hover:underline"
        >
          Volver a iniciar sesión
        </NuxtLink>
      </div>

      <!-- Form state -->
      <template v-else>
        <div class="px-6 pt-6 pb-2">
          <h2 class="text-2xl font-bold font-heading text-foreground">
            Recuperar contraseña
          </h2>
          <p class="text-sm text-muted-foreground mt-1">
            Te enviaremos un link a tu email para crear una nueva
          </p>
        </div>

        <form
          class="px-6 pb-6 pt-4 space-y-4"
          novalidate
          @submit.prevent="handleSubmit"
        >
          <div>
            <label
              for="email"
              class="block text-sm font-semibold text-foreground mb-1.5"
            >
              Email
            </label>
            <input
              id="email"
              v-model="email"
              type="email"
              placeholder="tu@empresa.com"
              autocomplete="email"
              :disabled="loading"
              :class="[
                'w-full h-11 px-4 border rounded-lg text-base transition-all duration-200',
                'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20',
                'placeholder:text-muted-foreground',
                emailError ? 'border-red-400' : 'border-input',
                loading && 'opacity-50 cursor-not-allowed',
              ]"
            />
            <p v-if="emailError" class="text-xs text-red-500 mt-1">
              {{ emailError }}
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
            <span>{{ loading ? "Enviando..." : "Enviar link" }}</span>
          </button>

          <p class="text-center text-sm text-muted-foreground">
            <NuxtLink
              to="/auth/login"
              class="text-primary font-semibold hover:underline cursor-pointer transition-all duration-200"
            >
              Volver a iniciar sesión
            </NuxtLink>
          </p>
        </form>
      </template>
    </div>
  </div>
</template>
