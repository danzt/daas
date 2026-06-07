<script setup lang="ts">
import { ArrowRight } from "lucide-vue-next";

/**
 * MobileSectionHeader — a section label with an optional "see all" link.
 * Used between content blocks on the dashboard and detail screens.
 */
defineProps<{
  title: string;
  to?: string;
  actionLabel?: string;
}>();
</script>

<template>
  <div class="flex items-center justify-between px-4 pb-2 pt-1">
    <h2 class="text-base font-bold tracking-tight text-foreground">
      {{ title }}
    </h2>
    <NuxtLink
      v-if="to"
      :to="to"
      class="flex items-center gap-0.5 text-xs font-semibold text-primary"
    >
      {{ actionLabel ?? "Ver todo" }}
      <ArrowRight class="size-3.5" />
    </NuxtLink>
  </div>
</template>
