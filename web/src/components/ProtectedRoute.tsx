import { Navigate, useLocation } from 'react-router'
import { useAuthStore } from '@/store/auth'

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token)
  const { pathname } = useLocation()

  if (!token) {
    return <Navigate to="/login" replace state={{ from: pathname }} />
  }

  return children
}