<script setup lang="ts">
import { ChevronLeft } from "lucide-vue-next";

/**
 * MobileScreen — the scaffold every exclusive mobile screen sits inside.
 *
 * Owns: the sticky contextual top bar, safe-area insets, and the scroll
 * region that clears the bottom navigation. Pages just declare a title
 * (and optionally a back affordance + a right-side action) and drop their
 * content in the default slot.
 *
 * variant:
 *   - "light" (default): white bar, dark title — inner/list screens
 *   - "brand": purple bar, white title — home / hero screens
 * Use the #header slot to fully replace the bar (e.g. a custom greeting).
 */
const props = withDefaults(
  defineProps<{
    title?: string;
    subtitle?: string;
    back?: boolean;
    variant?: "light" | "brand";
  }>(),
  { back: false, variant: "light" },
);

const router = useRouter();

function goBack() {
  router.back();
}

const isBrand = computed(() => props.variant === "brand");
</script>

<template>
  <div class="flex h-full flex-col">
    <!-- ── Contextual top bar ─────────────────────────────── -->
    <slot name="header">
      <header
        class="sticky top-0 z-30 shrink-0"
        :class="isBrand ? 'bg-primary' : 'bg-white border-b border-border'"
        style="padding-top: env(safe-area-inset-top)"
      >
        <div class="flex h-14 items-center gap-2 px-3">
          <button
            v-if="back"
            type="button"
            aria-label="Volver"
            class="no-min-tap -ml-1 flex size-10 shrink-0 items-center justify-center rounded-full transition-colors"
            :class="
              isBrand
                ? 'text-white active:bg-white/15'
                : 'text-foreground active:bg-accent'
            "
            @click="goBack"
          >
            <ChevronLeft class="size-6" />
          </button>

          <div class="min-w-0 flex-1" :class="back ? '' : 'px-1'">
            <h1
              class="truncate text-lg font-bold leading-tight tracking-tight"
              :class="isBrand ? 'text-white' : 'text-foreground'"
            >
              {{ title }}
            </h1>
            <p
              v-if="subtitle"
              class="truncate text-xs leading-tight"
              :class="isBrand ? 'text-white/70' : 'text-muted-foreground'"
            >
              {{ subtitle }}
            </p>
          </div>

          <div class="flex shrink-0 items-center gap-1">
            <slot name="action" />
          </div>
        </div>
      </header>
    </slot>

    <!-- ── Scrollable content ─────────────────────────────── -->
    <div
      class="flex-1 overflow-y-auto"
      style="
        padding-bottom: calc(4.75rem + env(safe-area-inset-bottom));
        overscroll-behavior-y: contain;
      "
    >
      <slot />
    </div>
  </div>
</template>
