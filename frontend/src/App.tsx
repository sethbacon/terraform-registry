import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { CssBaseline } from '@mui/material';
import { ThemeProvider } from './contexts/ThemeContext';
import { AuthProvider } from './contexts/AuthContext';
import Layout from './components/Layout';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import CallbackPage from './pages/CallbackPage';
import ModulesPage from './pages/ModulesPage';
import ModuleDetailPage from './pages/ModuleDetailPage';
import ProvidersPage from './pages/ProvidersPage';
import ProviderDetailPage from './pages/ProviderDetailPage';
import DashboardPage from './pages/admin/DashboardPage';
import UsersPage from './pages/admin/UsersPage';
import OrganizationsPage from './pages/admin/OrganizationsPage';
import APIKeysPage from './pages/admin/APIKeysPage';
import UploadPage from './pages/admin/UploadPage';
import SCMProvidersPage from './pages/admin/SCMProvidersPage';
import MirrorsPage from './pages/admin/MirrorsPage';
import RolesPage from './pages/admin/RolesPage';
import StoragePage from './pages/admin/StoragePage';
import ProtectedRoute from './components/ProtectedRoute';

function App() {
  return (
    <ThemeProvider>
      <CssBaseline />
      <AuthProvider>
        <Router>
          <Routes>
            {/* Public routes */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/auth/callback" element={<CallbackPage />} />

            {/* Layout routes */}
            <Route element={<Layout />}>
              <Route path="/" element={<HomePage />} />

              {/* Modules */}
              <Route path="/modules" element={<ModulesPage />} />
              <Route path="/modules/:namespace/:name/:system" element={<ModuleDetailPage />} />

              {/* Providers */}
              <Route path="/providers" element={<ProvidersPage />} />
              <Route path="/providers/:namespace/:type" element={<ProviderDetailPage />} />

              {/* Admin routes (protected with scope requirements) */}
              <Route path="/admin" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
              <Route path="/admin/users" element={<ProtectedRoute requiredScope="users:read"><UsersPage /></ProtectedRoute>} />
              <Route path="/admin/organizations" element={<ProtectedRoute requiredScope="organizations:read"><OrganizationsPage /></ProtectedRoute>} />
              <Route path="/admin/roles" element={<ProtectedRoute requiredScope="users:read"><RolesPage /></ProtectedRoute>} />
              <Route path="/admin/apikeys" element={<ProtectedRoute><APIKeysPage /></ProtectedRoute>} />
              <Route path="/admin/upload" element={<ProtectedRoute requiredScope="modules:write"><UploadPage /></ProtectedRoute>} />
              <Route path="/admin/scm-providers" element={<ProtectedRoute requiredScope="scm:read"><SCMProvidersPage /></ProtectedRoute>} />
              <Route path="/admin/mirrors" element={<ProtectedRoute requiredScope="mirrors:read"><MirrorsPage /></ProtectedRoute>} />
              <Route path="/admin/storage" element={<ProtectedRoute requiredScope="admin"><StoragePage /></ProtectedRoute>} />

              {/* Catch all */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
          </Routes>
        </Router>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
