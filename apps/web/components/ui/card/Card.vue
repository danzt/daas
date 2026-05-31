<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div
    data-slot="card"
    :class="[
      'flex flex-col gap-4 overflow-hidden rounded-xl bg-card py-4 text-sm text-card-foreground ring-1 ring-foreground/10',
      $props.class,
    ]"
  >
    <slot />
  </div>
</template>
