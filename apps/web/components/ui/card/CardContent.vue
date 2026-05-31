<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div data-slot="card-content" :class="['px-4', $props.class]">
    <slot />
  </div>
</template>
