<script setup lang="ts">
interface Props {
  value: string;
  class?: string;
}

const props = defineProps<Props>();

const activeValue = inject<Ref<string | undefined>>("tabs-value");
const isActive = computed(() => activeValue?.value === props.value);
</script>

<template>
  <div v-if="isActive" :class="['mt-6', props.class]" role="tabpanel">
    <slot />
  </div>
</template>
