import { defineStore } from "pinia";
import { ref, computed, watch } from "vue";

export interface CartItem {
  productId: string;
  name: string;
  price: number;
  qty: number;
}

export const useCartStore = defineStore("cart", () => {
  const route = useRoute();
  const tenantSlug = computed(
    () => (route.params.tenantSlug as string) ?? "default",
  );
  const storageKey = computed(() => `daas_cart_${tenantSlug.value}`);

  const items = ref<CartItem[]>([]);

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
    }
  }

  function addItem(item: CartItem) {
    const existing = items.value.find((i) => i.productId === item.productId);
    if (existing) {
      existing.qty += item.qty;
    } else {
      items.value.push(item);
    }
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

  // Auto-restore on tenant slug change
  watch(tenantSlug, () => restore(), { immediate: true });

  return { items, addItem, removeItem, clearCart, totalItems };
});
