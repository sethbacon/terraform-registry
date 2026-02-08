import React from 'react';
import {
  Container,
  Paper,
  Typography,
  Button,
  Box,
  Stack,
  Divider,
  Alert,
} from '@mui/material';
import { Login as LoginIcon } from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';

const LoginPage: React.FC = () => {
  const { login } = useAuth();
  const navigate = useNavigate();
  const isDev = import.meta.env.DEV;

  const handleDevLogin = async () => {
    // SECURITY: Store the development API key - user data will be fetched from API
    // This ensures scopes and permissions always come from the server
    localStorage.setItem('auth_token', 'dev_qHlTX4JvjK1yVUgRukLlgiwFQmFOiHdEhHYVJNfhNXc');

    // Clear any cached user data to force fresh fetch from API
    localStorage.removeItem('user');
    localStorage.removeItem('role_template');
    localStorage.removeItem('allowed_scopes');

    console.log('Dev login: setting API key, will fetch user from API');
    // Pass a placeholder - login() will fetch actual user data from API
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await login({} as any);
    console.log('Dev login: user data fetched, navigating to home');
    navigate('/');
  };

  const handleOIDCLogin = () => {
    window.location.href = '/api/auth/login/oidc';
  };

  const handleAzureADLogin = () => {
    window.location.href = '/api/auth/login/azuread';
  };

  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Paper elevation={3} sx={{ p: 4, width: '100%' }}>
          <Box sx={{ textAlign: 'center', mb: 4 }}>
            <LoginIcon sx={{ fontSize: 60, color: 'primary.main', mb: 2 }} />
            <Typography variant="h4" component="h1" gutterBottom>
              Terraform Registry
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Sign in to continue
            </Typography>
          </Box>

          <Stack spacing={2}>
            {isDev && (
              <>
                <Alert severity="info">
                  Development mode - Click below to login without OAuth
                </Alert>
                <Button
                  variant="contained"
                  size="large"
                  fullWidth
                  onClick={handleDevLogin}
                  color="success"
                  sx={{ py: 1.5 }}
                >
                  Dev Login (Admin)
                </Button>
                <Divider>
                  <Typography variant="body2" color="text.secondary">
                    OR USE PRODUCTION AUTH
                  </Typography>
                </Divider>
              </>
            )}

            <Button
              variant="contained"
              size="large"
              fullWidth
              onClick={handleOIDCLogin}
              sx={{ py: 1.5 }}
            >
              Sign in with OIDC
            </Button>

            <Divider>
              <Typography variant="body2" color="text.secondary">
                OR
              </Typography>
            </Divider>

            <Button
              variant="contained"
              size="large"
              fullWidth
              onClick={handleAzureADLogin}
              sx={{ 
                py: 1.5,
                backgroundColor: '#0078d4',
                '&:hover': {
                  backgroundColor: '#106ebe',
                },
              }}
            >
              Sign in with Azure AD
            </Button>
          </Stack>

          <Box sx={{ mt: 3, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              This application uses single sign-on for authentication.
              <br />
              Contact your administrator if you need access.
            </Typography>
          </Box>
        </Paper>
      </Box>
    </Container>
  );
};

export default LoginPage;
