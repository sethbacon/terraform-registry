import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Container,
  Typography,
  Box,
  Paper,
  Breadcrumbs,
  Link,
  Chip,
  Divider,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  CircularProgress,
  Alert,
  Button,
  Stack,
  IconButton,
  Tooltip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';
import {
  ArrowBack,
  ContentCopy,
} from '@mui/icons-material';
import api from '../services/api';
import { Provider, ProviderVersion } from '../types';

const ProviderDetailPage: React.FC = () => {
  const { namespace, name } = useParams<{
    namespace: string;
    name: string;
  }>();
  const navigate = useNavigate();

  const [provider, setProvider] = useState<Provider | null>(null);
  const [versions, setVersions] = useState<ProviderVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<ProviderVersion | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copiedSource, setCopiedSource] = useState(false);

  useEffect(() => {
    loadProviderDetails();
  }, [namespace, name]);

  const loadProviderDetails = async () => {
    if (!namespace || !name) return;

    try {
      setLoading(true);
      setError(null);

      const [providerData, versionsData] = await Promise.all([
        api.searchProviders({ namespace, name, limit: 1 }),
        api.getProviderVersions(namespace, name),
      ]);

      if (providerData.providers.length === 0) {
        setError('Provider not found');
        return;
      }

      setProvider(providerData.providers[0]);
      setVersions(versionsData.versions);
      
      if (versionsData.versions.length > 0) {
        setSelectedVersion(versionsData.versions[0]);
      }
    } catch (err) {
      console.error('Failed to load provider details:', err);
      setError('Failed to load provider details. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleCopySource = () => {
    if (!provider || !selectedVersion) return;
    
    const source = `${namespace}/${name}`;
    navigator.clipboard.writeText(source);
    setCopiedSource(true);
    setTimeout(() => setCopiedSource(false), 2000);
  };

  const getTerraformExample = () => {
    if (!provider || !selectedVersion) return '';

    return `terraform {
  required_providers {
    ${name} = {
      source  = "${window.location.origin}/v1/providers/${namespace}/${name}"
      version = "${selectedVersion.version}"
    }
  }
}

provider "${name}" {
  # Configure provider settings here
}`;
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !provider) {
    return (
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Alert severity="error">{error || 'Provider not found'}</Alert>
        <Button
          startIcon={<ArrowBack />}
          onClick={() => navigate('/providers')}
          sx={{ mt: 2 }}
        >
          Back to Providers
        </Button>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      {/* Breadcrumbs */}
      <Breadcrumbs sx={{ mb: 3 }}>
        <Link
          component="button"
          variant="body1"
          onClick={() => navigate('/providers')}
          sx={{ cursor: 'pointer' }}
        >
          Providers
        </Link>
        <Typography color="text.primary">{namespace}</Typography>
        <Typography color="text.primary">{name}</Typography>
      </Breadcrumbs>

      {/* Header */}
      <Box sx={{ mb: 4 }}>
        <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
          <IconButton onClick={() => navigate('/providers')}>
            <ArrowBack />
          </IconButton>
          <Typography variant="h4" component="h1">
            {name}
          </Typography>
        </Stack>
        <Typography variant="body1" color="text.secondary" gutterBottom>
          {provider.description || 'No description available'}
        </Typography>
        <Stack direction="row" spacing={1} sx={{ mt: 2 }}>
          <Chip label={namespace} />
          <Chip label={`Latest: ${provider.latest_version}`} color="secondary" />
          <Chip label={`${provider.download_count} downloads`} />
        </Stack>
      </Box>

      <Box sx={{ display: 'flex', gap: 3, flexDirection: { xs: 'column', md: 'row' } }}>
        {/* Main Content */}
        <Box sx={{ flex: 1 }}>
          {/* Usage Example */}
          <Paper sx={{ p: 3, mb: 3 }}>
            <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
              <Typography variant="h6">Usage Example</Typography>
              <Tooltip title={copiedSource ? 'Copied!' : 'Copy source'}>
                <IconButton onClick={handleCopySource} size="small">
                  <ContentCopy />
                </IconButton>
              </Tooltip>
            </Stack>
            <Box
              component="pre"
              sx={{
                p: 2,
                backgroundColor: '#f5f5f5',
                borderRadius: 1,
                overflow: 'auto',
                fontSize: '0.875rem',
              }}
            >
              <code>{getTerraformExample()}</code>
            </Box>
          </Paper>

          {/* Provider Information */}
          <Paper sx={{ p: 3, mb: 3 }}>
            <Typography variant="h6" gutterBottom>
              Provider Information
            </Typography>
            <Divider sx={{ mb: 2 }} />
            <Box sx={{ '& > *': { mb: 1 } }}>
              <Typography variant="body2">
                <strong>Namespace:</strong> {namespace}
              </Typography>
              <Typography variant="body2">
                <strong>Name:</strong> {name}
              </Typography>
              <Typography variant="body2">
                <strong>Latest Version:</strong> {provider.latest_version}
              </Typography>
              <Typography variant="body2">
                <strong>Total Downloads:</strong> {provider.download_count}
              </Typography>
              <Typography variant="body2">
                <strong>Organization:</strong> {provider.organization_name}
              </Typography>
            </Box>
          </Paper>

          {/* Selected Version Details */}
          {selectedVersion && (
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                Version {selectedVersion.version} Details
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <Typography variant="body2" sx={{ mb: 2 }}>
                <strong>Published:</strong>{' '}
                {new Date(selectedVersion.published_at).toLocaleDateString()}
              </Typography>
              <Typography variant="body2" sx={{ mb: 2 }}>
                <strong>Downloads:</strong> {selectedVersion.download_count}
              </Typography>

              {/* Platforms Table */}
              {selectedVersion.platforms && selectedVersion.platforms.length > 0 && (
                <>
                  <Typography variant="subtitle1" sx={{ mt: 3, mb: 2 }}>
                    Available Platforms
                  </Typography>
                  <TableContainer>
                    <Table size="small">
                      <TableHead>
                        <TableRow>
                          <TableCell>OS</TableCell>
                          <TableCell>Architecture</TableCell>
                          <TableCell>SHA256 Sum</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {selectedVersion.platforms.map((platform) => (
                          <TableRow key={platform.id}>
                            <TableCell>{platform.os}</TableCell>
                            <TableCell>{platform.arch}</TableCell>
                            <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
                              {platform.shasum?.substring(0, 16)}...
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </TableContainer>
                </>
              )}
            </Paper>
          )}
        </Box>

        {/* Sidebar - Versions List */}
        <Box sx={{ width: { xs: '100%', md: 300 } }}>
          <Paper>
            <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
              <Typography variant="h6">Versions</Typography>
            </Box>
            <List sx={{ maxHeight: 600, overflow: 'auto' }}>
              {versions.map((version) => (
                <ListItem key={version.id} disablePadding>
                  <ListItemButton
                    selected={selectedVersion?.id === version.id}
                    onClick={() => setSelectedVersion(version)}
                  >
                    <ListItemText
                      primary={version.version}
                      secondary={new Date(version.published_at).toLocaleDateString()}
                    />
                  </ListItemButton>
                </ListItem>
              ))}
            </List>
          </Paper>
        </Box>
      </Box>
    </Container>
  );
};

export default ProviderDetailPage;
