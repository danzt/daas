<script setup lang="ts">
import {
  ShoppingBag,
  FileText,
  ClipboardList,
  AlertTriangle,
} from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";

definePageMeta({
  layout: "default",
  middleware: "auth",
});

const store = useAuthStore();

const stats = [
  {
    label: "Ventas hoy",
    value: "0",
    icon: ShoppingBag,
    color: "text-cta",
    bg: "bg-cta/10",
  },
  {
    label: "Facturas fiscales",
    value: "0",
    icon: FileText,
    color: "text-primary",
    bg: "bg-primary/10",
  },
  {
    label: "Facturas internas",
    value: "0",
    icon: ClipboardList,
    color: "text-secondary",
    bg: "bg-secondary/10",
  },
  {
    label: "Stock bajo",
    value: "0",
    icon: AlertTriangle,
    color: "text-yellow-600",
    bg: "bg-yellow-100",
  },
];
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="mb-8">
      <h1 class="text-2xl font-bold font-heading text-text-brand">Dashboard</h1>
      <p class="text-muted-foreground mt-1">
        Bienvenido,
        <span class="font-semibold text-text-brand">{{
          store.tenant?.name || store.user?.email
        }}</span>
      </p>
    </div>

    <!-- Welcome card -->
    <div
      class="bg-primary rounded-xl p-6 mb-8 flex items-center justify-between"
    >
      <div>
        <h2 class="text-xl font-bold font-heading text-white">
          Bienvenido a DaaS
        </h2>
        <p class="text-white/80 text-sm mt-1">
          Tu plataforma de gestion de inventario
        </p>
      </div>
      <div class="hidden sm:block">
        <span class="text-white/20 text-7xl font-bold font-heading select-none"
          >DaaS</span
        >
      </div>
    </div>

    <!-- Stats grid -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div
        v-for="stat in stats"
        :key="stat.label"
        class="bg-card rounded-xl border shadow-sm p-5"
      >
        <div class="flex items-center justify-between mb-3">
          <span
            class="text-xs font-semibold text-muted-foreground uppercase tracking-wide"
          >
            {{ stat.label }}
          </span>
          <div
            :class="[
              'w-8 h-8 rounded-lg flex items-center justify-center',
              stat.bg,
            ]"
          >
            <component :is="stat.icon" :class="['w-4 h-4', stat.color]" />
          </div>
        </div>
        <p :class="['text-3xl font-bold font-heading', stat.color]">
          {{ stat.value }}
        </p>
      </div>
    </div>
  </div>
</template>
