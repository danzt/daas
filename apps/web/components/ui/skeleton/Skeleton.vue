<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div :class="['animate-pulse rounded-md bg-gray-200', $props.class]" />
</template>
