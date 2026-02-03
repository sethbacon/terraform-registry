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
} from '@mui/material';
import {
  ArrowBack,
  Download,
  ContentCopy,
} from '@mui/icons-material';
import api from '../services/api';
import { Module, ModuleVersion } from '../types';

const ModuleDetailPage: React.FC = () => {
  const { namespace, name, provider } = useParams<{
    namespace: string;
    name: string;
    provider: string;
  }>();
  const navigate = useNavigate();

  const [module, setModule] = useState<Module | null>(null);
  const [versions, setVersions] = useState<ModuleVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<ModuleVersion | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copiedSource, setCopiedSource] = useState(false);

  useEffect(() => {
    loadModuleDetails();
  }, [namespace, name, provider]);

  const loadModuleDetails = async () => {
    if (!namespace || !name || !provider) return;

    try {
      setLoading(true);
      setError(null);

      const [moduleData, versionsData] = await Promise.all([
        api.searchModules({ namespace, name, provider, limit: 1 }),
        api.getModuleVersions(namespace, name, provider),
      ]);

      if (moduleData.modules.length === 0) {
        setError('Module not found');
        return;
      }

      setModule(moduleData.modules[0]);
      setVersions(versionsData.versions);
      
      // Select latest version by default
      if (versionsData.versions.length > 0) {
        setSelectedVersion(versionsData.versions[0]);
      }
    } catch (err) {
      console.error('Failed to load module details:', err);
      setError('Failed to load module details. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleCopySource = () => {
    if (!module || !selectedVersion) return;
    
    const source = `${namespace}/${name}/${provider}`;
    navigator.clipboard.writeText(source);
    setCopiedSource(true);
    setTimeout(() => setCopiedSource(false), 2000);
  };

  const getTerraformExample = () => {
    if (!module || !selectedVersion) return '';

    return `module "${name}" {
  source  = "${window.location.origin}/v1/modules/${namespace}/${name}/${provider}/${selectedVersion.version}"
  
  # Add your module variables here
}`;
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !module) {
    return (
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Alert severity="error">{error || 'Module not found'}</Alert>
        <Button
          startIcon={<ArrowBack />}
          onClick={() => navigate('/modules')}
          sx={{ mt: 2 }}
        >
          Back to Modules
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
          onClick={() => navigate('/modules')}
          sx={{ cursor: 'pointer' }}
        >
          Modules
        </Link>
        <Typography color="text.primary">{namespace}</Typography>
        <Typography color="text.primary">{name}</Typography>
        <Typography color="text.primary">{provider}</Typography>
      </Breadcrumbs>

      {/* Header */}
      <Box sx={{ mb: 4 }}>
        <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
          <IconButton onClick={() => navigate('/modules')}>
            <ArrowBack />
          </IconButton>
          <Typography variant="h4" component="h1">
            {name}
          </Typography>
        </Stack>
        <Typography variant="body1" color="text.secondary" gutterBottom>
          {module.description || 'No description available'}
        </Typography>
        <Stack direction="row" spacing={1} sx={{ mt: 2 }}>
          <Chip label={`${namespace}/${provider}`} />
          <Chip label={`Latest: ${module.latest_version}`} color="primary" />
          <Chip label={`${module.download_count} downloads`} />
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

          {/* Module Information */}
          <Paper sx={{ p: 3, mb: 3 }}>
            <Typography variant="h6" gutterBottom>
              Module Information
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
                <strong>Provider:</strong> {provider}
              </Typography>
              <Typography variant="body2">
                <strong>Latest Version:</strong> {module.latest_version}
              </Typography>
              <Typography variant="body2">
                <strong>Total Downloads:</strong> {module.download_count}
              </Typography>
              <Typography variant="body2">
                <strong>Organization:</strong> {module.organization_name}
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
              <Typography variant="body2" sx={{ mb: 1 }}>
                <strong>Published:</strong>{' '}
                {new Date(selectedVersion.published_at).toLocaleDateString()}
              </Typography>
              <Typography variant="body2" sx={{ mb: 1 }}>
                <strong>Downloads:</strong> {selectedVersion.download_count}
              </Typography>
              {selectedVersion.source_url && (
                <Typography variant="body2">
                  <strong>Source:</strong>{' '}
                  <Link href={selectedVersion.source_url} target="_blank" rel="noopener">
                    {selectedVersion.source_url}
                  </Link>
                </Typography>
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

export default ModuleDetailPage;
