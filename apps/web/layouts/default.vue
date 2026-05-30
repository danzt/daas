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
  Menu,
  X,
} from "lucide-vue-next";
import { useAuthStore } from "~/stores/auth";

const store = useAuthStore();
const router = useRouter();
const mobileMenuOpen = ref(false);

const navItems = [
  { label: "Dashboard", icon: LayoutDashboard, to: "/dashboard" },
  { label: "Productos", icon: Package, to: "/products" },
  { label: "Inventario", icon: Warehouse, to: "/inventory" },
  { label: "Ventas", icon: ShoppingCart, to: "/sales" },
  { label: "Facturas", icon: FileText, to: "/invoices" },
  { label: "Proveedores", icon: Truck, to: "/suppliers" },
  { label: "Reportes", icon: BarChart2, to: "/reports" },
  { label: "Configuración", icon: Settings, to: "/settings" },
];

async function handleLogout() {
  await store.logout();
  router.push("/auth/login");
}
</script>

<template>
  <div class="min-h-screen bg-background flex">
    <!-- Desktop Sidebar -->
    <aside
      class="hidden md:flex flex-col w-64 bg-white border-r border-gray-100 shadow-sm fixed inset-y-0 left-0 z-30"
    >
      <!-- Logo -->
      <div class="flex items-center px-6 py-5 border-b border-gray-100">
        <span class="text-2xl font-bold font-heading text-primary">DaaS</span>
      </div>

      <!-- Nav -->
      <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        <NuxtLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 cursor-pointer group"
          :active-class="'bg-primary/10 text-primary'"
          :inactive-class="'text-gray-600 hover:bg-primary/5 hover:text-primary'"
        >
          <component
            :is="item.icon"
            class="w-5 h-5 flex-shrink-0 transition-colors duration-200"
          />
          <span>{{ item.label }}</span>
        </NuxtLink>
      </nav>

      <!-- User info + logout -->
      <div class="px-3 py-4 border-t border-gray-100">
        <div class="px-3 py-2 rounded-lg bg-background">
          <p
            class="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-0.5"
          >
            {{ store.user?.role === "owner" ? "Propietario" : "Empleado" }}
          </p>
          <p class="text-sm font-medium text-text-brand truncate">
            {{ store.user?.email }}
          </p>
        </div>
        <button
          type="button"
          class="mt-2 flex items-center gap-2 w-full px-3 py-2.5 rounded-lg text-sm font-medium text-red-600 hover:bg-red-50 transition-all duration-200 cursor-pointer"
          @click="handleLogout"
        >
          <LogOut class="w-4 h-4" />
          <span>Cerrar sesión</span>
        </button>
      </div>
    </aside>

    <!-- Mobile Header -->
    <header
      class="md:hidden fixed top-0 left-0 right-0 z-30 bg-white border-b border-gray-100 shadow-sm h-14 flex items-center px-4 justify-between"
    >
      <span class="text-xl font-bold font-heading text-primary">DaaS</span>
      <button
        type="button"
        class="p-2 rounded-lg text-gray-600 hover:bg-gray-100 transition-all duration-200 cursor-pointer h-11 w-11 flex items-center justify-center"
        @click="mobileMenuOpen = !mobileMenuOpen"
      >
        <component :is="mobileMenuOpen ? X : Menu" class="w-5 h-5" />
      </button>
    </header>

    <!-- Mobile Drawer -->
    <Transition
      enter-active-class="transition-transform duration-300"
      enter-from-class="-translate-x-full"
      enter-to-class="translate-x-0"
      leave-active-class="transition-transform duration-300"
      leave-from-class="translate-x-0"
      leave-to-class="-translate-x-full"
    >
      <div v-if="mobileMenuOpen" class="md:hidden fixed inset-0 z-40">
        <div
          class="absolute inset-0 bg-black/40"
          @click="mobileMenuOpen = false"
        />
        <aside class="relative w-72 h-full bg-white flex flex-col shadow-xl">
          <div
            class="flex items-center justify-between px-6 py-5 border-b border-gray-100"
          >
            <span class="text-2xl font-bold font-heading text-primary"
              >DaaS</span
            >
            <button
              type="button"
              class="p-2 rounded-lg text-gray-500 hover:bg-gray-100 cursor-pointer"
              @click="mobileMenuOpen = false"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
          <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
            <NuxtLink
              v-for="item in navItems"
              :key="item.to"
              :to="item.to"
              class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 cursor-pointer"
              :active-class="'bg-primary/10 text-primary'"
              :inactive-class="'text-gray-600 hover:bg-primary/5 hover:text-primary'"
              @click="mobileMenuOpen = false"
            >
              <component :is="item.icon" class="w-5 h-5 flex-shrink-0" />
              <span>{{ item.label }}</span>
            </NuxtLink>
          </nav>
          <div class="px-3 py-4 border-t border-gray-100">
            <p class="text-sm font-medium text-text-brand px-3 truncate">
              {{ store.user?.email }}
            </p>
            <button
              type="button"
              class="mt-2 flex items-center gap-2 w-full px-3 py-2.5 rounded-lg text-sm font-medium text-red-600 hover:bg-red-50 transition-all duration-200 cursor-pointer"
              @click="handleLogout"
            >
              <LogOut class="w-4 h-4" />
              <span>Cerrar sesión</span>
            </button>
          </div>
        </aside>
      </div>
    </Transition>

    <!-- Main content area -->
    <main class="flex-1 md:ml-64 pt-14 md:pt-0 min-h-screen">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <slot />
      </div>
    </main>
  </div>
</template>
