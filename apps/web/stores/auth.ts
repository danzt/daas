import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface User {
  id: string
  email: string
  name: string
}

export interface Tenant {
  id: string
  name: string
  slug: string
  country: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const tenant = ref<Tenant | null>(null)

  const isAuthenticated = computed(() => user.value !== null)

  async function login(_email: string, _password: string): Promise<void> {
    // TODO: implement real login via Supabase Auth in Sprint 1
    throw new Error('Not implemented')
  }

  async function logout(): Promise<void> {
    // TODO: implement real logout via Supabase Auth in Sprint 1
    user.value = null
    tenant.value = null
  }

  return {
    user,
    tenant,
    isAuthenticated,
    login,
    logout,
  }
})
