<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div :class="['px-6 pb-6 pt-0', $props.class]">
    <slot />
  </div>
</template>
