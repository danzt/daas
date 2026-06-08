<script setup lang="ts">
import {
  Tag,
  Pencil,
  Trash2,
  Loader2,
  AlertTriangle,
  FolderOpen,
} from "lucide-vue-next";
import type { Category } from "~/components/products/ProductFormModal.vue";

const props = defineProps<{
  categories: Category[];
  loading: boolean;
  loadError: string;
  isOwner: boolean;
  deletingId: string | null;
}>();

const emit = defineEmits<{
  create: [];
  edit: [cat: Category];
  remove: [id: string];
  retry: [];
}>();

const confirmId = ref<string | null>(null);
</script>

<template>
  <MobileScreen title="Categorías" back subtitle="Catálogo de productos">
    <div v-if="loadError" class="px-4 py-12 text-center">
      <AlertTriangle class="mx-auto mb-3 size-10 text-destructive" />
      <p class="font-medium text-destructive">{{ loadError }}</p>
      <button
        type="button"
        class="mt-4 h-10 rounded-xl border border-border px-5 text-sm font-semibold text-muted-foreground active:bg-accent"
        @click="emit('retry')"
      >
        Reintentar
      </button>
    </div>

    <div
      v-else-if="loading && !categories.length"
      class="flex justify-center py-16"
    >
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <div v-else-if="!categories.length" class="px-6 py-16 text-center">
      <FolderOpen class="mx-auto mb-4 size-12 text-muted-foreground/20" />
      <h3 class="mb-1 text-base font-semibold text-muted-foreground">
        Sin categorías
      </h3>
      <p class="text-sm text-muted-foreground">
        Creá tu primera categoría para organizar el catálogo
      </p>
    </div>

    <div v-else class="divide-y divide-border border-t border-border bg-white">
      <div
        v-for="cat in categories"
        :key="cat.id"
        class="flex items-center gap-3 px-4 py-3.5"
      >
        <div
          class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary"
        >
          <Tag class="size-5" />
        </div>
        <p
          class="min-w-0 flex-1 truncate text-sm font-semibold text-foreground"
        >
          {{ cat.name }}
        </p>

        <!-- Inline delete confirm -->
        <template v-if="isOwner">
          <template v-if="confirmId === cat.id">
            <button
              type="button"
              :disabled="deletingId === cat.id"
              class="no-min-tap h-9 rounded-xl bg-red-500 px-3 text-xs font-bold text-white active:bg-red-600 disabled:opacity-50"
              @click="emit('remove', cat.id)"
            >
              <Loader2
                v-if="deletingId === cat.id"
                class="size-4 animate-spin"
              />
              <span v-else>Eliminar</span>
            </button>
            <button
              type="button"
              class="no-min-tap h-9 rounded-xl border border-border px-3 text-xs font-bold text-muted-foreground active:bg-accent"
              @click="confirmId = null"
            >
              No
            </button>
          </template>
          <template v-else>
            <button
              type="button"
              aria-label="Editar"
              class="no-min-tap flex size-9 items-center justify-center rounded-xl text-muted-foreground active:bg-accent"
              @click="emit('edit', cat)"
            >
              <Pencil class="size-4" />
            </button>
            <button
              type="button"
              aria-label="Eliminar"
              class="no-min-tap flex size-9 items-center justify-center rounded-xl text-red-500 active:bg-red-50"
              @click="confirmId = cat.id"
            >
              <Trash2 class="size-4" />
            </button>
          </template>
        </template>
      </div>
    </div>

    <MobileFab v-if="isOwner" label="Nueva" @click="emit('create')" />
  </MobileScreen>
</template>
