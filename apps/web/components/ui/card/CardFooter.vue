<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div
    data-slot="card-footer"
    :class="[
      'flex items-center rounded-b-xl border-t bg-muted/50 p-4',
      $props.class,
    ]"
  >
    <slot />
  </div>
</template>
