<script setup lang="ts">
interface Props {
  isActive?: boolean;
  tooltip?: string;
  size?: "default" | "lg";
}
const props = withDefaults(defineProps<Props>(), {
  isActive: false,
  size: "default",
});
</script>

<template>
  <button
    type="button"
    data-slot="sidebar-menu-button"
    :data-active="props.isActive"
    :data-size="props.size"
    :title="props.tooltip"
    class="peer/menu-button flex w-full items-center gap-2 overflow-hidden rounded-md p-2 text-left text-sm outline-hidden transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:ring-2 active:bg-sidebar-accent active:text-sidebar-accent-foreground disabled:pointer-events-none disabled:opacity-50 data-[active=true]:bg-sidebar-accent data-[active=true]:font-medium data-[active=true]:text-sidebar-accent-foreground group-data-[collapsible=icon]:!size-8 group-data-[collapsible=icon]:!p-2 [&>span:last-child]:truncate [&>svg]:size-4 [&>svg]:shrink-0 data-[size=lg]:h-12 data-[size=lg]:text-sm group-data-[collapsible=icon]:[&>span:last-child]:hidden"
  >
    <slot />
  </button>
</template>
