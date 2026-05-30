<script setup lang="ts">
interface Props {
  variant?: "default" | "cta" | "outline" | "ghost" | "destructive";
  size?: "sm" | "md" | "lg" | "icon";
  disabled?: boolean;
  type?: "button" | "submit" | "reset";
  class?: string;
}

const props = withDefaults(defineProps<Props>(), {
  variant: "default",
  size: "md",
  disabled: false,
  type: "button",
});

const variantClasses: Record<string, string> = {
  default: "bg-primary text-white hover:opacity-90",
  cta: "bg-cta text-white hover:opacity-90",
  outline:
    "border-2 border-primary text-primary bg-transparent hover:bg-primary hover:text-white",
  ghost: "bg-transparent text-primary hover:bg-primary/10",
  destructive: "bg-red-600 text-white hover:bg-red-700",
};

const sizeClasses: Record<string, string> = {
  sm: "h-9 px-3 text-sm",
  md: "h-11 px-6 text-base",
  lg: "h-12 px-8 text-lg",
  icon: "h-11 w-11 p-0",
};
</script>

<template>
  <button
    :type="props.type"
    :disabled="props.disabled"
    :class="[
      'inline-flex items-center justify-center gap-2 rounded-lg font-semibold transition-all duration-200 cursor-pointer focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2',
      variantClasses[props.variant],
      sizeClasses[props.size],
      props.disabled && 'opacity-50 cursor-not-allowed pointer-events-none',
      props.class,
    ]"
  >
    <slot />
  </button>
</template>
