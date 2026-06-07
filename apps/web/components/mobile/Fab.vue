<script setup lang="ts">
import { Plus } from "lucide-vue-next";

/**
 * MobileFab — floating action button, the primary CTA on a screen.
 * Sits above the bottom nav, right-aligned. Orange CTA color.
 * Renders a NuxtLink when `to` is set, otherwise emits click.
 */
const props = defineProps<{
  to?: string;
  label?: string;
  icon?: Component;
}>();

const emit = defineEmits<{ click: [] }>();

const tag = computed(() =>
  props.to ? resolveComponent("NuxtLink") : "button",
);
</script>

<template>
  <component
    :is="tag"
    :to="to"
    type="button"
    class="no-min-tap fixed right-4 z-40 flex items-center gap-2 rounded-full bg-cta px-5 font-bold text-white shadow-lg shadow-cta/30 transition-transform duration-150 active:scale-95"
    :class="label ? 'h-14' : 'h-14 w-14 justify-center px-0'"
    style="bottom: calc(4.75rem + env(safe-area-inset-bottom))"
    @click="emit('click')"
  >
    <component :is="icon ?? Plus" class="size-6 shrink-0" />
    <span v-if="label" class="pr-1 text-sm">{{ label }}</span>
  </component>
</template>
