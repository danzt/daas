<script setup lang="ts">
import { useSidebar } from "./useSidebar";

interface Props {
  side?: "left" | "right";
  variant?: "sidebar" | "floating" | "inset";
  collapsible?: "offcanvas" | "icon" | "none";
}
const props = withDefaults(defineProps<Props>(), {
  side: "left",
  variant: "inset",
  collapsible: "icon",
});

const { state, isMobile, openMobile, setOpenMobile } = useSidebar();
</script>

<template>
  <!-- Mobile: slide-over drawer -->
  <template v-if="isMobile">
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-200"
        enter-from-class="opacity-0"
        leave-active-class="transition-opacity duration-200"
        leave-to-class="opacity-0"
      >
        <div
          v-if="openMobile"
          class="fixed inset-0 z-50 md:hidden"
          @click.self="setOpenMobile(false)"
        >
          <div class="absolute inset-0 bg-black/50" />
          <div
            class="absolute inset-y-0 left-0 w-72 bg-sidebar text-sidebar-foreground shadow-xl flex flex-col"
          >
            <slot />
          </div>
        </div>
      </Transition>
    </Teleport>
  </template>

  <!-- Desktop -->
  <div
    v-else
    class="group peer hidden text-sidebar-foreground md:block"
    :data-state="state"
    :data-collapsible="state === 'collapsed' ? props.collapsible : ''"
    :data-variant="props.variant"
    :data-side="props.side"
  >
    <!-- Gap that pushes the inset -->
    <div
      class="relative h-svh bg-transparent transition-[width] duration-200 ease-linear"
      :class="[
        props.variant === 'inset' || props.variant === 'floating'
          ? 'group-data-[collapsible=icon]:w-[calc(var(--sidebar-width-icon)+1rem)] w-[var(--sidebar-width)]'
          : 'group-data-[collapsible=icon]:w-[var(--sidebar-width-icon)] w-[var(--sidebar-width)]',
      ]"
    />

    <!-- Fixed sidebar -->
    <div
      class="fixed inset-y-0 left-0 z-10 hidden h-svh transition-[width] duration-200 ease-linear md:flex"
      :class="[
        props.variant === 'inset' || props.variant === 'floating'
          ? 'p-2 group-data-[collapsible=icon]:w-[calc(var(--sidebar-width-icon)+1rem)] w-[var(--sidebar-width)]'
          : 'group-data-[collapsible=icon]:w-[var(--sidebar-width-icon)] w-[var(--sidebar-width)] border-r border-sidebar-border',
      ]"
    >
      <div
        class="flex size-full flex-col bg-sidebar"
        :class="[
          props.variant === 'floating' &&
            'rounded-lg shadow-sm ring-1 ring-sidebar-border',
        ]"
      >
        <slot />
      </div>
    </div>
  </div>
</template>
