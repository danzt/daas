<script setup lang="ts">
interface Props {
  value: string;
  class?: string;
}

const props = defineProps<Props>();

const activeValue = inject<Ref<string | undefined>>("tabs-value");
const setTab = inject<(v: string) => void>("tabs-set");

const isActive = computed(() => activeValue?.value === props.value);
</script>

<template>
  <button
    type="button"
    role="tab"
    :aria-selected="isActive"
    :class="[
      'px-4 py-2.5 text-sm font-semibold transition-all duration-200 cursor-pointer border-b-2 -mb-px focus-visible:outline-2 focus-visible:outline-primary',
      isActive
        ? 'border-primary text-primary'
        : 'border-transparent text-gray-500 hover:text-primary hover:border-primary/40',
      props.class,
    ]"
    @click="setTab?.(props.value)"
  >
    <slot />
  </button>
</template>
