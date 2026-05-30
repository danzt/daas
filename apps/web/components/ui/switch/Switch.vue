<script setup lang="ts">
interface Props {
  modelValue?: boolean;
  disabled?: boolean;
  class?: string;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  disabled: false,
});

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
}>();

function toggle() {
  if (!props.disabled) {
    emit("update:modelValue", !props.modelValue);
  }
}
</script>

<template>
  <button
    type="button"
    role="switch"
    :aria-checked="props.modelValue"
    :disabled="props.disabled"
    :class="[
      'relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200 cursor-pointer focus-visible:outline-2 focus-visible:outline-primary',
      props.modelValue ? 'bg-primary' : 'bg-gray-200',
      props.disabled && 'opacity-50 cursor-not-allowed',
      props.class,
    ]"
    @click="toggle"
  >
    <span
      :class="[
        'inline-block h-4 w-4 rounded-full bg-white shadow-sm transition-transform duration-200',
        props.modelValue ? 'translate-x-6' : 'translate-x-1',
      ]"
    />
  </button>
</template>
