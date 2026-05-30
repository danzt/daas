<script setup lang="ts">
interface Props {
  modelValue?: string;
  class?: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

provide(
  "tabs-value",
  computed(() => props.modelValue),
);
provide("tabs-set", (val: string) => emit("update:modelValue", val));
</script>

<template>
  <div :class="['w-full', props.class]">
    <slot />
  </div>
</template>
