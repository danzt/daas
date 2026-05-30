<script setup lang="ts">
interface Props {
  for?: string;
  class?: string;
}

defineProps<Props>();
</script>

<template>
  <label
    :for="$props.for"
    :class="[
      'block text-sm font-semibold text-text-brand mb-1.5 cursor-default',
      $props.class,
    ]"
  >
    <slot />
  </label>
</template>
