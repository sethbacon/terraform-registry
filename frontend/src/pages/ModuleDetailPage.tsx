import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
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
  Select,
  MenuItem,
  FormControl,
} from '@mui/material';
import {
  ArrowBack,
  Download,
  ContentCopy,
  Add,
} from '@mui/icons-material';
import api from '../services/api';
import { Module, ModuleVersion } from '../types';
import { useAuth } from '../contexts/AuthContext';

const ModuleDetailPage: React.FC = () => {
  const { namespace, name, system } = useParams<{
    namespace: string;
    name: string;
    system: string;
  }>();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();

  const [module, setModule] = useState<Module | null>(null);
  const [versions, setVersions] = useState<ModuleVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<ModuleVersion | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copiedSource, setCopiedSource] = useState(false);

  useEffect(() => {
    loadModuleDetails();
  }, [namespace, name, system]);

  const loadModuleDetails = async () => {
    if (!namespace || !name || !system) return;

    try {
      setLoading(true);
      setError(null);

      const [moduleData, versionsData] = await Promise.all([
        api.searchModules({ namespace, name, provider: system, limit: 1 }),
        api.getModuleVersions(namespace, name, system),
      ]);

      if (moduleData.modules.length === 0) {
        setError('Module not found');
        return;
      }

      setModule(moduleData.modules[0]);
      
      // Backend returns { modules: [{ versions: [...] }] }
      const versions = versionsData.modules?.[0]?.versions || [];
      setVersions(versions);
      
      // Select latest version by default
      if (versions.length > 0) {
        setSelectedVersion(versions[0]);
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
    
    const source = `${namespace}/${name}/${system}`;
    navigator.clipboard.writeText(source);
    setCopiedSource(true);
    setTimeout(() => setCopiedSource(false), 2000);
  };

  const handlePublishNewVersion = () => {
    navigate('/admin/upload', {
      state: {
        tab: 0,
        moduleData: {
          namespace,
          name,
          provider: system,
        },
      },
    });
  };

  const getTerraformExample = () => {
    if (!module || !selectedVersion) return '';

    return `module "${name}" {
  source  = "${window.location.origin}/v1/modules/${namespace}/${name}/${system}/${selectedVersion.version}"
  
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
        <Typography color="text.primary">{system}</Typography>
        {selectedVersion && (
          <Typography color="text.primary">v{selectedVersion.version}</Typography>
        )}
      </Breadcrumbs>

      {/* Header */}
      <Box sx={{ mb: 4 }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
          <Stack direction="row" alignItems="center" spacing={2}>
            <IconButton onClick={() => navigate('/modules')}>
              <ArrowBack />
            </IconButton>
            <Typography variant="h4" component="h1">
            {name}
          </Typography>
          </Stack>
          {isAuthenticated && (
            <Button
              variant="contained"
              startIcon={<Add />}
              onClick={handlePublishNewVersion}
            >
              Publish New Version
            </Button>
          )}
        </Stack>
        <Typography variant="body1" color="text.secondary" gutterBottom>
          {module.description || 'No description available'}
        </Typography>
        <Stack direction="row" spacing={1} alignItems="center" sx={{ mt: 2 }}>
          <Chip label={`${namespace}/${system}`} />
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

          {/* README */}
          {selectedVersion && selectedVersion.readme && (
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                README
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <Box
                sx={{
                  '& h1': { fontSize: '2rem', fontWeight: 600, mt: 2, mb: 1 },
                  '& h2': { fontSize: '1.5rem', fontWeight: 600, mt: 2, mb: 1 },
                  '& h3': { fontSize: '1.25rem', fontWeight: 600, mt: 2, mb: 1 },
                  '& p': { mb: 2 },
                  '& code': {
                    backgroundColor: '#f5f5f5',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    fontFamily: 'monospace',
                    fontSize: '0.875rem',
                  },
                  '& pre': {
                    backgroundColor: '#f5f5f5',
                    padding: 2,
                    borderRadius: 1,
                    overflow: 'auto',
                  },
                  '& pre code': {
                    backgroundColor: 'transparent',
                    padding: 0,
                  },
                  '& ul, & ol': { pl: 3, mb: 2 },
                  '& li': { mb: 1 },
                  '& table': { borderCollapse: 'collapse', width: '100%', mb: 2 },
                  '& th, & td': {
                    border: '1px solid #ddd',
                    padding: '8px 12px',
                    textAlign: 'left',
                  },
                  '& th': { backgroundColor: '#f5f5f5', fontWeight: 600 },
                }}
              >
                <ReactMarkdown remarkPlugins={[remarkGfm]}>{selectedVersion.readme}</ReactMarkdown>
              </Box>
            </Paper>
          )}
        </Box>

        {/* Sidebar - Module Information and Version Details */}
        <Box sx={{ width: { xs: '100%', md: 350 } }}>
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
                <strong>Provider:</strong> {system}
              </Typography>
              <Typography variant="body2">
                <strong>Latest Version:</strong> {versions.length > 0 ? versions[0].version : 'N/A'}
              </Typography>
              <Typography variant="body2">
                <strong>Total Downloads:</strong> {module.download_count ?? 0}
              </Typography>
              <Typography variant="body2">
                <strong>Organization:</strong> {module.organization_name || 'N/A'}
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
                {new Date(selectedVersion.published_at).toISOString().split('T')[0]}
              </Typography>
              <Typography variant="body2" sx={{ mb: 1 }}>
                <strong>Downloads:</strong> {selectedVersion.download_count ?? 0}
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
      </Box>
    </Container>
  );
};

export default ModuleDetailPage;
