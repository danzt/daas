<script setup lang="ts">
import { provideSidebar } from "./useSidebar";

interface Props {
  defaultOpen?: boolean;
}
const props = withDefaults(defineProps<Props>(), { defaultOpen: true });

provideSidebar(props.defaultOpen);
</script>

<template>
  <div
    data-slot="sidebar-wrapper"
    class="group/sidebar-wrapper flex min-h-svh w-full bg-sidebar"
    style="--sidebar-width: 16rem; --sidebar-width-icon: 3rem"
  >
    <slot />
  </div>
</template>
