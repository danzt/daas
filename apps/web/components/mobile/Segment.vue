<script setup lang="ts">
/**
 * MobileSegment — horizontally scrollable filter chips (segmented control).
 * Used for status/type filters at the top of list screens, ML/Amazon style.
 * v-model holds the active option key.
 */
defineProps<{
  options: { key: string; label: string; count?: number }[];
}>();

const model = defineModel<string>({ required: true });
</script>

<template>
  <div class="no-scrollbar flex gap-2 overflow-x-auto px-4 py-2">
    <button
      v-for="opt in options"
      :key="opt.key"
      type="button"
      class="no-min-tap flex shrink-0 items-center gap-1.5 rounded-full px-4 py-2 text-sm font-semibold transition-colors duration-150"
      :class="
        model === opt.key
          ? 'bg-primary text-white'
          : 'bg-muted text-muted-foreground active:bg-accent'
      "
      @click="model = opt.key"
    >
      {{ opt.label }}
      <span
        v-if="opt.count != null"
        class="rounded-full px-1.5 text-xs font-bold tabular-nums"
        :class="
          model === opt.key
            ? 'bg-white/20 text-white'
            : 'bg-white text-muted-foreground'
        "
      >
        {{ opt.count }}
      </span>
    </button>
  </div>
</template>
