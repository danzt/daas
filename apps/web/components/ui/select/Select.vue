<script setup lang="ts">
interface Option {
  value: string;
  label: string;
}

interface Props {
  modelValue?: string;
  options: Option[];
  placeholder?: string;
  disabled?: boolean;
  class?: string;
  id?: string;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  placeholder: "Seleccionar...",
});

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();
</script>

<template>
  <select
    :id="props.id"
    :value="props.modelValue"
    :disabled="props.disabled"
    :class="[
      'w-full h-11 px-4 border border-gray-200 rounded-lg text-base transition-all duration-200 bg-white',
      'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer',
      props.disabled && 'opacity-50 cursor-not-allowed bg-gray-50',
      props.class,
    ]"
    @change="
      emit('update:modelValue', ($event.target as HTMLSelectElement).value)
    "
  >
    <option value="" disabled>
      {{ props.placeholder }}
    </option>
    <option v-for="opt in props.options" :key="opt.value" :value="opt.value">
      {{ opt.label }}
    </option>
  </select>
</template>
