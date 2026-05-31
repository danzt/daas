<script setup lang="ts">
interface Props {
  variant?:
    | "default"
    | "success"
    | "warning"
    | "danger"
    | "secondary"
    | "outline";
  class?: string;
}

const props = withDefaults(defineProps<Props>(), {
  variant: "default",
});

const variantClasses: Record<string, string> = {
  default: "bg-primary text-primary-foreground",
  secondary: "bg-secondary text-secondary-foreground",
  outline: "border bg-background text-foreground",
  success:
    "bg-emerald-50 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300",
  warning: "bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300",
  danger: "bg-rose-50 text-rose-700 dark:bg-rose-950 dark:text-rose-300",
};
</script>

<template>
  <span
    :class="[
      'inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium whitespace-nowrap',
      variantClasses[props.variant],
      props.class,
    ]"
  >
    <slot />
  </span>
</template>
