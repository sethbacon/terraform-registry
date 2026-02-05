import React, { useState } from 'react';
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
  Stepper,
  Step,
  StepLabel,
  SelectChangeEvent,
} from '@mui/material';
import { CloudUpload } from '@mui/icons-material';
import api from '../../services/api';

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

const UploadPage: React.FC = () => {
  const [tabValue, setTabValue] = useState(0);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Module upload state
  const [moduleFile, setModuleFile] = useState<File | null>(null);
  const [moduleNamespace, setModuleNamespace] = useState('');
  const [moduleName, setModuleName] = useState('');
  const [moduleProvider, setModuleProvider] = useState('');
  const [moduleVersion, setModuleVersion] = useState('');

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

      await api.uploadModule(
        moduleNamespace,
        moduleName,
        moduleProvider,
        moduleVersion,
        moduleFile
      );

      setSuccess(`Module ${moduleNamespace}/${moduleName}/${moduleProvider} v${moduleVersion} uploaded successfully!`);
      
      // Reset form
      setModuleFile(null);
      setModuleNamespace('');
      setModuleName('');
      setModuleProvider('');
      setModuleVersion('');
      
      // Reset file input
      const fileInput = document.getElementById('module-file-input') as HTMLInputElement;
      if (fileInput) fileInput.value = '';
    } catch (err: any) {
      console.error('Failed to upload module:', err);
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

      await api.uploadProvider(
        providerNamespace,
        providerName,
        providerVersion,
        providerOS,
        providerArch,
        providerFile
      );

      setSuccess(`Provider ${providerNamespace}/${providerName} v${providerVersion} (${providerOS}/${providerArch}) uploaded successfully!`);
      
      // Reset form
      setProviderFile(null);
      setProviderNamespace('');
      setProviderName('');
      setProviderVersion('');
      setProviderOS('');
      setProviderArch('');
      
      // Reset file input
      const fileInput = document.getElementById('provider-file-input') as HTMLInputElement;
      if (fileInput) fileInput.value = '';
    } catch (err: any) {
      console.error('Failed to upload provider:', err);
      setError(err.response?.data?.error || 'Failed to upload provider. Please try again.');
    } finally {
      setUploading(false);
    }
  };

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
          <Tab label="Upload Module" />
          <Tab label="Upload Provider" />
        </Tabs>

        {/* Module Upload Tab */}
        <TabPanel value={tabValue} index={0}>
          <Box sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Upload Terraform Module
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Upload a .tar.gz file containing your Terraform module
            </Typography>

            <Stack spacing={3}>
              <TextField
                label="Namespace"
                value={moduleNamespace}
                onChange={(e) => setModuleNamespace(e.target.value)}
                placeholder="e.g., myorg"
                required
                fullWidth
              />
              <TextField
                label="Module Name"
                value={moduleName}
                onChange={(e) => setModuleName(e.target.value)}
                placeholder="e.g., vpc"
                required
                fullWidth
              />
              <TextField
                label="Provider"
                value={moduleProvider}
                onChange={(e) => setModuleProvider(e.target.value)}
                placeholder="e.g., aws"
                required
                fullWidth
              />
              <TextField
                label="Version"
                value={moduleVersion}
                onChange={(e) => setModuleVersion(e.target.value)}
                placeholder="e.g., 1.0.0"
                required
                fullWidth
                helperText="Semantic version (e.g., 1.0.0, 2.1.3)"
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
        </TabPanel>

        {/* Provider Upload Tab */}
        <TabPanel value={tabValue} index={1}>
          <Box sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Upload Terraform Provider
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Upload a provider binary for a specific platform
            </Typography>

            <Stack spacing={3}>
              <TextField
                label="Namespace"
                value={providerNamespace}
                onChange={(e) => setProviderNamespace(e.target.value)}
                placeholder="e.g., myorg"
                required
                fullWidth
              />
              <TextField
                label="Provider Name"
                value={providerName}
                onChange={(e) => setProviderName(e.target.value)}
                placeholder="e.g., custom"
                required
                fullWidth
              />
              <TextField
                label="Version"
                value={providerVersion}
                onChange={(e) => setProviderVersion(e.target.value)}
                placeholder="e.g., 1.0.0"
                required
                fullWidth
                helperText="Semantic version (e.g., 1.0.0, 2.1.3)"
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

      {/* Upload Guidelines */}
      <Paper sx={{ p: 3, mt: 3 }}>
        <Typography variant="h6" gutterBottom>
          Upload Guidelines
        </Typography>
        <Typography variant="body2" component="div" color="text.secondary">
          <strong>Modules:</strong>
          <ul>
            <li>Must be a .tar.gz archive containing Terraform .tf files</li>
            <li>Should include a README.md file</li>
            <li>Version must follow semantic versioning (e.g., 1.0.0)</li>
          </ul>
          <strong>Providers:</strong>
          <ul>
            <li>Must be a .zip file containing the provider binary</li>
            <li>Binary must be named terraform-provider-NAME_vVERSION</li>
            <li>Upload separate binaries for each OS/architecture combination</li>
          </ul>
        </Typography>
      </Paper>
    </Container>
  );
};

export default UploadPage;
