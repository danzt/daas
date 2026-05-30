<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div
    :class="[
      'bg-white rounded-xl border border-gray-100 shadow-md',
      $props.class,
    ]"
  >
    <slot />
  </div>
</template>
