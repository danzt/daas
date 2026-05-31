<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div
    data-slot="card-description"
    :class="['text-sm text-muted-foreground', $props.class]"
  >
    <slot />
  </div>
</template>
