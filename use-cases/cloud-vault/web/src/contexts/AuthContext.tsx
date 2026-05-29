import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { authAPI, User, LoginRequest, UpdateProfileRequest, ChangePasswordRequest } from '../api/auth'

interface AuthContextType {
  user: User | null
  loading: boolean
  login: (req: LoginRequest) => Promise<void>
  logout: () => Promise<void>
  updateProfile: (req: UpdateProfileRequest) => Promise<void>
  changePassword: (req: ChangePasswordRequest) => Promise<void>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    refreshUser()
  }, [])

  const refreshUser = async () => {
    try {
      const currentUser = await authAPI.getMe()
      setUser(currentUser)
    } catch {
      setUser(null)
    } finally {
      setLoading(false)
    }
  }

  const login = async (req: LoginRequest) => {
    const response = await authAPI.login(req)
    setUser(response.user)
  }

  const logout = async () => {
    try {
      await authAPI.logout()
    } finally {
      setUser(null)
    }
  }

  const updateProfile = async (req: UpdateProfileRequest) => {
    const updatedUser = await authAPI.updateProfile(req)
    setUser(updatedUser)
  }

  const changePassword = async (req: ChangePasswordRequest) => {
    await authAPI.changePassword(req)
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        login,
        logout,
        updateProfile,
        changePassword,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
