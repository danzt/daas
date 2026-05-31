<script setup lang="ts">
import { PanelLeft } from "lucide-vue-next";
import { useSidebar } from "./useSidebar";

const { toggleSidebar } = useSidebar();
</script>

<template>
  <button
    type="button"
    data-slot="sidebar-trigger"
    class="inline-flex h-7 w-7 items-center justify-center rounded-md text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
    @click="toggleSidebar"
  >
    <PanelLeft class="size-4" />
    <span class="sr-only">Toggle Sidebar</span>
  </button>
</template>
