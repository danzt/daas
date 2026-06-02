import { defineStore } from "pinia";
import { ref, computed, watch } from "vue";

export interface CartItem {
  productId: string;
  name: string;
  price: number;
  qty: number;
  category?: string;
  is_fiscal?: boolean;
  stock_qty?: number;
}

export const useCartStore = defineStore("cart", () => {
  const route = useRoute();
  const tenantSlug = computed(
    () => (route.params.tenantSlug as string) ?? "default",
  );
  const storageKey = computed(() => `daas_cart_${tenantSlug.value}`);

  const items = ref<CartItem[]>([]);
  // Transient feedback flag — set briefly after addItem so the UI can flash an
  // animation on the cart icon.
  const lastAddedAt = ref<number>(0);

  function persist() {
    if (import.meta.server) return;
    localStorage.setItem(storageKey.value, JSON.stringify(items.value));
  }

  function restore() {
    if (import.meta.server) return;
    const raw = localStorage.getItem(storageKey.value);
    if (raw) {
      try {
        items.value = JSON.parse(raw) as CartItem[];
      } catch {
        // ignore malformed data
      }
    } else {
      items.value = [];
    }
  }

  function addItem(item: CartItem) {
    const existing = items.value.find((i) => i.productId === item.productId);
    const maxQty = item.stock_qty ?? Number.MAX_SAFE_INTEGER;
    if (existing) {
      existing.qty = Math.min(existing.qty + item.qty, maxQty);
      // refresh metadata in case price/category changed
      existing.price = item.price;
      existing.category = item.category;
      existing.is_fiscal = item.is_fiscal;
      existing.stock_qty = item.stock_qty;
    } else {
      items.value.push({ ...item, qty: Math.min(item.qty, maxQty) });
    }
    lastAddedAt.value = Date.now();
    persist();
  }

  function updateQty(productId: string, qty: number) {
    const item = items.value.find((i) => i.productId === productId);
    if (!item) return;
    const maxQty = item.stock_qty ?? Number.MAX_SAFE_INTEGER;
    item.qty = Math.max(1, Math.min(qty, maxQty));
    persist();
  }

  function removeItem(productId: string) {
    items.value = items.value.filter((i) => i.productId !== productId);
    persist();
  }

  function clearCart() {
    items.value = [];
    persist();
  }

  const totalItems = computed(() =>
    items.value.reduce((sum, i) => sum + i.qty, 0),
  );

  const subtotal = computed(() =>
    items.value.reduce((sum, i) => sum + i.price * i.qty, 0),
  );

  const isEmpty = computed(() => items.value.length === 0);

  // Re-hydrate when the tenant slug changes between routes. The initial
  // hydration is triggered explicitly by the public layout on the client.
  watch(tenantSlug, restore);

  return {
    items,
    lastAddedAt,
    restore,
    addItem,
    updateQty,
    removeItem,
    clearCart,
    totalItems,
    subtotal,
    isEmpty,
  };
});
