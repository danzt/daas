<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div :class="['flex border-b gap-1', $props.class]" role="tablist">
    <slot />
  </div>
</template>
