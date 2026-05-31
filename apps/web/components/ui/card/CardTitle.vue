<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div
    data-slot="card-title"
    :class="['text-base leading-snug font-medium', $props.class]"
  >
    <slot />
  </div>
</template>
