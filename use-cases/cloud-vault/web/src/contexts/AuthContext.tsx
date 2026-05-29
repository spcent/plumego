import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { authAPI, User, LoginRequest, UpdateProfileRequest, ChangePasswordRequest, SetupRequest } from '../api/auth'

interface AuthContextType {
  user: User | null
  loading: boolean
  setupRequired: boolean
  login: (req: LoginRequest) => Promise<void>
  setup: (req: SetupRequest) => Promise<void>
  logout: () => Promise<void>
  updateProfile: (req: UpdateProfileRequest) => Promise<void>
  changePassword: (req: ChangePasswordRequest) => Promise<void>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [setupRequired, setSetupRequired] = useState(false)

  useEffect(() => {
    checkSetupAndRefresh()
  }, [])

  const checkSetupAndRefresh = async () => {
    try {
      const status = await authAPI.getStatus()
      setSetupRequired(!status.initialized)

      if (status.initialized) {
        const currentUser = await authAPI.getMe()
        setUser(currentUser)
      }
    } catch {
      setUser(null)
    } finally {
      setLoading(false)
    }
  }

  const setup = async (req: SetupRequest) => {
    const response = await authAPI.setup(req)
    setUser(response.user)
    setSetupRequired(false)
  }

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
        setupRequired,
        login,
        setup,
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
