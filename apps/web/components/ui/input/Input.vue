<script setup lang="ts">
interface Props {
  type?: string;
  placeholder?: string;
  disabled?: boolean;
  modelValue?: string | number;
  min?: string | number;
  max?: string | number;
  step?: string | number;
  class?: string;
  id?: string;
}

const props = withDefaults(defineProps<Props>(), {
  type: "text",
  disabled: false,
});

const emit = defineEmits<{
  "update:modelValue": [value: string | number];
}>();

function handleInput(e: Event) {
  const target = e.target as HTMLInputElement;
  if (props.type === "number") {
    emit("update:modelValue", target.valueAsNumber);
  } else {
    emit("update:modelValue", target.value);
  }
}
</script>

<template>
  <input
    :id="props.id"
    :type="props.type"
    :placeholder="props.placeholder"
    :disabled="props.disabled"
    :value="props.modelValue"
    :min="props.min"
    :max="props.max"
    :step="props.step"
    :class="[
      'w-full h-9 px-3 border border-input bg-background rounded-md text-sm transition-all duration-200',
      'focus:border-ring focus:outline-none focus:ring-2 focus:ring-ring/20',
      'placeholder:text-muted-foreground',
      props.disabled && 'opacity-50 cursor-not-allowed bg-muted',
      props.class,
    ]"
    @input="handleInput"
  />
</template>
