<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <h2 :class="['text-xl font-bold text-text-brand font-heading', $props.class]">
    <slot />
  </h2>
</template>
