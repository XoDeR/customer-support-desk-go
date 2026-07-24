import { create } from "zustand";
import type { User } from "./types";

const TOKEN_KEY = "support-desk.tokens";

export interface Tokens {
  access_token: string;
  refresh_token: string;
}

const savedTokens = (): Tokens | null => {
  try {
    return JSON.parse(localStorage.getItem(TOKEN_KEY) ?? "null");
  } catch {
    return null;
  }
};

interface AuthState {
  tokens: Tokens | null;
  user: User | null;
  setTokens: (tokens: Tokens | null) => void;
  setUser: (user: User | null) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  tokens: savedTokens(),
  user: null,
  setTokens: (tokens) => {
    if (tokens) localStorage.setItem(TOKEN_KEY, JSON.stringify(tokens));
    else localStorage.removeItem(TOKEN_KEY);
    set({ tokens });
  },
  setUser: (user) => set({ user }),
  clear: () => {
    localStorage.removeItem(TOKEN_KEY);
    set({ tokens: null, user: null });
  },
}));
