<script setup lang="ts">
interface Props {
  open?: boolean;
  side?: "right" | "left";
}

const props = withDefaults(defineProps<Props>(), {
  open: false,
  side: "right",
});

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

function close() {
  emit("update:open", false);
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-200 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="props.open"
        class="fixed inset-0 z-50 flex"
        :class="props.side === 'right' ? 'justify-end' : 'justify-start'"
        role="dialog"
        aria-modal="true"
      >
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm"
          @click="close"
        />

        <!-- Panel -->
        <Transition
          enter-active-class="transition-transform duration-300 ease-out"
          :enter-from-class="
            props.side === 'right' ? 'translate-x-full' : '-translate-x-full'
          "
          enter-to-class="translate-x-0"
          leave-active-class="transition-transform duration-200 ease-in"
          leave-from-class="translate-x-0"
          :leave-to-class="
            props.side === 'right' ? 'translate-x-full' : '-translate-x-full'
          "
        >
          <div
            v-if="props.open"
            class="relative z-10 flex h-full w-full max-w-md flex-col bg-card border-l border-border shadow-2xl"
          >
            <slot />
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
