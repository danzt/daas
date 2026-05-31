<script setup lang="ts">
import {
  LayoutDashboard,
  Package,
  Warehouse,
  ShoppingCart,
  FileText,
  Truck,
  BarChart2,
  Settings,
  LogOut,
  ChevronRight,
  Menu,
  X,
  Tag,
  History,
  Bell,
  Search,
  ChevronsUpDown,
} from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";

const store = useAuthStore();
const router = useRouter();
const route = useRoute();

const sidebarOpen = ref(true);
const mobileSidebarOpen = ref(false);

interface NavItem {
  label: string;
  icon: Component;
  to: string;
  children?: { label: string; to: string }[];
}

interface NavGroup {
  id: string;
  label: string;
  items: NavItem[];
}

const navGroups: NavGroup[] = [
  {
    id: "main",
    label: "Principal",
    items: [{ label: "Dashboard", icon: LayoutDashboard, to: "/dashboard" }],
  },
  {
    id: "catalog",
    label: "Catálogo",
    items: [
      {
        label: "Productos",
        icon: Package,
        to: "/products",
        children: [
          { label: "Lista de productos", to: "/products" },
          { label: "Categorías", to: "/products/categories" },
        ],
      },
      {
        label: "Inventario",
        icon: Warehouse,
        to: "/inventory",
        children: [
          { label: "Stock actual", to: "/inventory" },
          { label: "Movimientos", to: "/inventory/movements" },
        ],
      },
    ],
  },
  {
    id: "sales",
    label: "Ventas",
    items: [
      {
        label: "Ventas",
        icon: ShoppingCart,
        to: "/sales-orders",
        children: [{ label: "Órdenes de venta", to: "/sales-orders" }],
      },
      {
        label: "Facturas",
        icon: FileText,
        to: "/invoices",
        children: [
          { label: "Facturas internas", to: "/invoices" },
          { label: "Facturas fiscales", to: "/invoices/fiscal" },
        ],
      },
    ],
  },
  {
    id: "ops",
    label: "Operaciones",
    items: [
      {
        label: "Proveedores",
        icon: Truck,
        to: "/suppliers",
        children: [
          { label: "Lista de proveedores", to: "/suppliers" },
          { label: "Órdenes de compra", to: "/suppliers/purchase-orders" },
        ],
      },
    ],
  },
  {
    id: "reports",
    label: "Análisis",
    items: [
      {
        label: "Reportes",
        icon: BarChart2,
        to: "/reports",
        children: [
          { label: "Ventas", to: "/reports" },
          { label: "Inventario", to: "/reports?tab=inventory" },
          { label: "Compras", to: "/reports?tab=purchases" },
        ],
      },
    ],
  },
];

// Track which collapsible items are open
const openItems = ref<Record<string, boolean>>({});

function isItemActive(item: NavItem): boolean {
  if (item.children) {
    return item.children.some((c) => route.path.startsWith(c.to));
  }
  return route.path === item.to || route.path.startsWith(item.to + "/");
}

function isChildActive(to: string): boolean {
  return route.path === to || route.path.startsWith(to + "/");
}

function toggleItem(id: string) {
  openItems.value[id] = !openItems.value[id];
}

function isItemOpen(item: NavItem): boolean {
  const key = item.to;
  if (key in openItems.value) return openItems.value[key];
  return isItemActive(item);
}

async function handleLogout() {
  await store.logout();
  router.push("/auth/login");
}

function closeMobile() {
  mobileSidebarOpen.value = false;
}

const userInitials = computed(() => {
  const email = store.user?.email ?? "";
  return email.slice(0, 2).toUpperCase();
});

const roleLabel = computed(() =>
  store.user?.role === "owner" ? "Propietario" : "Empleado",
);
</script>

<template>
  <div class="flex h-screen bg-[hsl(var(--background))] overflow-hidden">
    <!-- ═══ Desktop Sidebar ═══════════════════════════════════════════════════ -->
    <aside
      :class="[
        'hidden md:flex flex-col border-r bg-sidebar border-sidebar-border transition-all duration-300 ease-in-out',
        sidebarOpen ? 'w-64' : 'w-16',
      ]"
    >
      <!-- Logo -->
      <div
        class="flex items-center h-14 px-4 border-b border-sidebar-border flex-shrink-0"
      >
        <div
          class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center flex-shrink-0"
        >
          <span class="text-white text-xs font-bold font-heading">D</span>
        </div>
        <Transition
          enter-active-class="transition-all duration-200"
          enter-from-class="opacity-0 translate-x-2"
          enter-to-class="opacity-100 translate-x-0"
          leave-active-class="transition-all duration-100"
          leave-from-class="opacity-100"
          leave-to-class="opacity-0"
        >
          <span
            v-if="sidebarOpen"
            class="ml-3 text-lg font-bold font-heading text-[hsl(var(--foreground))]"
          >
            DaaS
          </span>
        </Transition>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 overflow-y-auto py-3 px-2">
        <div v-for="group in navGroups" :key="group.id" class="mb-4">
          <!-- Group label -->
          <Transition
            enter-active-class="transition-opacity duration-200"
            enter-from-class="opacity-0"
            enter-to-class="opacity-100"
            leave-active-class="transition-opacity duration-100"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
          >
            <p
              v-if="sidebarOpen"
              class="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground px-3 mb-1"
            >
              {{ group.label }}
            </p>
            <div v-else class="h-4 mb-1 flex items-center justify-center">
              <div class="w-4 h-px bg-sidebar-border" />
            </div>
          </Transition>

          <!-- Items -->
          <div v-for="item in group.items" :key="item.to" class="mb-0.5">
            <!-- Item with children -->
            <template v-if="item.children && sidebarOpen">
              <button
                type="button"
                :class="[
                  'w-full flex items-center gap-2.5 h-9 px-3 rounded-md text-sm font-medium transition-colors duration-150 cursor-pointer group',
                  isItemActive(item)
                    ? 'bg-primary/10 text-primary'
                    : 'text-[hsl(var(--sidebar-foreground))] hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                ]"
                @click="toggleItem(item.to)"
              >
                <component
                  :is="item.icon"
                  class="w-4 h-4 flex-shrink-0"
                  :class="
                    isItemActive(item)
                      ? 'text-primary'
                      : 'text-muted-foreground group-hover:text-sidebar-accent-foreground'
                  "
                />
                <span class="flex-1 text-left">{{ item.label }}</span>
                <ChevronRight
                  class="w-3.5 h-3.5 transition-transform duration-200 text-muted-foreground"
                  :class="isItemOpen(item) ? 'rotate-90' : ''"
                />
              </button>
              <!-- Children -->
              <Transition
                enter-active-class="transition-all duration-200 overflow-hidden"
                enter-from-class="max-h-0 opacity-0"
                enter-to-class="max-h-40 opacity-100"
                leave-active-class="transition-all duration-150 overflow-hidden"
                leave-from-class="max-h-40 opacity-100"
                leave-to-class="max-h-0 opacity-0"
              >
                <div
                  v-if="isItemOpen(item)"
                  class="ml-4 mt-0.5 border-l border-sidebar-border pl-3 space-y-0.5"
                >
                  <NuxtLink
                    v-for="child in item.children"
                    :key="child.to"
                    :to="child.to"
                    :class="[
                      'flex items-center h-8 px-2 rounded-md text-sm transition-colors duration-150',
                      isChildActive(child.to)
                        ? 'text-primary font-medium'
                        : 'text-muted-foreground hover:text-[hsl(var(--foreground))] hover:bg-sidebar-accent',
                    ]"
                  >
                    {{ child.label }}
                  </NuxtLink>
                </div>
              </Transition>
            </template>

            <!-- Simple item -->
            <NuxtLink
              v-else
              :to="item.to"
              :title="!sidebarOpen ? item.label : undefined"
              :class="[
                'flex items-center gap-2.5 h-9 rounded-md text-sm font-medium transition-colors duration-150 group',
                sidebarOpen ? 'px-3' : 'px-2.5 justify-center',
                isItemActive(item)
                  ? 'bg-primary/10 text-primary'
                  : 'text-[hsl(var(--sidebar-foreground))] hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
              ]"
            >
              <component
                :is="item.icon"
                class="w-4 h-4 flex-shrink-0"
                :class="
                  isItemActive(item)
                    ? 'text-primary'
                    : 'text-muted-foreground group-hover:text-sidebar-accent-foreground'
                "
              />
              <span v-if="sidebarOpen">{{ item.label }}</span>
            </NuxtLink>
          </div>
        </div>
      </nav>

      <!-- Settings link -->
      <div class="px-2 pb-2 border-t border-sidebar-border pt-2">
        <NuxtLink
          to="/settings"
          :title="!sidebarOpen ? 'Configuración' : undefined"
          :class="[
            'flex items-center gap-2.5 h-9 rounded-md text-sm font-medium transition-colors duration-150 group',
            sidebarOpen ? 'px-3' : 'px-2.5 justify-center',
            route.path.startsWith('/settings')
              ? 'bg-primary/10 text-primary'
              : 'text-[hsl(var(--sidebar-foreground))] hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
          ]"
        >
          <Settings
            class="w-4 h-4 flex-shrink-0"
            :class="
              route.path.startsWith('/settings')
                ? 'text-primary'
                : 'text-muted-foreground group-hover:text-sidebar-accent-foreground'
            "
          />
          <span v-if="sidebarOpen">Configuración</span>
        </NuxtLink>
      </div>

      <!-- User footer -->
      <div class="px-2 pb-3 border-t border-sidebar-border pt-3">
        <div
          v-if="sidebarOpen"
          class="flex items-center gap-2.5 px-2 py-2 rounded-lg hover:bg-sidebar-accent transition-colors duration-150 group cursor-default"
        >
          <div
            class="w-8 h-8 rounded-full bg-primary flex items-center justify-center flex-shrink-0 text-white text-xs font-bold"
          >
            {{ userInitials }}
          </div>
          <div class="flex-1 min-w-0">
            <p
              class="text-xs font-semibold text-[hsl(var(--foreground))] truncate"
            >
              {{ store.tenant?.name || store.user?.email }}
            </p>
            <p class="text-[10px] text-muted-foreground">{{ roleLabel }}</p>
          </div>
          <button
            type="button"
            title="Cerrar sesión"
            class="w-7 h-7 flex items-center justify-center rounded-md text-muted-foreground hover:text-destructive hover:bg-red-50 transition-colors duration-150 cursor-pointer opacity-0 group-hover:opacity-100"
            @click="handleLogout"
          >
            <LogOut class="w-3.5 h-3.5" />
          </button>
        </div>
        <button
          v-else
          type="button"
          title="Cerrar sesión"
          class="w-full flex justify-center items-center h-9 rounded-md text-muted-foreground hover:text-destructive hover:bg-red-50 transition-colors duration-150 cursor-pointer"
          @click="handleLogout"
        >
          <LogOut class="w-4 h-4" />
        </button>
      </div>
    </aside>

    <!-- ═══ Main area ═════════════════════════════════════════════════════════ -->
    <div class="flex-1 flex flex-col min-h-0 overflow-hidden">
      <!-- Header -->
      <header
        class="h-14 flex items-center justify-between px-4 lg:px-6 border-b bg-[hsl(var(--card))] flex-shrink-0"
      >
        <!-- Left: sidebar toggle + breadcrumb area -->
        <div class="flex items-center gap-3">
          <!-- Desktop toggle -->
          <button
            type="button"
            class="hidden md:flex w-8 h-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-[hsl(var(--foreground))] transition-colors duration-150 cursor-pointer"
            @click="sidebarOpen = !sidebarOpen"
          >
            <Menu class="w-4 h-4" />
          </button>
          <!-- Mobile toggle -->
          <button
            type="button"
            class="md:hidden w-8 h-8 flex items-center justify-center rounded-md text-muted-foreground hover:bg-muted cursor-pointer"
            @click="mobileSidebarOpen = true"
          >
            <Menu class="w-4 h-4" />
          </button>

          <!-- Logo on mobile -->
          <span class="md:hidden text-base font-bold font-heading text-primary">
            DaaS
          </span>
        </div>

        <!-- Right: actions -->
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="w-8 h-8 flex items-center justify-center rounded-md text-muted-foreground hover:bg-muted transition-colors duration-150 cursor-pointer"
          >
            <Bell class="w-4 h-4" />
          </button>
          <!-- User avatar -->
          <button
            type="button"
            class="flex items-center gap-2 pl-1 pr-2 h-8 rounded-lg hover:bg-muted transition-colors duration-150 cursor-pointer"
            @click="handleLogout"
          >
            <div
              class="w-6 h-6 rounded-full bg-primary flex items-center justify-center text-white text-[10px] font-bold"
            >
              {{ userInitials }}
            </div>
            <span
              class="hidden sm:block text-xs font-medium text-[hsl(var(--foreground))] max-w-[120px] truncate"
            >
              {{ store.tenant?.name || store.user?.email }}
            </span>
          </button>
        </div>
      </header>

      <!-- Page content -->
      <main class="flex-1 overflow-y-auto">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <slot />
        </div>
      </main>
    </div>

    <!-- ═══ Mobile drawer ══════════════════════════════════════════════════════ -->
    <Transition
      enter-active-class="transition-opacity duration-300"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="mobileSidebarOpen" class="md:hidden fixed inset-0 z-50">
        <div
          class="absolute inset-0 bg-black/40 backdrop-blur-sm"
          @click="closeMobile"
        />
        <Transition
          enter-active-class="transition-transform duration-300"
          enter-from-class="-translate-x-full"
          enter-to-class="translate-x-0"
          leave-active-class="transition-transform duration-200"
          leave-from-class="translate-x-0"
          leave-to-class="-translate-x-full"
        >
          <aside
            v-if="mobileSidebarOpen"
            class="relative w-72 h-full bg-sidebar flex flex-col shadow-xl"
          >
            <!-- Mobile header -->
            <div
              class="flex items-center justify-between h-14 px-4 border-b border-sidebar-border"
            >
              <div class="flex items-center gap-2.5">
                <div
                  class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center"
                >
                  <span class="text-white text-xs font-bold font-heading"
                    >D</span
                  >
                </div>
                <span
                  class="text-lg font-bold font-heading text-[hsl(var(--foreground))]"
                  >DaaS</span
                >
              </div>
              <button
                type="button"
                class="w-8 h-8 flex items-center justify-center rounded-md text-muted-foreground hover:bg-muted cursor-pointer"
                @click="closeMobile"
              >
                <X class="w-4 h-4" />
              </button>
            </div>

            <!-- Mobile nav -->
            <nav class="flex-1 overflow-y-auto py-3 px-2">
              <div v-for="group in navGroups" :key="group.id" class="mb-4">
                <p
                  class="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground px-3 mb-1"
                >
                  {{ group.label }}
                </p>
                <div v-for="item in group.items" :key="item.to" class="mb-0.5">
                  <template v-if="item.children">
                    <button
                      type="button"
                      :class="[
                        'w-full flex items-center gap-2.5 h-9 px-3 rounded-md text-sm font-medium transition-colors duration-150 cursor-pointer group',
                        isItemActive(item)
                          ? 'bg-primary/10 text-primary'
                          : 'text-[hsl(var(--sidebar-foreground))] hover:bg-sidebar-accent',
                      ]"
                      @click="toggleItem(item.to)"
                    >
                      <component
                        :is="item.icon"
                        class="w-4 h-4 flex-shrink-0 text-muted-foreground"
                      />
                      <span class="flex-1 text-left">{{ item.label }}</span>
                      <ChevronRight
                        class="w-3.5 h-3.5 text-muted-foreground transition-transform duration-200"
                        :class="isItemOpen(item) ? 'rotate-90' : ''"
                      />
                    </button>
                    <div
                      v-if="isItemOpen(item)"
                      class="ml-4 mt-0.5 border-l border-sidebar-border pl-3 space-y-0.5"
                    >
                      <NuxtLink
                        v-for="child in item.children"
                        :key="child.to"
                        :to="child.to"
                        :class="[
                          'flex items-center h-8 px-2 rounded-md text-sm transition-colors duration-150',
                          isChildActive(child.to)
                            ? 'text-primary font-medium'
                            : 'text-muted-foreground hover:text-[hsl(var(--foreground))]',
                        ]"
                        @click="closeMobile"
                      >
                        {{ child.label }}
                      </NuxtLink>
                    </div>
                  </template>

                  <NuxtLink
                    v-else
                    :to="item.to"
                    :class="[
                      'flex items-center gap-2.5 h-9 px-3 rounded-md text-sm font-medium transition-colors duration-150 group',
                      isItemActive(item)
                        ? 'bg-primary/10 text-primary'
                        : 'text-[hsl(var(--sidebar-foreground))] hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                    ]"
                    @click="closeMobile"
                  >
                    <component
                      :is="item.icon"
                      class="w-4 h-4 flex-shrink-0 text-muted-foreground"
                    />
                    <span>{{ item.label }}</span>
                  </NuxtLink>
                </div>
              </div>
            </nav>

            <!-- Mobile footer -->
            <div class="px-2 pb-4 pt-2 border-t border-sidebar-border">
              <NuxtLink
                to="/settings"
                class="flex items-center gap-2.5 h-9 px-3 rounded-md text-sm font-medium text-[hsl(var(--sidebar-foreground))] hover:bg-sidebar-accent transition-colors duration-150"
                @click="closeMobile"
              >
                <Settings class="w-4 h-4 text-muted-foreground" />
                Configuración
              </NuxtLink>
              <button
                type="button"
                class="w-full mt-1 flex items-center gap-2.5 h-9 px-3 rounded-md text-sm font-medium text-destructive hover:bg-red-50 transition-colors duration-150 cursor-pointer"
                @click="handleLogout"
              >
                <LogOut class="w-4 h-4" />
                Cerrar sesión
              </button>
            </div>
          </aside>
        </Transition>
      </div>
    </Transition>
  </div>
</template>
