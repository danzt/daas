<script setup lang="ts">
import { Truck, FolderOpen, Loader2, ClipboardList } from "lucide-vue-next";
import type { Supplier } from "~/components/suppliers/SupplierFormModal.vue";

const props = defineProps<{
  suppliers: Supplier[];
  loading: boolean;
  loadError: string;
  isOwner: boolean;
}>();

const emit = defineEmits<{
  create: [];
  edit: [supplier: Supplier];
  retry: [];
}>();

const search = ref("");

const filtered = computed(() => {
  const q = search.value.trim().toLowerCase();
  if (!q) return props.suppliers;
  return props.suppliers.filter(
    (s) =>
      s.name.toLowerCase().includes(q) ||
      (s.rif ?? "").toLowerCase().includes(q) ||
      (s.contact_name ?? "").toLowerCase().includes(q),
  );
});

function initials(name: string) {
  return (name || "?")
    .split(" ")
    .slice(0, 2)
    .map((w) => w[0])
    .join("")
    .toUpperCase();
}
</script>

<template>
  <MobileScreen title="Proveedores" subtitle="Compras y abastecimiento">
    <template #action>
      <NuxtLink
        to="/suppliers/purchase-orders"
        aria-label="Órdenes de compra"
        class="no-min-tap flex size-10 items-center justify-center rounded-full text-foreground active:bg-accent"
      >
        <ClipboardList class="size-5" />
      </NuxtLink>
    </template>

    <MobileSearchBar v-model="search" placeholder="Buscar por nombre o RIF…" />

    <div
      v-if="loadError"
      class="mx-4 mt-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
    >
      {{ loadError }}
    </div>

    <div
      v-else-if="loading && !suppliers.length"
      class="flex justify-center py-16"
    >
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <div v-else-if="!filtered.length" class="px-6 py-16 text-center">
      <FolderOpen class="mx-auto mb-4 size-12 text-muted-foreground/20" />
      <p class="font-medium text-muted-foreground">Sin proveedores</p>
    </div>

    <div v-else class="divide-y divide-border border-t border-border bg-white">
      <MobileListItem
        v-for="sup in filtered"
        :key="sup.id"
        :chevron="true"
        :class="!sup.active ? 'opacity-55' : ''"
        @click="emit('edit', sup)"
      >
        <template #leading>
          <div
            class="flex size-11 items-center justify-center rounded-2xl bg-primary/10 text-sm font-black text-primary"
          >
            {{ initials(sup.name) }}
          </div>
        </template>

        <p class="truncate text-sm font-semibold leading-tight text-foreground">
          {{ sup.name }}
        </p>
        <div class="mt-1 flex items-center gap-1.5">
          <span
            v-if="sup.rif"
            class="rounded bg-muted px-1.5 py-px font-mono text-[10px] text-muted-foreground"
            >{{ sup.rif }}</span
          >
          <span
            v-if="sup.contact_name"
            class="truncate text-[11px] text-muted-foreground"
          >
            {{ sup.contact_name }}
          </span>
        </div>
      </MobileListItem>
    </div>

    <MobileFab v-if="isOwner" label="Nuevo" @click="emit('create')" />
  </MobileScreen>
</template>
