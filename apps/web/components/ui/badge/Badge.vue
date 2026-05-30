<script setup lang="ts">
interface Props {
  variant?: "default" | "success" | "warning" | "danger" | "secondary";
  class?: string;
}

const props = withDefaults(defineProps<Props>(), {
  variant: "default",
});

const variantClasses: Record<string, string> = {
  default: "bg-primary/10 text-primary",
  success: "bg-green-100 text-green-700",
  warning: "bg-yellow-100 text-yellow-700",
  danger: "bg-red-100 text-red-700",
  secondary: "bg-gray-100 text-gray-600",
};
</script>

<template>
  <span
    :class="[
      'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold',
      variantClasses[props.variant],
      props.class,
    ]"
  >
    <slot />
  </span>
</template>
