<script setup lang="ts">
import { ChevronRight } from "lucide-vue-next";

/**
 * MobileListItem — a single tappable row in a native-style list.
 *
 * Slots:
 *   - #leading: avatar / icon bubble / status indicator (optional)
 *   - default: primary content (title, meta) — usually MobileListItem text helpers
 *   - #trailing: right-aligned value (price, count) (optional)
 *
 * Renders as a NuxtLink when `to` is set, otherwise a button (emits click),
 * otherwise a plain div. Full-row tappable, 56px+ min height.
 */
const props = withDefaults(
  defineProps<{
    to?: string;
    chevron?: boolean;
    accent?: "none" | "primary" | "success" | "warning" | "danger";
  }>(),
  { chevron: true, accent: "none" },
);

const emit = defineEmits<{ click: [] }>();

const tag = computed(() => {
  if (props.to) return resolveComponent("NuxtLink");
  return "div";
});

const accentBar = computed(
  () =>
    ({
      none: "",
      primary: "bg-primary",
      success: "bg-emerald-500",
      warning: "bg-orange-400",
      danger: "bg-red-500",
    })[props.accent],
);
</script>

<template>
  <component
    :is="tag"
    :to="to"
    class="flex items-stretch gap-0 bg-white transition-colors duration-100 active:bg-gray-50"
    @click="emit('click')"
  >
    <!-- Accent severity bar -->
    <div
      v-if="accent !== 'none'"
      class="my-2 w-1 shrink-0 rounded-r"
      :class="accentBar"
    />

    <div class="flex min-h-[60px] flex-1 items-center gap-3 px-4 py-3 min-w-0">
      <!-- Leading -->
      <div v-if="$slots.leading" class="shrink-0">
        <slot name="leading" />
      </div>

      <!-- Main -->
      <div class="min-w-0 flex-1">
        <slot />
      </div>

      <!-- Trailing -->
      <div
        v-if="$slots.trailing || chevron"
        class="flex shrink-0 items-center gap-1"
      >
        <slot name="trailing" />
        <ChevronRight v-if="chevron" class="size-4 text-muted-foreground/40" />
      </div>
    </div>
  </component>
</template>
