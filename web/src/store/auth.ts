import { create } from 'zustand'
import type { AuthResponse } from '@/lib/types'

const TOKEN_KEY = 'geomap_token'

interface AuthState {
  token: string | null
  setToken: (token: string | null) => void
  signIn: (publicKey: string, privateKey: string) => Promise<void>
  signOut: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: typeof localStorage !== 'undefined' ? localStorage.getItem(TOKEN_KEY) : null,

  setToken: (token) => {
    if (token) {
      localStorage.setItem(TOKEN_KEY, token)
    } else {
      localStorage.removeItem(TOKEN_KEY)
    }
    set({ token })
  },

  signIn: async (publicKey, privateKey) => {
    const res = await fetch('/api/v1/auth', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ public_key: publicKey, private_key: privateKey }),
    })
    const body: (AuthResponse & { error?: { message?: string } }) | null = await res
      .json()
      .catch(() => null)
    if (!res.ok || !body?.token) {
      throw new Error(body?.error?.message || 'Login gagal, periksa API key kamu')
    }
    localStorage.setItem(TOKEN_KEY, body.token)
    set({ token: body.token })
  },

  signOut: () => {
    localStorage.removeItem(TOKEN_KEY)
    set({ token: null })
  },
}))