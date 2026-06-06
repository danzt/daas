<script setup lang="ts">
import {
  Palette,
  ImagePlus,
  Trash2,
  Loader2,
  Upload,
  Store,
  Eye,
} from "lucide-vue-next";
import { useApiFetch } from "~/composables/useAuth";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

// ─── Types ────────────────────────────────────────────────────────────────────

interface BrandingData {
  store_name: string;
  tagline: string;
  logo_url: string;
  banner_url: string;
  primary_color: string;
  notifications_whatsapp_phone: string;
}

// ─── State ────────────────────────────────────────────────────────────────────

const branding = ref<BrandingData>({
  store_name: "",
  tagline: "",
  logo_url: "",
  banner_url: "",
  primary_color: "",
  notifications_whatsapp_phone: "",
});

const form = ref({
  store_name: "",
  tagline: "",
  primary_color: "",
  notifications_whatsapp_phone: "",
});

const loading = ref(false);
const saving = ref(false);
const loadError = ref("");
const saveSuccess = ref(false);

const uploadingLogo = ref(false);
const uploadingBanner = ref(false);
const logoError = ref("");
const bannerError = ref("");

// ─── Preset colors ────────────────────────────────────────────────────────────

const colorPresets = [
  { label: "Purple", hex: "#7C3AED" },
  { label: "Blue", hex: "#2563EB" },
  { label: "Green", hex: "#059669" },
  { label: "Orange", hex: "#EA580C" },
  { label: "Red", hex: "#DC2626" },
];

// ─── Validation ───────────────────────────────────────────────────────────────

const hexColorRegex = /^#[0-9A-Fa-f]{6}$/;
const colorValid = computed(() => {
  const c = form.value.primary_color;
  return c === "" || hexColorRegex.test(c);
});

const previewColor = computed(() => {
  if (
    form.value.primary_color &&
    hexColorRegex.test(form.value.primary_color)
  ) {
    return form.value.primary_color;
  }
  return "#7C3AED";
});

const previewStoreName = computed(
  () => form.value.store_name || branding.value.store_name || "Mi Tienda",
);
const previewTagline = computed(
  () =>
    form.value.tagline || branding.value.tagline || "Tu tienda de confianza",
);

// ─── API ──────────────────────────────────────────────────────────────────────

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    const data = await useApiFetch<BrandingData>("/api/v1/branding");
    branding.value = data;
    form.value.store_name = data.store_name ?? "";
    form.value.tagline = data.tagline ?? "";
    form.value.primary_color = data.primary_color ?? "";
    form.value.notifications_whatsapp_phone =
      data.notifications_whatsapp_phone ?? "";
  } catch {
    loadError.value = "No se pudo cargar la configuración de branding";
  } finally {
    loading.value = false;
  }
}

async function save() {
  if (!colorValid.value) return;
  saving.value = true;
  saveSuccess.value = false;
  try {
    const updated = await useApiFetch<BrandingData>("/api/v1/branding", {
      method: "PUT",
      body: {
        store_name: form.value.store_name || null,
        tagline: form.value.tagline || null,
        primary_color: form.value.primary_color || null,
        notifications_whatsapp_phone:
          form.value.notifications_whatsapp_phone || null,
      },
    });
    branding.value = updated;
    saveSuccess.value = true;
    setTimeout(() => (saveSuccess.value = false), 2500);
  } catch {
    // keep form as-is
  } finally {
    saving.value = false;
  }
}

// ─── Image uploads ────────────────────────────────────────────────────────────

async function uploadLogo(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  logoError.value = "";
  if (file.size > 2 * 1024 * 1024) {
    logoError.value = "El logo no puede superar 2 MB";
    input.value = "";
    return;
  }
  if (!file.type.startsWith("image/")) {
    logoError.value = "Solo se admiten archivos de imagen";
    input.value = "";
    return;
  }

  uploadingLogo.value = true;
  try {
    const fd = new FormData();
    fd.append("logo", file);
    const updated = await useApiFetch<BrandingData>("/api/v1/branding/logo", {
      method: "POST",
      body: fd,
    });
    branding.value = updated;
  } catch {
    logoError.value = "No se pudo subir el logo";
  } finally {
    uploadingLogo.value = false;
    input.value = "";
  }
}

async function deleteLogo() {
  uploadingLogo.value = true;
  logoError.value = "";
  try {
    const updated = await useApiFetch<BrandingData>("/api/v1/branding/logo", {
      method: "DELETE",
    });
    branding.value = updated;
  } catch {
    logoError.value = "No se pudo eliminar el logo";
  } finally {
    uploadingLogo.value = false;
  }
}

async function uploadBanner(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  bannerError.value = "";
  if (file.size > 5 * 1024 * 1024) {
    bannerError.value = "El banner no puede superar 5 MB";
    input.value = "";
    return;
  }
  if (!file.type.startsWith("image/")) {
    bannerError.value = "Solo se admiten archivos de imagen";
    input.value = "";
    return;
  }

  uploadingBanner.value = true;
  try {
    const fd = new FormData();
    fd.append("banner", file);
    const updated = await useApiFetch<BrandingData>("/api/v1/branding/banner", {
      method: "POST",
      body: fd,
    });
    branding.value = updated;
  } catch {
    bannerError.value = "No se pudo subir el banner";
  } finally {
    uploadingBanner.value = false;
    input.value = "";
  }
}

async function deleteBanner() {
  uploadingBanner.value = true;
  bannerError.value = "";
  try {
    const updated = await useApiFetch<BrandingData>("/api/v1/branding/banner", {
      method: "DELETE",
    });
    branding.value = updated;
  } catch {
    bannerError.value = "No se pudo eliminar el banner";
  } finally {
    uploadingBanner.value = false;
  }
}

function selectPreset(hex: string) {
  form.value.primary_color = hex;
}

onMounted(load);
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6 max-w-5xl">
    <!-- Page header -->
    <div class="flex items-start justify-between gap-4 flex-wrap">
      <div>
        <h1 class="text-2xl font-bold text-foreground">Branding de tienda</h1>
        <p class="text-sm text-muted-foreground mt-0.5">
          Personalizá la apariencia de tu storefront público
        </p>
      </div>
    </div>

    <!-- Load error -->
    <div
      v-if="loadError"
      class="bg-destructive/10 border border-destructive/30 text-destructive text-sm rounded-lg px-4 py-3"
    >
      {{ loadError }}
    </div>

    <!-- Loading skeleton -->
    <div
      v-if="loading"
      class="flex items-center justify-center py-20 text-muted-foreground"
    >
      <Loader2 class="w-6 h-6 animate-spin mr-2" />
      Cargando branding...
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-[1fr_320px] gap-6">
      <!-- ── Left column: settings ───────────────────────────────────── -->
      <div class="space-y-5">
        <!-- 1. Identidad -->
        <section class="border bg-card rounded-xl p-5 space-y-4">
          <div class="flex items-center gap-2 mb-1">
            <Store class="w-4 h-4 text-primary" />
            <h2 class="font-semibold text-foreground">Identidad</h2>
          </div>

          <div class="space-y-1.5">
            <label class="block text-sm font-medium text-foreground">
              Nombre de la tienda
            </label>
            <input
              v-model="form.store_name"
              type="text"
              maxlength="120"
              placeholder="Mi Tienda Online"
              class="w-full h-10 px-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all"
            />
            <p class="text-xs text-muted-foreground">
              Si lo dejás vacío, se usará el nombre registrado de tu tenant.
            </p>
          </div>

          <div class="space-y-1.5">
            <label class="block text-sm font-medium text-foreground">
              Eslogan
              <span class="text-muted-foreground font-normal">(opcional)</span>
            </label>
            <input
              v-model="form.tagline"
              type="text"
              maxlength="200"
              placeholder="Tu tienda de confianza en Venezuela"
              class="w-full h-10 px-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all"
            />
          </div>
        </section>

        <!-- 2. Color primario -->
        <section class="border bg-card rounded-xl p-5 space-y-4">
          <div class="flex items-center gap-2 mb-1">
            <Palette class="w-4 h-4 text-primary" />
            <h2 class="font-semibold text-foreground">Color primario</h2>
          </div>

          <!-- Preset swatches -->
          <div class="flex items-center gap-2 flex-wrap">
            <button
              v-for="preset in colorPresets"
              :key="preset.hex"
              type="button"
              :title="preset.label"
              :style="{ backgroundColor: preset.hex }"
              :class="[
                'w-8 h-8 rounded-lg border-2 transition-all cursor-pointer',
                form.primary_color === preset.hex
                  ? 'border-foreground scale-110 shadow-md'
                  : 'border-transparent hover:scale-105',
              ]"
              @click="selectPreset(preset.hex)"
            />
          </div>

          <!-- Hex input + preview chip -->
          <div class="flex items-center gap-3">
            <div class="flex-1 relative">
              <input
                v-model="form.primary_color"
                type="text"
                maxlength="7"
                placeholder="#7C3AED"
                :class="[
                  'w-full h-10 px-3 rounded-lg border text-sm font-mono focus:outline-none focus:ring-2 transition-all',
                  colorValid
                    ? 'border-input bg-background focus:border-primary focus:ring-primary/20'
                    : 'border-destructive bg-destructive/5 focus:border-destructive focus:ring-destructive/20',
                ]"
              />
            </div>

            <!-- Live preview chip -->
            <div
              class="flex items-center gap-2 px-3 py-1.5 rounded-lg border text-sm font-medium shrink-0"
              :style="{
                backgroundColor: previewColor + '20',
                color: previewColor,
                borderColor: previewColor + '40',
              }"
            >
              <span
                class="w-3 h-3 rounded-full shrink-0"
                :style="{ backgroundColor: previewColor }"
              />
              Vista previa
            </div>
          </div>

          <p v-if="!colorValid" class="text-xs text-destructive">
            El color debe ser un hex válido de 6 dígitos, ej: #7C3AED
          </p>
          <p v-else class="text-xs text-muted-foreground">
            Si lo dejás vacío, se usará el color por defecto
            <span class="font-mono">#7C3AED</span>.
          </p>
        </section>

        <!-- 3. Logo -->
        <section class="border bg-card rounded-xl p-5 space-y-4">
          <div class="flex items-center gap-2 mb-1">
            <ImagePlus class="w-4 h-4 text-primary" />
            <h2 class="font-semibold text-foreground">Logo</h2>
            <span class="text-xs text-muted-foreground ml-auto"
              >Máx. 2 MB · imagen/*</span
            >
          </div>

          <!-- Current logo -->
          <div v-if="branding.logo_url" class="flex items-center gap-4">
            <img
              :src="branding.logo_url"
              alt="Logo actual"
              class="w-[120px] h-[120px] rounded-xl object-contain border bg-muted/30"
            />
            <div class="space-y-2">
              <p class="text-xs text-muted-foreground">Logo actual</p>
              <button
                type="button"
                :disabled="uploadingLogo"
                class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-destructive/40 hover:bg-destructive/10 text-xs text-destructive transition-colors disabled:opacity-50 cursor-pointer"
                @click="deleteLogo"
              >
                <Loader2
                  v-if="uploadingLogo"
                  class="w-3.5 h-3.5 animate-spin"
                />
                <Trash2 v-else class="w-3.5 h-3.5" />
                Eliminar logo
              </button>
            </div>
          </div>

          <!-- Upload zone -->
          <label
            :class="[
              'flex flex-col items-center justify-center gap-2 border-2 border-dashed rounded-xl p-6 cursor-pointer transition-colors',
              uploadingLogo
                ? 'opacity-60 pointer-events-none border-muted'
                : 'border-primary/30 hover:border-primary hover:bg-primary/5',
            ]"
          >
            <Loader2
              v-if="uploadingLogo"
              class="w-6 h-6 text-primary animate-spin"
            />
            <Upload v-else class="w-6 h-6 text-primary/60" />
            <span class="text-sm font-medium text-foreground">
              {{
                uploadingLogo
                  ? "Subiendo..."
                  : branding.logo_url
                    ? "Reemplazar logo"
                    : "Subir logo"
              }}
            </span>
            <span class="text-xs text-muted-foreground"
              >JPG, PNG, WEBP · Máx. 2 MB</span
            >
            <input
              type="file"
              accept="image/*"
              class="sr-only"
              :disabled="uploadingLogo"
              @change="uploadLogo"
            />
          </label>

          <p v-if="logoError" class="text-xs text-destructive">
            {{ logoError }}
          </p>
        </section>

        <!-- 4. Banner -->
        <section class="border bg-card rounded-xl p-5 space-y-4">
          <div class="flex items-center gap-2 mb-1">
            <ImagePlus class="w-4 h-4 text-primary" />
            <h2 class="font-semibold text-foreground">Banner héroe</h2>
            <span class="text-xs text-muted-foreground ml-auto"
              >Máx. 5 MB · imagen/*</span
            >
          </div>

          <!-- Current banner -->
          <div v-if="branding.banner_url" class="space-y-2">
            <img
              :src="branding.banner_url"
              alt="Banner actual"
              class="w-full rounded-xl object-cover max-h-48 border"
            />
            <div class="flex items-center justify-between">
              <p class="text-xs text-muted-foreground">Banner actual</p>
              <button
                type="button"
                :disabled="uploadingBanner"
                class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-destructive/40 hover:bg-destructive/10 text-xs text-destructive transition-colors disabled:opacity-50 cursor-pointer"
                @click="deleteBanner"
              >
                <Loader2
                  v-if="uploadingBanner"
                  class="w-3.5 h-3.5 animate-spin"
                />
                <Trash2 v-else class="w-3.5 h-3.5" />
                Eliminar banner
              </button>
            </div>
          </div>

          <!-- Upload zone -->
          <label
            :class="[
              'flex flex-col items-center justify-center gap-2 border-2 border-dashed rounded-xl p-6 cursor-pointer transition-colors',
              uploadingBanner
                ? 'opacity-60 pointer-events-none border-muted'
                : 'border-primary/30 hover:border-primary hover:bg-primary/5',
            ]"
          >
            <Loader2
              v-if="uploadingBanner"
              class="w-6 h-6 text-primary animate-spin"
            />
            <Upload v-else class="w-6 h-6 text-primary/60" />
            <span class="text-sm font-medium text-foreground">
              {{
                uploadingBanner
                  ? "Subiendo..."
                  : branding.banner_url
                    ? "Reemplazar banner"
                    : "Subir banner"
              }}
            </span>
            <span class="text-xs text-muted-foreground"
              >JPG, PNG, WEBP · Máx. 5 MB · Ancho recomendado: 1440px</span
            >
            <input
              type="file"
              accept="image/*"
              class="sr-only"
              :disabled="uploadingBanner"
              @change="uploadBanner"
            />
          </label>

          <p v-if="bannerError" class="text-xs text-destructive">
            {{ bannerError }}
          </p>
        </section>

        <!-- 5. Notificaciones WhatsApp -->
        <section class="border bg-card rounded-xl p-5 space-y-4">
          <div class="flex items-center gap-2 mb-1">
            <svg
              class="w-4 h-4 text-primary"
              viewBox="0 0 24 24"
              fill="currentColor"
              aria-hidden="true"
            >
              <path
                d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413Z"
              />
            </svg>
            <h2 class="font-semibold text-foreground">
              Notificaciones WhatsApp
            </h2>
          </div>

          <div class="space-y-1.5">
            <label class="block text-sm font-medium text-foreground">
              Número de WhatsApp para alertas
              <span class="text-muted-foreground font-normal">(opcional)</span>
            </label>
            <input
              v-model="form.notifications_whatsapp_phone"
              type="tel"
              maxlength="30"
              placeholder="+58412XXXXXXX"
              class="w-full h-10 px-3 rounded-lg border border-input bg-background text-sm font-mono focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all"
            />
            <p class="text-xs text-muted-foreground">
              Este número recibirá una alerta de WhatsApp cada vez que un
              cliente suba un comprobante de pago. Usá formato internacional
              E.164 (ej: <span class="font-mono">+584120000000</span>).
            </p>
          </div>
        </section>

        <!-- Save button -->
        <div class="flex items-center gap-3">
          <button
            type="button"
            :disabled="saving || !colorValid"
            class="flex items-center gap-2 px-5 py-2.5 bg-primary text-primary-foreground rounded-lg text-sm font-semibold hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            @click="save"
          >
            <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
            Guardar cambios
          </button>

          <Transition
            enter-active-class="transition duration-300 ease-out"
            enter-from-class="opacity-0 translate-x-2"
            enter-to-class="opacity-100 translate-x-0"
            leave-active-class="transition duration-200 ease-in"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
          >
            <span
              v-if="saveSuccess"
              class="text-sm text-emerald-600 font-medium"
            >
              Cambios guardados
            </span>
          </Transition>
        </div>
      </div>

      <!-- ── Right column: live preview ─────────────────────────────── -->
      <aside class="hidden lg:block">
        <div class="sticky top-20">
          <div class="border bg-card rounded-xl overflow-hidden shadow-sm">
            <!-- Preview header -->
            <div
              class="px-3 py-2 border-b bg-muted/40 flex items-center gap-1.5"
            >
              <Eye class="w-3.5 h-3.5 text-muted-foreground" />
              <span
                class="text-xs font-medium text-muted-foreground uppercase tracking-wider"
                >Vista previa</span
              >
            </div>

            <!-- Mini storefront header mockup -->
            <div class="p-4 space-y-3">
              <div
                class="rounded-xl p-4 flex items-center gap-3"
                :style="{
                  backgroundColor: previewColor + '12',
                  borderColor: previewColor + '30',
                }"
                style="border-width: 1px"
              >
                <!-- Logo or initials -->
                <div
                  class="w-12 h-12 rounded-xl flex items-center justify-center shrink-0 overflow-hidden"
                  :style="{ backgroundColor: previewColor + '20' }"
                >
                  <img
                    v-if="branding.logo_url"
                    :src="branding.logo_url"
                    alt="Logo"
                    class="w-full h-full object-contain"
                  />
                  <span
                    v-else
                    class="text-lg font-bold"
                    :style="{ color: previewColor }"
                  >
                    {{ (previewStoreName[0] ?? "T").toUpperCase() }}
                  </span>
                </div>

                <div class="min-w-0">
                  <p class="font-bold text-foreground text-sm truncate">
                    {{ previewStoreName }}
                  </p>
                  <p class="text-xs text-muted-foreground truncate">
                    {{ previewTagline }}
                  </p>
                </div>
              </div>

              <!-- Banner preview -->
              <div
                v-if="branding.banner_url"
                class="rounded-lg overflow-hidden"
              >
                <img
                  :src="branding.banner_url"
                  alt="Banner preview"
                  class="w-full h-20 object-cover"
                />
              </div>
              <div
                v-else
                class="rounded-lg h-16 flex items-center justify-center text-xs text-muted-foreground"
                :style="{
                  background: `linear-gradient(135deg, ${previewColor}15, ${previewColor}05)`,
                }"
              >
                Sin banner
              </div>

              <!-- Primary color swatch -->
              <div
                class="flex items-center gap-2 text-xs text-muted-foreground"
              >
                <span
                  class="w-5 h-5 rounded-md border"
                  :style="{ backgroundColor: previewColor }"
                />
                <span class="font-mono">{{ previewColor }}</span>
                <span class="ml-auto">Color principal</span>
              </div>

              <!-- CTA button preview -->
              <button
                type="button"
                class="w-full py-2 rounded-lg text-white text-xs font-bold pointer-events-none"
                :style="{ backgroundColor: previewColor }"
              >
                Ver catálogo
              </button>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>
