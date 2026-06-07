<script setup lang="ts">
/**
 * MobileStatTile — a compact KPI tile for dashboard scrollers / grids.
 * Big value, label, optional icon bubble and trend chip.
 * Tappable when `to` is set.
 */
const props = withDefaults(
  defineProps<{
    label: string;
    value: string | number;
    sub?: string;
    icon?: Component;
    to?: string;
    tone?: "default" | "primary" | "success" | "warning" | "danger";
  }>(),
  { tone: "default" },
);

const tag = computed(() => (props.to ? resolveComponent("NuxtLink") : "div"));

const bubble = computed(
  () =>
    ({
      default: "bg-muted text-muted-foreground",
      primary: "bg-primary/10 text-primary",
      success: "bg-emerald-50 text-emerald-600",
      warning: "bg-orange-50 text-orange-500",
      danger: "bg-red-50 text-red-500",
    })[props.tone],
);

const valueColor = computed(
  () =>
    ({
      default: "text-foreground",
      primary: "text-primary",
      success: "text-emerald-600",
      warning: "text-orange-500",
      danger: "text-red-500",
    })[props.tone],
);
</script>

<template>
  <component
    :is="tag"
    :to="to"
    class="flex flex-col gap-2 rounded-2xl border border-border bg-white p-3.5 transition-transform duration-100 active:scale-[0.98]"
  >
    <div class="flex items-center justify-between">
      <div
        v-if="icon"
        class="flex size-8 items-center justify-center rounded-xl"
        :class="bubble"
      >
        <component :is="icon" class="size-4" />
      </div>
      <slot name="badge" />
    </div>
    <div>
      <p
        class="text-[22px] font-black leading-none tracking-tight tabular-nums"
        :class="valueColor"
      >
        {{ value }}
      </p>
      <p class="mt-1.5 text-xs font-medium text-muted-foreground leading-tight">
        {{ label }}
      </p>
      <p
        v-if="sub"
        class="mt-0.5 text-[11px] text-muted-foreground/70 leading-tight"
      >
        {{ sub }}
      </p>
    </div>
  </component>
</template>
