<script setup lang="ts">
import {
  TrendingUp,
  TrendingDown,
  SlidersHorizontal,
  Loader2,
  AlertTriangle,
  FolderOpen,
} from "lucide-vue-next";
import type { Product } from "~/components/products/ProductFormModal.vue";

interface Movement {
  id: string;
  product_id: string;
  type: "entry" | "exit" | "adjustment";
  quantity: number;
  unit_cost?: number;
  reference_type: string;
  notes: string;
  created_at: string;
}

const props = defineProps<{
  movements: Movement[];
  products: Product[];
  loading: boolean;
  loadError: string;
}>();

const typeFilter = ref("all");

const TYPE: Record<
  string,
  { label: string; bubble: string; fg: string; icon: Component; sign: string }
> = {
  entry: {
    label: "Entrada",
    bubble: "bg-emerald-50",
    fg: "text-emerald-600",
    icon: TrendingUp,
    sign: "+",
  },
  exit: {
    label: "Salida",
    bubble: "bg-red-50",
    fg: "text-red-500",
    icon: TrendingDown,
    sign: "−",
  },
  adjustment: {
    label: "Ajuste",
    bubble: "bg-blue-50",
    fg: "text-blue-600",
    icon: SlidersHorizontal,
    sign: "",
  },
};

const productMap = computed<Record<string, Product>>(() => {
  const m: Record<string, Product> = {};
  for (const p of props.products) m[p.id] = p;
  return m;
});

const segOptions = computed(() => [
  { key: "all", label: "Todos", count: props.movements.length },
  {
    key: "entry",
    label: "Entradas",
    count: props.movements.filter((m) => m.type === "entry").length,
  },
  {
    key: "exit",
    label: "Salidas",
    count: props.movements.filter((m) => m.type === "exit").length,
  },
  {
    key: "adjustment",
    label: "Ajustes",
    count: props.movements.filter((m) => m.type === "adjustment").length,
  },
]);

const filtered = computed(() =>
  typeFilter.value === "all"
    ? props.movements
    : props.movements.filter((m) => m.type === typeFilter.value),
);

function productName(id: string) {
  return productMap.value[id]?.name ?? "Producto eliminado";
}
function fmtDate(d: string) {
  return new Date(d).toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}
</script>

<template>
  <MobileScreen title="Movimientos" back subtitle="Historial de inventario">
    <MobileSegment v-model="typeFilter" :options="segOptions" />

    <div v-if="loadError" class="px-4 py-12 text-center">
      <AlertTriangle class="mx-auto mb-3 size-10 text-destructive" />
      <p class="font-medium text-destructive">{{ loadError }}</p>
    </div>

    <div
      v-else-if="loading && !movements.length"
      class="flex justify-center py-16"
    >
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <div v-else-if="!filtered.length" class="px-6 py-16 text-center">
      <FolderOpen class="mx-auto mb-4 size-12 text-muted-foreground/20" />
      <p class="font-medium text-muted-foreground">Sin movimientos</p>
    </div>

    <div v-else class="divide-y divide-border border-t border-border bg-white">
      <div
        v-for="mv in filtered"
        :key="mv.id"
        class="flex items-center gap-3 px-4 py-3.5"
      >
        <div
          class="flex size-11 shrink-0 items-center justify-center rounded-2xl"
          :class="TYPE[mv.type]?.bubble"
        >
          <component
            :is="TYPE[mv.type]?.icon"
            class="size-5"
            :class="TYPE[mv.type]?.fg"
          />
        </div>
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm font-semibold text-foreground">
            {{ productName(mv.product_id) }}
          </p>
          <div class="mt-0.5 flex items-center gap-1.5">
            <span class="text-[11px] font-medium" :class="TYPE[mv.type]?.fg">{{
              TYPE[mv.type]?.label
            }}</span>
            <span class="text-[11px] text-muted-foreground">{{
              fmtDate(mv.created_at)
            }}</span>
          </div>
        </div>
        <span
          class="font-mono text-base font-black tabular-nums"
          :class="TYPE[mv.type]?.fg"
        >
          {{ TYPE[mv.type]?.sign }}{{ Math.abs(mv.quantity) }}
        </span>
      </div>
    </div>
  </MobileScreen>
</template>
