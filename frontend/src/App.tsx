import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme, CssBaseline } from '@mui/material';
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
import ProtectedRoute from './components/ProtectedRoute';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#5C4EE5',
    },
    secondary: {
      main: '#00D9C0',
    },
  },
  typography: {
    fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
  },
});

function App() {
  return (
    <ThemeProvider theme={theme}>
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
              
              {/* Admin routes (protected) */}
              <Route path="/admin" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
              <Route path="/admin/users" element={<ProtectedRoute><UsersPage /></ProtectedRoute>} />
              <Route path="/admin/organizations" element={<ProtectedRoute><OrganizationsPage /></ProtectedRoute>} />
              <Route path="/admin/apikeys" element={<ProtectedRoute><APIKeysPage /></ProtectedRoute>} />
              <Route path="/admin/upload" element={<ProtectedRoute><UploadPage /></ProtectedRoute>} />
              
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
