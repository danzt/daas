<script setup lang="ts">
import { Search, X } from "lucide-vue-next";

/**
 * MobileSearchBar — pill search input for list screens.
 * v-model binds the query. Clear button appears when non-empty.
 */
withDefaults(defineProps<{ placeholder?: string }>(), {
  placeholder: "Buscar…",
});

const model = defineModel<string>({ default: "" });
</script>

<template>
  <div class="px-4 py-2">
    <div class="relative">
      <Search
        class="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
      />
      <input
        v-model="model"
        type="search"
        :placeholder="placeholder"
        class="h-11 w-full rounded-full border border-border bg-muted/50 pl-10 pr-10 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20"
      />
      <button
        v-if="model"
        type="button"
        aria-label="Limpiar"
        class="no-min-tap absolute right-2 top-1/2 flex size-7 -translate-y-1/2 items-center justify-center rounded-full text-muted-foreground active:bg-accent"
        @click="model = ''"
      >
        <X class="size-4" />
      </button>
    </div>
  </div>
</template>
