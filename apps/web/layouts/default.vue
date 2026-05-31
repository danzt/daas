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
  Bell,
  Command,
  Search,
} from "lucide-vue-next";
import {
  SidebarProvider,
  Sidebar,
  SidebarInset,
  SidebarTrigger,
  SidebarHeader,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarMenuSub,
  SidebarMenuSubItem,
  SidebarMenuSubButton,
} from "~/components/ui/sidebar";
import { useAuthStore } from "~/stores/auth";

const store = useAuthStore();
const router = useRouter();
const route = useRoute();

interface NavSubItem {
  title: string;
  url: string;
}

interface NavItem {
  title: string;
  url: string;
  icon: Component;
  subItems?: NavSubItem[];
}

interface NavGroup {
  id: number;
  label: string;
  items: NavItem[];
}

const sidebarItems: NavGroup[] = [
  {
    id: 1,
    label: "Principal",
    items: [{ title: "Dashboard", url: "/dashboard", icon: LayoutDashboard }],
  },
  {
    id: 2,
    label: "Catálogo",
    items: [
      {
        title: "Productos",
        url: "/products",
        icon: Package,
        subItems: [
          { title: "Lista de productos", url: "/products" },
          { title: "Categorías", url: "/products/categories" },
        ],
      },
      {
        title: "Inventario",
        url: "/inventory",
        icon: Warehouse,
        subItems: [
          { title: "Stock actual", url: "/inventory" },
          { title: "Movimientos", url: "/inventory/movements" },
        ],
      },
    ],
  },
  {
    id: 3,
    label: "Ventas",
    items: [
      {
        title: "Órdenes de venta",
        url: "/sales-orders",
        icon: ShoppingCart,
      },
      {
        title: "Facturas",
        url: "/invoices",
        icon: FileText,
        subItems: [
          { title: "Facturas internas", url: "/invoices" },
          { title: "Facturas fiscales", url: "/invoices/fiscal" },
        ],
      },
    ],
  },
  {
    id: 4,
    label: "Operaciones",
    items: [
      {
        title: "Proveedores",
        url: "/suppliers",
        icon: Truck,
        subItems: [
          { title: "Lista de proveedores", url: "/suppliers" },
          { title: "Órdenes de compra", url: "/suppliers/purchase-orders" },
        ],
      },
    ],
  },
  {
    id: 5,
    label: "Análisis",
    items: [{ title: "Reportes", url: "/reports", icon: BarChart2 }],
  },
];

const openCollapsibles = ref<Record<string, boolean>>({});

function isItemActive(item: NavItem): boolean {
  if (item.subItems) {
    return item.subItems.some((s) => route.path.startsWith(s.url));
  }
  return route.path === item.url || route.path.startsWith(item.url + "/");
}

function isSubItemActive(url: string): boolean {
  return route.path === url || route.path.startsWith(url + "/");
}

function toggleCollapsible(key: string) {
  openCollapsibles.value[key] = !openCollapsibles.value[key];
}

function isCollapsibleOpen(item: NavItem): boolean {
  if (item.url in openCollapsibles.value) {
    return openCollapsibles.value[item.url];
  }
  return isItemActive(item);
}

async function handleLogout() {
  await store.logout();
  router.push("/auth/login");
}

const userInitials = computed(() => {
  const email = store.user?.email ?? "";
  return email.slice(0, 2).toUpperCase();
});

const tenantLabel = computed(
  () => store.tenant?.name || store.user?.email || "",
);
const roleLabel = computed(() =>
  store.user?.role === "owner" ? "Propietario" : "Empleado",
);
</script>

<template>
  <SidebarProvider :default-open="true">
    <Sidebar variant="inset" collapsible="icon">
      <!-- ── Header ─── -->
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <NuxtLink to="/dashboard" class="block">
              <SidebarMenuButton :is-active="false">
                <Command class="size-4" />
                <span class="font-semibold text-base">DaaS</span>
              </SidebarMenuButton>
            </NuxtLink>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <!-- ── Content ─── -->
      <SidebarContent>
        <SidebarGroup v-for="group in sidebarItems" :key="group.id">
          <SidebarGroupLabel>{{ group.label }}</SidebarGroupLabel>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in group.items" :key="item.title">
              <!-- Item with sub-items -->
              <template v-if="item.subItems">
                <SidebarMenuButton
                  :is-active="isItemActive(item)"
                  :tooltip="item.title"
                  @click="toggleCollapsible(item.url)"
                >
                  <component :is="item.icon" />
                  <span>{{ item.title }}</span>
                  <ChevronRight
                    class="ml-auto size-4 transition-transform duration-200"
                    :class="isCollapsibleOpen(item) ? 'rotate-90' : ''"
                  />
                </SidebarMenuButton>
                <SidebarMenuSub v-if="isCollapsibleOpen(item)">
                  <SidebarMenuSubItem
                    v-for="sub in item.subItems"
                    :key="sub.url"
                  >
                    <NuxtLink :to="sub.url" class="block">
                      <SidebarMenuSubButton
                        :is-active="isSubItemActive(sub.url)"
                      >
                        <span>{{ sub.title }}</span>
                      </SidebarMenuSubButton>
                    </NuxtLink>
                  </SidebarMenuSubItem>
                </SidebarMenuSub>
              </template>

              <!-- Simple item -->
              <NuxtLink v-else :to="item.url" class="block">
                <SidebarMenuButton
                  :is-active="isItemActive(item)"
                  :tooltip="item.title"
                >
                  <component :is="item.icon" />
                  <span>{{ item.title }}</span>
                </SidebarMenuButton>
              </NuxtLink>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>

      <!-- ── Footer ─── -->
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <NuxtLink to="/settings" class="block">
              <SidebarMenuButton
                :is-active="route.path.startsWith('/settings')"
              >
                <Settings />
                <span>Configuración</span>
              </SidebarMenuButton>
            </NuxtLink>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg">
              <div
                class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground text-xs font-semibold shrink-0"
              >
                {{ userInitials }}
              </div>
              <div class="grid flex-1 text-left text-sm leading-tight">
                <span class="truncate font-medium">{{ tenantLabel }}</span>
                <span class="truncate text-xs text-muted-foreground">
                  {{ roleLabel }}
                </span>
              </div>
              <button
                type="button"
                title="Cerrar sesión"
                class="ml-auto rounded-md p-1 text-muted-foreground hover:bg-destructive/10 hover:text-destructive transition-colors"
                @click.stop="handleLogout"
              >
                <LogOut class="size-4" />
              </button>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>

    <SidebarInset>
      <header
        class="flex h-12 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear"
      >
        <div class="flex w-full items-center justify-between px-4 lg:px-6">
          <div class="flex items-center gap-1 lg:gap-2">
            <SidebarTrigger />
            <div class="mx-2 h-4 w-px bg-border" />
            <button
              type="button"
              class="inline-flex items-center gap-2 rounded-md border bg-background px-2.5 h-7 text-xs text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
            >
              <Search class="size-3.5" />
              <span>Buscar…</span>
              <kbd
                class="ml-2 hidden sm:inline-flex items-center gap-0.5 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground"
              >
                <span>⌘</span>K
              </kbd>
            </button>
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="inline-flex h-7 w-7 items-center justify-center rounded-md text-foreground hover:bg-accent transition-colors"
            >
              <Bell class="size-4" />
            </button>
          </div>
        </div>
      </header>

      <div class="flex-1 overflow-y-auto">
        <slot />
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>
