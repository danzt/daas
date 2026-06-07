<script setup lang="ts">
import {
  CreditCard,
  Palette,
  Users,
  Zap,
  Building2,
  ChevronRight,
  LogOut,
  ShieldCheck,
} from "lucide-vue-next";

const props = defineProps<{
  tenantName: string;
  userEmail: string;
  isOwner: boolean;
}>();

const emit = defineEmits<{ logout: [] }>();

const initials = computed(() =>
  (props.tenantName || props.userEmail || "?").slice(0, 2).toUpperCase(),
);

interface Item {
  label: string;
  desc: string;
  icon: Component;
  to: string;
  ownerOnly?: boolean;
}

const groups: { title: string; items: Item[] }[] = [
  {
    title: "Negocio",
    items: [
      {
        label: "Métodos de pago",
        desc: "Cuentas y datos de cobro",
        icon: CreditCard,
        to: "/settings/payment-methods",
        ownerOnly: true,
      },
      {
        label: "Branding",
        desc: "Logo, colores y tienda",
        icon: Palette,
        to: "/settings/branding",
        ownerOnly: true,
      },
    ],
  },
  {
    title: "Configuración",
    items: [
      {
        label: "Empresa",
        desc: "Datos fiscales del tenant",
        icon: Building2,
        to: "/settings?tab=empresa",
      },
      {
        label: "Usuarios",
        desc: "Equipo y permisos",
        icon: Users,
        to: "/settings?tab=usuarios",
        ownerOnly: true,
      },
      {
        label: "Integración fiscal",
        desc: "Conexión con el SENIAT",
        icon: Zap,
        to: "/settings?tab=integraciones",
        ownerOnly: true,
      },
    ],
  },
];

function visible(items: Item[]) {
  return items.filter((i) => !i.ownerOnly || props.isOwner);
}
</script>

<template>
  <MobileScreen title="Más" subtitle="Ajustes y cuenta">
    <!-- Account header -->
    <div class="px-4 pt-3">
      <div
        class="flex items-center gap-3 rounded-2xl border border-border bg-white p-4"
      >
        <div
          class="flex size-14 shrink-0 items-center justify-center rounded-2xl bg-primary text-lg font-black text-white"
        >
          {{ initials }}
        </div>
        <div class="min-w-0 flex-1">
          <p class="truncate text-base font-bold text-foreground">
            {{ tenantName || "Mi empresa" }}
          </p>
          <p class="truncate text-xs text-muted-foreground">{{ userEmail }}</p>
          <span
            class="mt-1 inline-flex items-center gap-1 rounded-full bg-primary/10 px-2 py-px text-[10px] font-bold text-primary"
          >
            <ShieldCheck class="size-3" />
            {{ isOwner ? "Propietario" : "Empleado" }}
          </span>
        </div>
      </div>
    </div>

    <!-- Menu groups -->
    <div v-for="group in groups" :key="group.title" class="pt-5">
      <MobileSectionHeader :title="group.title" />
      <div class="divide-y divide-border border-y border-border bg-white">
        <MobileListItem
          v-for="item in visible(group.items)"
          :key="item.to"
          :to="item.to"
        >
          <template #leading>
            <div
              class="flex size-10 items-center justify-center rounded-xl bg-muted text-muted-foreground"
            >
              <component :is="item.icon" class="size-5" />
            </div>
          </template>
          <p class="text-sm font-semibold text-foreground">{{ item.label }}</p>
          <p class="text-xs text-muted-foreground">{{ item.desc }}</p>
        </MobileListItem>
      </div>
    </div>

    <!-- Logout -->
    <div class="px-4 pt-6">
      <button
        type="button"
        class="flex w-full items-center justify-center gap-2 rounded-2xl border border-destructive/30 bg-destructive/5 py-3.5 text-sm font-bold text-destructive active:bg-destructive/10"
        @click="emit('logout')"
      >
        <LogOut class="size-5" />
        Cerrar sesión
      </button>
    </div>

    <p class="px-4 pt-4 text-center text-[11px] text-muted-foreground">
      DaaS · Gestión de inventario
    </p>
  </MobileScreen>
</template>
