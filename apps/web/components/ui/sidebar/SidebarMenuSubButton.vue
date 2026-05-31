<script setup lang="ts">
interface Props {
  isActive?: boolean;
}
const props = withDefaults(defineProps<Props>(), { isActive: false });
</script>

<template>
  <button
    type="button"
    data-slot="sidebar-menu-sub-button"
    :data-active="props.isActive"
    class="flex h-7 min-w-0 items-center gap-2 overflow-hidden rounded-md px-2 text-sm text-sidebar-foreground outline-hidden transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:ring-2 active:bg-sidebar-accent disabled:pointer-events-none disabled:opacity-50 data-[active=true]:bg-sidebar-accent data-[active=true]:text-sidebar-accent-foreground [&>span:last-child]:truncate [&>svg]:size-4 [&>svg]:shrink-0"
  >
    <slot />
  </button>
</template>
