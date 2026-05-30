<script setup lang="ts">
interface Props {
  type?: string;
  placeholder?: string;
  disabled?: boolean;
  modelValue?: string;
  class?: string;
  id?: string;
}

const props = withDefaults(defineProps<Props>(), {
  type: "text",
  disabled: false,
});

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();
</script>

<template>
  <input
    :id="props.id"
    :type="props.type"
    :placeholder="props.placeholder"
    :disabled="props.disabled"
    :value="props.modelValue"
    :class="[
      'w-full h-11 px-4 border border-gray-200 rounded-lg text-base transition-all duration-200',
      'focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20',
      'placeholder:text-gray-400',
      props.disabled && 'opacity-50 cursor-not-allowed bg-gray-50',
      props.class,
    ]"
    @input="
      emit('update:modelValue', ($event.target as HTMLInputElement).value)
    "
  />
</template>
