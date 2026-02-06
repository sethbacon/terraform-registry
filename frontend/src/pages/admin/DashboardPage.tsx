import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
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
  ViewModule,
  Extension,
  CloudUpload,
  People,
  Business,
  Download,
  GitHub,
  Key,
} from '@mui/icons-material';
import api from '../../services/api';

interface StatCard {
  title: string;
  value: number | string;
  icon: React.ReactNode;
  color: string;
  route: string;
}

const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const [stats, setStats] = useState<{
    totalModules: number;
    totalProviders: number;
    manualProviders: number;
    mirroredProviders: number;
    totalProviderVersions: number;
    manualProviderVersions: number;
    mirroredProviderVersions: number;
    totalUsers: number;
    totalOrganizations: number;
    totalDownloads: number;
    totalSCMProviders: number;
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

      // Try to use the new stats endpoint first
      try {
        const dashboardStats = await api.getDashboardStats();
        setStats({
          totalModules: dashboardStats.modules.total || 0,
          totalProviders: dashboardStats.providers.total || 0,
          manualProviders: dashboardStats.providers.manual || 0,
          mirroredProviders: dashboardStats.providers.mirrored || 0,
          totalProviderVersions: dashboardStats.providers.total_versions || 0,
          manualProviderVersions: dashboardStats.providers.manual_versions || 0,
          mirroredProviderVersions: dashboardStats.providers.mirrored_versions || 0,
          totalUsers: dashboardStats.users || 0,
          totalOrganizations: dashboardStats.organizations || 0,
          totalDownloads: dashboardStats.downloads || 0,
          totalSCMProviders: dashboardStats.scm_providers || 0,
        });
      } catch (statsError) {
        // Fallback to old method if new endpoint doesn't exist yet
        console.log('Using fallback stats method');
        const [modulesRes, providersRes, usersRes, orgsRes, scmProvidersRes] = await Promise.all([
          api.searchModules({ limit: 1 }).catch(() => ({ modules: [], meta: { total: 0 } })),
          api.searchProviders({ limit: 1 }).catch(() => ({ providers: [], meta: { total: 0 } })),
          api.searchUsers('', 1, 1).catch(() => ({ users: [], pagination: { total: 0 } })),
          api.listOrganizations().catch(() => []),
          api.listSCMProviders().catch(() => []),
        ]);

        const totalModuleDownloads = (modulesRes.modules || []).reduce(
          (sum: number, m: any) => sum + (m.download_count || 0),
          0
        );
        const totalProviderDownloads = (providersRes.providers || []).reduce(
          (sum: number, p: any) => sum + (p.download_count || 0),
          0
        );

        setStats({
          totalModules: modulesRes.meta?.total || 0,
          totalProviders: providersRes.meta?.total || 0,
          manualProviders: providersRes.meta?.total || 0,
          mirroredProviders: 0,
          totalProviderVersions: 0,
          manualProviderVersions: 0,
          mirroredProviderVersions: 0,
          totalUsers: usersRes.pagination?.total || 0,
          totalOrganizations: orgsRes.length || 0,
          totalDownloads: totalModuleDownloads + totalProviderDownloads,
          totalSCMProviders: Array.isArray(scmProvidersRes) ? scmProvidersRes.length : 0,
        });
      }
    } catch (err) {
      console.error('Failed to load dashboard stats:', err);
      // In dev mode, just show zeros instead of error
      setStats({
        totalModules: 0,
        totalProviders: 0,
        manualProviders: 0,
        mirroredProviders: 0,
        totalProviderVersions: 0,
        manualProviderVersions: 0,
        mirroredProviderVersions: 0,
        totalUsers: 0,
        totalOrganizations: 0,
        totalDownloads: 0,
        totalSCMProviders: 0,
      });
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

  if (error) {
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
      icon: <ViewModule sx={{ fontSize: 40 }} />,
      color: '#5C4EE5',
      route: '/modules',
    },
    {
      title: 'Total Providers',
      value: stats.totalProviders,
      icon: <Extension sx={{ fontSize: 40 }} />,
      color: '#00D9C0',
      route: '/providers',
    },
    {
      title: 'Total Users',
      value: stats.totalUsers,
      icon: <People sx={{ fontSize: 40 }} />,
      color: '#FF6B6B',
      route: '/admin/users',
    },
    {
      title: 'Organizations',
      value: stats.totalOrganizations,
      icon: <Business sx={{ fontSize: 40 }} />,
      color: '#4ECDC4',
      route: '/admin/organizations',
    },
    {
      title: 'Total Downloads',
      value: stats.totalDownloads,
      icon: <Download sx={{ fontSize: 40 }} />,
      color: '#FFB74D',
      route: '/modules',
    },
    {
      title: 'SCM Providers',
      value: stats.totalSCMProviders,
      icon: <GitHub sx={{ fontSize: 40 }} />,
      color: '#6E5494',
      route: '/admin/scm-providers',
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
              onClick={() => navigate(stat.route)}
              sx={{
                p: 3,
                display: 'flex',
                alignItems: 'center',
                gap: 2,
                cursor: 'pointer',
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
                {/* Show breakdown for providers */}
                {stat.title === 'Total Providers' && (
                  <Typography variant="caption" color="text.secondary" display="block" sx={{ mt: 0.5 }}>
                    {stats.manualProviders} manual, {stats.mirroredProviders} mirrored
                    {stats.totalProviderVersions > 0 && (
                      <><br />{stats.totalProviderVersions} versions ({stats.manualProviderVersions} manual, {stats.mirroredProviderVersions} mirrored)</>
                    )}
                  </Typography>
                )}
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
          <Paper 
            onClick={() => navigate('/admin/upload', { state: { tab: 0 } })}
            sx={{ 
              p: 3, 
              cursor: 'pointer',
              transition: 'transform 0.2s, box-shadow 0.2s',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow: 4,
              },
            }}
          >
            <CloudUpload sx={{ fontSize: 40, color: '#5C4EE5', mb: 2 }} />
            <Typography variant="h6" gutterBottom>
              Upload Module
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Upload a new Terraform module to your registry
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} md={6}>
          <Paper 
            onClick={() => navigate('/admin/upload', { state: { tab: 1 } })}
            sx={{ 
              p: 3, 
              cursor: 'pointer',
              transition: 'transform 0.2s, box-shadow 0.2s',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow: 4,
              },
            }}
          >
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
          <Paper 
            onClick={() => navigate('/admin/users')}
            sx={{ 
              p: 3, 
              cursor: 'pointer',
              transition: 'transform 0.2s, box-shadow 0.2s',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow: 4,
              },
            }}
          >
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
          <Paper 
            onClick={() => navigate('/admin/apikeys')}
            sx={{ 
              p: 3, 
              cursor: 'pointer',
              transition: 'transform 0.2s, box-shadow 0.2s',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow: 4,
              },
            }}
          >
            <Key sx={{ fontSize: 40, color: '#FFB74D', mb: 2 }} />
            <Typography variant="h6" gutterBottom>
              API Keys
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Generate and manage API keys for Terraform CLI
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12} md={6}>
          <Paper 
            onClick={() => navigate('/admin/scm-providers')}
            sx={{ 
              p: 3, 
              cursor: 'pointer',
              transition: 'transform 0.2s, box-shadow 0.2s',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow: 4,
              },
            }}
          >
            <GitHub sx={{ fontSize: 40, color: '#6E5494', mb: 2 }} />
            <Typography variant="h6" gutterBottom>
              SCM Providers
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Connect GitHub, Azure DevOps, or GitLab for automated publishing
            </Typography>
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
};

export default DashboardPage;
