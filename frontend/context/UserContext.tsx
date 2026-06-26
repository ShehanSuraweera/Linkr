"use client"

import { createContext, useContext, ReactNode } from "react"
import type { User } from "@/lib/types"

interface UserContextValue {
  user: User | null
}

const UserContext = createContext<UserContextValue>({ user: null })

interface UserProviderProps {
  children: ReactNode
  initialUser: User | null
}

export function UserProvider({ children, initialUser }: UserProviderProps) {
  return (
    <UserContext.Provider value={{ user: initialUser }}>
      {children}
    </UserContext.Provider>
  )
}

export function useUser() {
  return useContext(UserContext)
}
