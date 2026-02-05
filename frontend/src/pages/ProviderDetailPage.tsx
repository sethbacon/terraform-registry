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
  Select,
  MenuItem,
  FormControl,
} from '@mui/material';
import {
  ArrowBack,
  ContentCopy,
} from '@mui/icons-material';
import api from '../services/api';
import { Provider, ProviderVersion } from '../types';

const ProviderDetailPage: React.FC = () => {
  const { namespace, type } = useParams<{
    namespace: string;
    type: string;
  }>();
  const navigate = useNavigate();
  
  // Use 'type' as the name for display
  const name = type;

  const [provider, setProvider] = useState<Provider | null>(null);
  const [versions, setVersions] = useState<ProviderVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<ProviderVersion | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copiedSource, setCopiedSource] = useState(false);
  const [copiedChecksum, setCopiedChecksum] = useState<string | null>(null);

  useEffect(() => {
    loadProviderDetails();
  }, [namespace, type]);

  const loadProviderDetails = async () => {
    if (!namespace || !type) return;

    try {
      setLoading(true);
      setError(null);

      // Use searchProviders with namespace filter and then find by type
      const [providerData, versionsData] = await Promise.all([
        api.searchProviders({ query: type, limit: 100 }), // Search with type as query
        api.getProviderVersions(namespace, type),
      ]);

      // Filter results to find exact match for namespace/type
      const matchingProvider = providerData.providers.find(
        (p: Provider) => p.namespace === namespace && p.type === type
      );

      if (!matchingProvider) {
        setError('Provider not found');
        return;
      }

      setProvider(matchingProvider);
      
      // Backend returns { versions: [...] } directly
      const versions = versionsData.versions || [];
      setVersions(versions);
      
      if (versions.length > 0) {
        setSelectedVersion(versions[0]);
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

  const handleCopyChecksum = (checksum: string) => {
    navigator.clipboard.writeText(checksum);
    setCopiedChecksum(checksum);
    setTimeout(() => setCopiedChecksum(null), 2000);
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
        {selectedVersion && (
          <Typography color="text.primary">v{selectedVersion.version}</Typography>
        )}
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
        <Stack direction="row" spacing={1} alignItems="center" sx={{ mt: 2 }}>
          <Chip label={namespace} />
          {provider.source && (
            <Chip 
              label="Network Mirrored" 
              color="info" 
              size="small" 
              variant="outlined"
            />
          )}
          <FormControl size="small" sx={{ minWidth: 150 }}>
            <Select
              value={selectedVersion?.version || ''}
              onChange={(e) => {
                const version = versions.find(v => v.version === e.target.value);
                if (version) setSelectedVersion(version);
              }}
              displayEmpty
            >
              {versions.map((v) => (
                <MenuItem key={v.id} value={v.version}>
                  v{v.version}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <Chip label={`${provider.download_count ?? 0} downloads`} />
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

          {/* Platforms Table */}
          {selectedVersion && selectedVersion.platforms && selectedVersion.platforms.length > 0 && (
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                Available Platforms
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>OS</TableCell>
                      <TableCell>Architecture</TableCell>
                      <TableCell>SHA256 Sum</TableCell>
                      <TableCell width="50px"></TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {selectedVersion.platforms.map((platform) => (
                      <TableRow key={platform.id}>
                        <TableCell>{platform.os}</TableCell>
                        <TableCell>{platform.arch}</TableCell>
                        <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.75rem', wordBreak: 'break-all' }}>
                          {platform.shasum || 'N/A'}
                        </TableCell>
                        <TableCell>
                          {platform.shasum && (
                            <Tooltip title={copiedChecksum === platform.shasum ? "Copied!" : "Copy checksum"}>
                              <IconButton
                                size="small"
                                onClick={() => handleCopyChecksum(platform.shasum)}
                              >
                                <ContentCopy fontSize="small" />
                              </IconButton>
                            </Tooltip>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </Paper>
          )}
        </Box>

        {/* Sidebar - Provider Information and Version Details */}
        <Box sx={{ width: { xs: '100%', md: 350 } }}>
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
                <strong>Latest Version:</strong> {versions.length > 0 ? versions[0].version : 'N/A'}
              </Typography>
              <Typography variant="body2">
                <strong>Total Downloads:</strong> {provider.download_count ?? 0}
              </Typography>
              <Typography variant="body2">
                <strong>Organization:</strong> {provider.organization_name || 'N/A'}
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
                {new Date(selectedVersion.published_at).toISOString().split('T')[0]}
              </Typography>
              <Typography variant="body2" sx={{ mb: 1 }}>
                <strong>Downloads:</strong> {selectedVersion.download_count ?? 0}
              </Typography>
            </Paper>
          )}
        </Box>
      </Box>
    </Container>
  );
};

export default ProviderDetailPage;
