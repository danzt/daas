<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div :class="['bg-white rounded-xl shadow-xl p-6 w-full', $props.class]">
    <slot />
  </div>
</template>
