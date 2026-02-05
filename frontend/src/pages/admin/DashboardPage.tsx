import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Box,
  Grid,
  Paper,
  CircularProgress,
  Alert,
} from '@mui/material';
import {
  Extension,
  CloudUpload,
  People,
  Business,
  VpnKey,
  Download,
} from '@mui/icons-material';
import api from '../../services/api';

interface StatCard {
  title: string;
  value: number | string;
  icon: React.ReactNode;
  color: string;
}

const DashboardPage: React.FC = () => {
  const [stats, setStats] = useState<{
    totalModules: number;
    totalProviders: number;
    totalUsers: number;
    totalOrganizations: number;
    totalDownloads: number;
  } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadDashboardStats();
  }, []);

  const loadDashboardStats = async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch data from multiple endpoints with fallbacks for dev mode
      const [modulesRes, providersRes, usersRes, orgsRes] = await Promise.all([
        api.searchModules({ limit: 1 }).catch(() => ({ modules: [], meta: { total: 0 } })),
        api.searchProviders({ limit: 1 }).catch(() => ({ providers: [], meta: { total: 0 } })),
        api.searchUsers({ limit: 1 }).catch(() => ({ users: [], meta: { total: 0 } })),
        api.listOrganizations().catch(() => []),
      ]);

      // Calculate total downloads with safety checks
      const totalModuleDownloads = (modulesRes.modules || []).reduce(
        (sum, m) => sum + (m.download_count || 0),
        0
      );
      const totalProviderDownloads = (providersRes.providers || []).reduce(
        (sum, p) => sum + (p.download_count || 0),
        0
      );

      setStats({
        totalModules: modulesRes.meta?.total || 0,
        totalProviders: providersRes.meta?.total || 0,
        totalUsers: usersRes.meta?.total || 0,
        totalOrganizations: orgsRes.length || 0,
        totalDownloads: totalModuleDownloads + totalProviderDownloads,
      });
    } catch (err) {
      console.error('Failed to load dashboard stats:', err);
      // In dev mode, just show zeros instead of error
      setStats({
        totalModules: 0,
        totalProviders: 0,
        totalUsers: 0,
        totalOrganizations: 0,
        totalDownloads: 0,
      });
      if (!import.meta.env.DEV) {
        setError('Failed to load dashboard statistics. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error && !import.meta.env.DEV) {
    return (
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Alert severity="error">{error}</Alert>
      </Container>
    );
  }

  if (!stats) {
    return (
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Alert severity="error">Failed to load dashboard</Alert>
      </Container>
    );
  }

  const statCards: StatCard[] = [
    {
      title: 'Total Modules',
      value: stats.totalModules,
      icon: <Extension sx={{ fontSize: 40 }} />,
      color: '#5C4EE5',
    },
    {
      title: 'Total Providers',
      value: stats.totalProviders,
      icon: <CloudUpload sx={{ fontSize: 40 }} />,
      color: '#00D9C0',
    },
    {
      title: 'Total Users',
      value: stats.totalUsers,
      icon: <People sx={{ fontSize: 40 }} />,
      color: '#FF6B6B',
    },
    {
      title: 'Organizations',
      value: stats.totalOrganizations,
      icon: <Business sx={{ fontSize: 40 }} />,
      color: '#4ECDC4',
    },
    {
      title: 'Total Downloads',
      value: stats.totalDownloads,
      icon: <Download sx={{ fontSize: 40 }} />,
      color: '#FFB74D',
    },
  ];

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Typography variant="h4" gutterBottom>
        Admin Dashboard
      </Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
        Overview of your Terraform registry
      </Typography>

      {/* Statistics Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        {statCards.map((stat, index) => (
          <Grid item xs={12} sm={6} md={4} key={index}>
            <Paper
              sx={{
                p: 3,
                display: 'flex',
                alignItems: 'center',
                gap: 2,
                transition: 'transform 0.2s, box-shadow 0.2s',
                '&:hover': {
                  transform: 'translateY(-2px)',
                  boxShadow: 4,
                },
              }}
            >
              <Box sx={{ color: stat.color }}>{stat.icon}</Box>
              <Box sx={{ flex: 1 }}>
                <Typography variant="h4" component="div" fontWeight="bold">
                  {stat.value}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {stat.title}
                </Typography>
              </Box>
            </Paper>
          </Grid>
        ))}
      </Grid>

      {/* Quick Actions */}
      <Typography variant="h5" gutterBottom sx={{ mt: 4, mb: 2 }}>
        Quick Actions
      </Typography>
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3 }}>
            <Extension sx={{ fontSize: 40, color: '#5C4EE5', mb: 2 }} />
            <Typography variant="h6" gutterBottom>
              Upload Module
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Upload a new Terraform module to your registry
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3 }}>
            <CloudUpload sx={{ fontSize: 40, color: '#00D9C0', mb: 2 }} />
            <Typography variant="h6" gutterBottom>
              Upload Provider
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Upload a new Terraform provider to your registry
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3 }}>
            <People sx={{ fontSize: 40, color: '#FF6B6B', mb: 2 }} />
            <Typography variant="h6" gutterBottom>
              Manage Users
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Add, edit, or remove users and their permissions
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3 }}>
            <VpnKey sx={{ fontSize: 40, color: '#FFB74D', mb: 2 }} />
            <Typography variant="h6" gutterBottom>
              API Keys
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Generate and manage API keys for Terraform CLI
            </Typography>
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
};

export default DashboardPage;
