import { BrowserRouter, Routes, Route } from 'react-router'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '@/lib/query'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { AppLayout } from '@/components/app/AppLayout'
import HomePage from '@/pages/HomePage'
import LoginPage from '@/pages/LoginPage'
import DashboardPage from '@/pages/app/DashboardPage'
import WilayahPage from '@/pages/app/WilayahPage'
import SearchPage from '@/pages/app/SearchPage'
import ReversePage from '@/pages/app/ReversePage'
import DistancePage from '@/pages/app/DistancePage'

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/app"
            element={
              <ProtectedRoute>
                <AppLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<NavigateToDashboard />} />
            <Route path="dashboard" element={<DashboardPage />} />
            <Route path="wilayah" element={<WilayahPage />} />
            <Route path="wilayah/:kode" element={<WilayahPage />} />
            <Route path="search" element={<SearchPage />} />
            <Route path="reverse" element={<ReversePage />} />
            <Route path="distance" element={<DistancePage />} />
          </Route>
          <Route path="*" element={<NavigateToDashboard />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

import { Navigate } from 'react-router'

function NavigateToDashboard() {
  return <Navigate to="/app/dashboard" replace />
}