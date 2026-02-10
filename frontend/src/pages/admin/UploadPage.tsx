import React, { useState } from 'react';
import { useLocation } from 'react-router-dom';
import {
  Container,
  Typography,
  Box,
  Paper,
  Tabs,
  Tab,
  TextField,
  Button,
  Alert,
  Stack,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  CircularProgress,
  SelectChangeEvent,
  Card,
  CardActionArea,
  CardContent,
} from '@mui/material';
import {
  CloudUpload,
  AccountTree as SCMIcon,
  ArrowBack,
} from '@mui/icons-material';
import api from '../../services/api';
import PublishFromSCMWizard from '../../components/PublishFromSCMWizard';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

const TabPanel: React.FC<TabPanelProps> = ({ children, value, index }) => {
  return (
    <div role="tabpanel" hidden={value !== index}>
      {value === index && <Box sx={{ pt: 3 }}>{children}</Box>}
    </div>
  );
};

type ModuleMethod = 'choose' | 'upload' | 'scm';

const UploadPage: React.FC = () => {
  const location = useLocation();
  const state = location.state as {
    tab?: number;
    moduleData?: { namespace: string; name: string; provider: string };
    method?: ModuleMethod;
  };
  const initialTab = state?.tab ?? 0;
  const prefilledModule = state?.moduleData;

  const [tabValue, setTabValue] = useState(initialTab);
  const [moduleMethod, setModuleMethod] = useState<ModuleMethod>(state?.method ?? 'choose');

  // SCM new-module metadata (before wizard)
  const [scmNamespace, setScmNamespace] = useState(prefilledModule?.namespace || '');
  const [scmName, setScmName] = useState(prefilledModule?.name || '');
  const [scmSystem, setScmSystem] = useState(prefilledModule?.provider || '');
  const [scmDescription, setScmDescription] = useState('');
  const [scmModuleId, setScmModuleId] = useState<string | null>(null);
  const [scmCreating, setScmCreating] = useState(false);
  const [scmError, setScmError] = useState<string | null>(null);
  const [scmSuccess, setScmSuccess] = useState<string | null>(null);

  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Module upload state
  const [moduleFile, setModuleFile] = useState<File | null>(null);
  const [moduleNamespace, setModuleNamespace] = useState(prefilledModule?.namespace || '');
  const [moduleName, setModuleName] = useState(prefilledModule?.name || '');
  const [moduleProvider, setModuleProvider] = useState(prefilledModule?.provider || '');
  const [moduleVersion, setModuleVersion] = useState('');
  const [moduleDescription, setModuleDescription] = useState('');

  // Provider upload state
  const [providerFile, setProviderFile] = useState<File | null>(null);
  const [providerNamespace, setProviderNamespace] = useState('');
  const [providerName, setProviderName] = useState('');
  const [providerVersion, setProviderVersion] = useState('');
  const [providerOS, setProviderOS] = useState('');
  const [providerArch, setProviderArch] = useState('');

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
    setError(null);
    setSuccess(null);
    setModuleMethod('choose');
    setScmModuleId(null);
  };

  const handleModuleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setModuleFile(file);
      setError(null);
    }
  };

  const handleProviderFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setProviderFile(file);
      setError(null);
    }
  };

  const handleModuleUpload = async () => {
    if (!moduleFile || !moduleNamespace || !moduleName || !moduleProvider || !moduleVersion) {
      setError('Please fill in all required fields');
      return;
    }

    try {
      setUploading(true);
      setError(null);
      setSuccess(null);

      const formData = new FormData();
      formData.append('namespace', moduleNamespace);
      formData.append('name', moduleName);
      formData.append('system', moduleProvider);
      formData.append('version', moduleVersion);
      if (moduleDescription) formData.append('description', moduleDescription);
      formData.append('file', moduleFile);

      await api.uploadModule(formData);

      setSuccess(`Module ${moduleNamespace}/${moduleName}/${moduleProvider} v${moduleVersion} uploaded successfully!`);
      setModuleFile(null);
      setModuleNamespace('');
      setModuleName('');
      setModuleProvider('');
      setModuleVersion('');
      const fileInput = document.getElementById('module-file-input') as HTMLInputElement;
      if (fileInput) fileInput.value = '';
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to upload module. Please try again.');
    } finally {
      setUploading(false);
    }
  };

  const handleProviderUpload = async () => {
    if (!providerFile || !providerNamespace || !providerName || !providerVersion || !providerOS || !providerArch) {
      setError('Please fill in all required fields');
      return;
    }

    try {
      setUploading(true);
      setError(null);
      setSuccess(null);

      const formData = new FormData();
      formData.append('namespace', providerNamespace);
      formData.append('type', providerName);
      formData.append('version', providerVersion);
      formData.append('os', providerOS);
      formData.append('arch', providerArch);
      formData.append('file', providerFile);

      await api.uploadProvider(formData);

      setSuccess(`Provider ${providerNamespace}/${providerName} v${providerVersion} (${providerOS}/${providerArch}) uploaded successfully!`);
      setProviderFile(null);
      setProviderNamespace('');
      setProviderName('');
      setProviderVersion('');
      setProviderOS('');
      setProviderArch('');
      const fileInput = document.getElementById('provider-file-input') as HTMLInputElement;
      if (fileInput) fileInput.value = '';
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to upload provider. Please try again.');
    } finally {
      setUploading(false);
    }
  };

  const handleScmProceed = async () => {
    if (!scmNamespace || !scmName || !scmSystem) {
      setScmError('Namespace, name, and provider are required');
      return;
    }
    try {
      setScmCreating(true);
      setScmError(null);
      const module = await api.createModuleRecord({
        namespace: scmNamespace,
        name: scmName,
        system: scmSystem,
        description: scmDescription || undefined,
      });
      setScmModuleId(module.id);
    } catch (err: any) {
      setScmError(err.response?.data?.error || 'Failed to create module record');
    } finally {
      setScmCreating(false);
    }
  };

  const renderModuleMethodChooser = () => (
    <Box sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        How would you like to publish this module?
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
        Upload a packaged archive directly, or connect a git repository for automated publishing via webhooks.
      </Typography>
      <Box sx={{ display: 'flex', gap: 3, flexDirection: { xs: 'column', sm: 'row' } }}>
        <Card
          variant="outlined"
          sx={{ flex: 1, '&:hover': { borderColor: 'primary.main', boxShadow: 2 } }}
        >
          <CardActionArea sx={{ height: '100%' }} onClick={() => setModuleMethod('upload')}>
            <CardContent sx={{ textAlign: 'center', py: 4 }}>
              <CloudUpload sx={{ fontSize: 48, color: 'primary.main', mb: 2 }} />
              <Typography variant="h6" gutterBottom>
                Upload from File
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Package your module as a <strong>.tar.gz</strong> archive and upload it directly. Best for one-off or manual releases.
              </Typography>
            </CardContent>
          </CardActionArea>
        </Card>

        <Card
          variant="outlined"
          sx={{ flex: 1, '&:hover': { borderColor: 'primary.main', boxShadow: 2 } }}
        >
          <CardActionArea sx={{ height: '100%' }} onClick={() => setModuleMethod('scm')}>
            <CardContent sx={{ textAlign: 'center', py: 4 }}>
              <SCMIcon sx={{ fontSize: 48, color: 'secondary.main', mb: 2 }} />
              <Typography variant="h6" gutterBottom>
                Link from SCM Repository
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Connect a GitHub, Azure DevOps, GitLab, or Bitbucket repository. New versions publish automatically when tags are pushed.
              </Typography>
            </CardContent>
          </CardActionArea>
        </Card>
      </Box>
    </Box>
  );

  const renderScmMetadataForm = () => (
    <Box sx={{ p: 3 }}>
      <Button
        startIcon={<ArrowBack />}
        onClick={() => { setModuleMethod('choose'); setScmError(null); setScmModuleId(null); setScmSuccess(null); }}
        sx={{ mb: 2 }}
      >
        Back
      </Button>
      <Typography variant="h6" gutterBottom>
        Link Module to SCM Repository
      </Typography>

      {scmModuleId ? (
        <>
          <PublishFromSCMWizard
            moduleId={scmModuleId}
            onComplete={() => {
              setScmSuccess(`Module ${scmNamespace}/${scmName}/${scmSystem} linked to SCM repository! New versions will publish automatically when matching tags are pushed.`);
              setModuleMethod('choose');
              setScmModuleId(null);
              setScmNamespace('');
              setScmName('');
              setScmSystem('');
              setScmDescription('');
            }}
            onCancel={() => {
              setModuleMethod('choose');
              setScmModuleId(null);
            }}
          />
          {scmSuccess && <Alert severity="success" sx={{ mt: 2 }}>{scmSuccess}</Alert>}
        </>
      ) : (
        <>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            First, define the module identity. Then you'll choose a repository and configure publishing settings.
          </Typography>
          <Stack spacing={3} sx={{ maxWidth: 500 }}>
            <TextField
              label="Namespace"
              value={scmNamespace}
              onChange={(e) => setScmNamespace(e.target.value)}
              placeholder="e.g., bconline"
              required
              fullWidth
              helperText="Your organization identifier"
            />
            <TextField
              label="Module Name"
              value={scmName}
              onChange={(e) => setScmName(e.target.value)}
              placeholder="e.g., networking-vpc"
              required
              fullWidth
            />
            <TextField
              label="Provider"
              value={scmSystem}
              onChange={(e) => setScmSystem(e.target.value)}
              placeholder="e.g., aws"
              required
              fullWidth
              helperText="Cloud provider this module targets (aws, azure, google, etc.)"
            />
            <TextField
              label="Description (optional)"
              value={scmDescription}
              onChange={(e) => setScmDescription(e.target.value)}
              fullWidth
              multiline
              rows={2}
            />

            {scmError && <Alert severity="error">{scmError}</Alert>}

            <Button
              variant="contained"
              onClick={handleScmProceed}
              disabled={scmCreating || !scmNamespace || !scmName || !scmSystem}
              startIcon={scmCreating ? <CircularProgress size={18} /> : <SCMIcon />}
              size="large"
            >
              {scmCreating ? 'Creating...' : 'Continue to Repository Selection'}
            </Button>
          </Stack>
        </>
      )}
    </Box>
  );

  const renderFileUploadForm = () => (
    <Box sx={{ p: 3 }}>
      <Button
        startIcon={<ArrowBack />}
        onClick={() => { setModuleMethod('choose'); setError(null); setSuccess(null); }}
        sx={{ mb: 2 }}
      >
        Back
      </Button>
      <Typography variant="h6" gutterBottom>
        Upload Terraform Module
      </Typography>
      <Box sx={{ mb: 3, p: 2, bgcolor: (theme) => theme.palette.mode === 'dark' ? 'grey.800' : 'grey.50', borderRadius: 1 }}>
        <Typography variant="body2" color="text.secondary" gutterBottom>
          <strong>Requirements:</strong>
        </Typography>
        <Typography variant="body2" color="text.secondary" component="div">
          • Package your module as a <strong>.tar.gz</strong> or <strong>.tgz</strong> file<br />
          • Include all <strong>.tf</strong> files (main.tf, variables.tf, outputs.tf)<br />
          • Add a <strong>README.md</strong> with usage documentation<br />
          • Use semantic versioning (1.0.0, 2.1.3, etc.)<br />
          • Module address format: <strong>namespace/name/provider</strong>
        </Typography>
      </Box>

      <Stack spacing={3}>
        <TextField
          label="Namespace"
          value={moduleNamespace}
          onChange={(e) => setModuleNamespace(e.target.value)}
          placeholder="e.g., bconline"
          required
          fullWidth
          helperText="Your organization identifier (like a GitHub or DevOps org)."
        />
        <TextField
          label="Description"
          value={moduleDescription}
          onChange={(e) => setModuleDescription(e.target.value)}
          placeholder="e.g., Creates a VPC with public and private subnets"
          fullWidth
          multiline
          rows={3}
          helperText="Brief description of what this module does and its purpose."
        />
        <TextField
          label="Module Name"
          value={moduleName}
          onChange={(e) => setModuleName(e.target.value)}
          placeholder="e.g., networking-vpc"
          required
          fullWidth
          helperText="Descriptive name for what the module does"
        />
        <TextField
          label="Provider"
          value={moduleProvider}
          onChange={(e) => setModuleProvider(e.target.value)}
          placeholder="e.g., aws"
          required
          fullWidth
          helperText="Cloud provider this module targets (aws, azure, google, etc.)"
        />
        <TextField
          label="Version"
          value={moduleVersion}
          onChange={(e) => setModuleVersion(e.target.value)}
          placeholder="e.g., 1.0.0"
          required
          fullWidth
          helperText="Semantic version in format X.Y.Z (e.g., 1.0.0, 2.1.3). Use 0.x.x for pre-release."
        />

        <Box>
          <input
            id="module-file-input"
            type="file"
            accept=".tar.gz,.tgz"
            onChange={handleModuleFileChange}
            style={{ display: 'none' }}
          />
          <label htmlFor="module-file-input">
            <Button
              variant="outlined"
              component="span"
              startIcon={<CloudUpload />}
              fullWidth
              sx={{ py: 2 }}
            >
              {moduleFile ? moduleFile.name : 'Select Module File (.tar.gz)'}
            </Button>
          </label>
        </Box>

        {error && <Alert severity="error">{error}</Alert>}
        {success && <Alert severity="success">{success}</Alert>}

        <Button
          variant="contained"
          onClick={handleModuleUpload}
          disabled={uploading || !moduleFile}
          startIcon={uploading ? <CircularProgress size={20} /> : <CloudUpload />}
          size="large"
        >
          {uploading ? 'Uploading...' : 'Upload Module'}
        </Button>
      </Stack>
    </Box>
  );

  return (
    <Container maxWidth="md" sx={{ py: 4 }}>
      <Typography variant="h4" gutterBottom>
        Upload
      </Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
        Upload Terraform modules and providers to your registry
      </Typography>

      <Paper sx={{ width: '100%' }}>
        <Tabs value={tabValue} onChange={handleTabChange}>
          <Tab label="Module" />
          <Tab label="Provider" />
        </Tabs>

        {/* Module Tab */}
        <TabPanel value={tabValue} index={0}>
          {moduleMethod === 'choose' && renderModuleMethodChooser()}
          {moduleMethod === 'upload' && renderFileUploadForm()}
          {moduleMethod === 'scm' && renderScmMetadataForm()}
        </TabPanel>

        {/* Provider Upload Tab */}
        <TabPanel value={tabValue} index={1}>
          <Box sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Upload Terraform Provider
            </Typography>
            <Box sx={{ mb: 3, p: 2, bgcolor: (theme) => theme.palette.mode === 'dark' ? 'grey.800' : 'grey.50', borderRadius: 1 }}>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                <strong>Requirements:</strong>
              </Typography>
              <Typography variant="body2" color="text.secondary" component="div">
                • Package provider binary as a <strong>.zip</strong> file<br />
                • Upload each OS/Architecture combination separately<br />
                • Use semantic versioning matching the binary version<br />
                • Filename should be: <strong>terraform-provider-NAME_VERSION_OS_ARCH.zip</strong><br />
                • Provider address format: <strong>namespace/type</strong>
              </Typography>
            </Box>

            <Stack spacing={3}>
              <TextField
                label="Namespace"
                value={providerNamespace}
                onChange={(e) => setProviderNamespace(e.target.value)}
                placeholder="e.g., myorg"
                required
                fullWidth
                helperText="Your organization identifier."
              />
              <TextField
                label="Provider Name"
                value={providerName}
                onChange={(e) => setProviderName(e.target.value)}
                placeholder="e.g., custom-api"
                required
                fullWidth
                helperText="Provider type name (e.g., 'aws', 'azurerm', 'custom-api'). Lowercase only."
              />
              <TextField
                label="Version"
                value={providerVersion}
                onChange={(e) => setProviderVersion(e.target.value)}
                placeholder="e.g., 1.0.0"
                required
                fullWidth
                helperText="Semantic version in format X.Y.Z (e.g., 1.0.0, 2.1.3). Must match binary version."
              />

              <FormControl fullWidth required>
                <InputLabel>Operating System</InputLabel>
                <Select
                  value={providerOS}
                  label="Operating System"
                  onChange={(e: SelectChangeEvent) => setProviderOS(e.target.value)}
                >
                  <MenuItem value="linux">Linux</MenuItem>
                  <MenuItem value="darwin">macOS (Darwin)</MenuItem>
                  <MenuItem value="windows">Windows</MenuItem>
                </Select>
                <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, ml: 1.75 }}>
                  Target operating system for this provider binary
                </Typography>
              </FormControl>

              <FormControl fullWidth required>
                <InputLabel>Architecture</InputLabel>
                <Select
                  value={providerArch}
                  label="Architecture"
                  onChange={(e: SelectChangeEvent) => setProviderArch(e.target.value)}
                >
                  <MenuItem value="amd64">AMD64 (x86_64)</MenuItem>
                  <MenuItem value="arm64">ARM64</MenuItem>
                  <MenuItem value="386">386 (x86)</MenuItem>
                </Select>
                <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, ml: 1.75 }}>
                  CPU architecture for this provider binary (most common: amd64)
                </Typography>
              </FormControl>

              <Box>
                <input
                  id="provider-file-input"
                  type="file"
                  accept=".zip"
                  onChange={handleProviderFileChange}
                  style={{ display: 'none' }}
                />
                <label htmlFor="provider-file-input">
                  <Button
                    variant="outlined"
                    component="span"
                    startIcon={<CloudUpload />}
                    fullWidth
                    sx={{ py: 2 }}
                  >
                    {providerFile ? providerFile.name : 'Select Provider Binary (.zip)'}
                  </Button>
                </label>
              </Box>

              {error && <Alert severity="error">{error}</Alert>}
              {success && <Alert severity="success">{success}</Alert>}

              <Button
                variant="contained"
                onClick={handleProviderUpload}
                disabled={uploading || !providerFile}
                startIcon={uploading ? <CircularProgress size={20} /> : <CloudUpload />}
                size="large"
              >
                {uploading ? 'Uploading...' : 'Upload Provider'}
              </Button>
            </Stack>
          </Box>
        </TabPanel>
      </Paper>
    </Container>
  );
};

export default UploadPage;
