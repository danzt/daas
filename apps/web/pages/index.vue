<script setup lang="ts">
import { Package } from 'lucide-vue-next'
</script>

<template>
  <main class="min-h-screen bg-background flex flex-col items-center justify-center gap-md p-md">
    <div class="flex items-center gap-sm">
      <Package :size="32" class="text-primary" />
      <h1 class="font-heading text-2xl font-bold text-primary">
        DaaS
      </h1>
    </div>
    <p class="font-body text-text-brand">
      Inventory Management SaaS — Sprint 0
    </p>
    <Button variant="default" class="bg-cta text-white hover:opacity-90 transition-all">
      Get Started
    </Button>
  </main>
</template>
