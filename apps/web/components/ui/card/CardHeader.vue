<script setup lang="ts">
interface Props {
  class?: string;
}
defineProps<Props>();
</script>

<template>
  <div
    data-slot="card-header"
    :class="[
      'grid auto-rows-min items-start gap-1 px-4 has-[+div]:pb-0',
      $props.class,
    ]"
  >
    <slot />
  </div>
</template>
