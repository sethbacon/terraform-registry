import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Box,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Button,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  CircularProgress,
  Alert,
  Stack,
  InputAdornment,
} from '@mui/material';
import {
  Delete as DeleteIcon,
  Add as AddIcon,
  ContentCopy as CopyIcon,
  Visibility as VisibilityIcon,
  VisibilityOff as VisibilityOffIcon,
} from '@mui/icons-material';
import api from '../../services/api';
import { APIKey } from '../../types';

const APIKeysPage: React.FC = () => {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Dialog state
  const [openDialog, setOpenDialog] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [keyToDelete, setKeyToDelete] = useState<APIKey | null>(null);
  const [newKeyValue, setNewKeyValue] = useState<string | null>(null);
  const [copiedKey, setCopiedKey] = useState(false);
  const [showKeys, setShowKeys] = useState<Set<string>>(new Set());

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    organization_id: '',
    scopes: ['read', 'write'] as string[],
  });

  useEffect(() => {
    loadAPIKeys();
  }, []);

  const loadAPIKeys = async () => {
    try {
      setLoading(true);
      setError(null);
      const keys = await api.listAPIKeys();
      setApiKeys(keys);
    } catch (err) {
      console.error('Failed to load API keys:', err);
      setError('Failed to load API keys. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = () => {
    setFormData({
      name: '',
      description: '',
      organization_id: '',
      scopes: ['read', 'write'],
    });
    setNewKeyValue(null);
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setFormData({ name: '', description: '', organization_id: '', scopes: ['read', 'write'] });
    setNewKeyValue(null);
    setError(null);
  };

  const handleCreateAPIKey = async () => {
    try {
      setError(null);
      const response = await api.createAPIKey({
        name: formData.name,
        organization_id: formData.organization_id || 'default',
        scopes: formData.scopes,
      });
      setNewKeyValue(response.key);
      loadAPIKeys();
    } catch (err: any) {
      console.error('Failed to create API key:', err);
      setError(err.response?.data?.error || 'Failed to create API key. Please try again.');
    }
  };

  const handleDeleteClick = (key: APIKey) => {
    setKeyToDelete(key);
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!keyToDelete) return;

    try {
      setError(null);
      await api.deleteAPIKey(keyToDelete.id);
      setDeleteDialogOpen(false);
      setKeyToDelete(null);
      loadAPIKeys();
    } catch (err: any) {
      console.error('Failed to delete API key:', err);
      setError(err.response?.data?.error || 'Failed to delete API key. Please try again.');
    }
  };

  const handleCopyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    setCopiedKey(true);
    setTimeout(() => setCopiedKey(false), 2000);
  };

  const toggleShowKey = (keyId: string) => {
    const newShowKeys = new Set(showKeys);
    if (newShowKeys.has(keyId)) {
      newShowKeys.delete(keyId);
    } else {
      newShowKeys.add(keyId);
    }
    setShowKeys(newShowKeys);
  };

  const maskKey = (key: string) => {
    return key.substring(0, 8) + '...' + key.substring(key.length - 4);
  };

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box>
          <Typography variant="h4" gutterBottom>
            API Keys
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Manage API keys for Terraform CLI authentication
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleOpenDialog}
        >
          Create API Key
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* API Keys Table */}
      <Paper>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
            <CircularProgress />
          </Box>
        ) : apiKeys.length === 0 ? (
          <Box sx={{ textAlign: 'center', py: 8 }}>
            <Typography color="text.secondary">No API keys found</Typography>
            <Button
              variant="outlined"
              startIcon={<AddIcon />}
              onClick={handleOpenDialog}
              sx={{ mt: 2 }}
            >
              Create First API Key
            </Button>
          </Box>
        ) : (
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Name</TableCell>
                  <TableCell>Description</TableCell>
                  <TableCell>Key</TableCell>
                  <TableCell>Last Used</TableCell>
                  <TableCell>Created</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {apiKeys.map((apiKey) => (
                  <TableRow key={apiKey.id}>
                    <TableCell>
                      <Typography fontWeight="medium">{apiKey.name}</Typography>
                    </TableCell>
                    <TableCell>{apiKey.description || '-'}</TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <Typography
                          variant="body2"
                          sx={{ fontFamily: 'monospace', fontSize: '0.875rem' }}
                        >
                          {showKeys.has(apiKey.id) ? apiKey.key_prefix + '...' : maskKey(apiKey.key_prefix + '...')}
                        </Typography>
                        <IconButton
                          size="small"
                          onClick={() => toggleShowKey(apiKey.id)}
                        >
                          {showKeys.has(apiKey.id) ? (
                            <VisibilityOffIcon fontSize="small" />
                          ) : (
                            <VisibilityIcon fontSize="small" />
                          )}
                        </IconButton>
                      </Box>
                    </TableCell>
                    <TableCell>
                      {apiKey.last_used_at
                        ? new Date(apiKey.last_used_at).toLocaleDateString()
                        : 'Never'}
                    </TableCell>
                    <TableCell>{new Date(apiKey.created_at).toLocaleDateString()}</TableCell>
                    <TableCell align="right">
                      <IconButton
                        size="small"
                        onClick={() => handleDeleteClick(apiKey)}
                        color="error"
                      >
                        <DeleteIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </Paper>

      {/* Create API Key Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>Create API Key</DialogTitle>
        <DialogContent>
          {newKeyValue ? (
            <Box sx={{ mt: 2 }}>
              <Alert severity="success" sx={{ mb: 3 }}>
                API key created successfully! Make sure to copy it now - you won't be able to see
                it again.
              </Alert>
              <TextField
                label="API Key"
                value={newKeyValue}
                fullWidth
                InputProps={{
                  readOnly: true,
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton onClick={() => handleCopyKey(newKeyValue)}>
                        <CopyIcon />
                      </IconButton>
                    </InputAdornment>
                  ),
                  sx: { fontFamily: 'monospace' },
                }}
              />
              {copiedKey && (
                <Typography variant="caption" color="success.main" sx={{ mt: 1, display: 'block' }}>
                  Copied to clipboard!
                </Typography>
              )}
            </Box>
          ) : (
            <Stack spacing={3} sx={{ mt: 2 }}>
              <TextField
                label="Name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
                fullWidth
                helperText="A descriptive name for this API key"
              />
              <TextField
                label="Description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                multiline
                rows={3}
                fullWidth
              />
            </Stack>
          )}
        </DialogContent>
        <DialogActions>
          {newKeyValue ? (
            <Button onClick={handleCloseDialog} variant="contained">
              Done
            </Button>
          ) : (
            <>
              <Button onClick={handleCloseDialog}>Cancel</Button>
              <Button onClick={handleCreateAPIKey} variant="contained">
                Create
              </Button>
            </>
          )}
        </DialogActions>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>Delete API Key</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete API key "{keyToDelete?.name}"? This action cannot be
            undone and will break any existing integrations using this key.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleDeleteConfirm} color="error" variant="contained">
            Delete
          </Button>
        </DialogActions>
      </Dialog>

      {/* Usage Instructions */}
      <Paper sx={{ p: 3, mt: 3 }}>
        <Typography variant="h6" gutterBottom>
          Using API Keys
        </Typography>
        <Typography variant="body2" color="text.secondary" component="div">
          To use an API key with Terraform CLI, add it to your Terraform CLI configuration:
          <Box
            component="pre"
            sx={{
              mt: 2,
              p: 2,
              backgroundColor: '#f5f5f5',
              borderRadius: 1,
              overflow: 'auto',
              fontSize: '0.875rem',
            }}
          >
            {`# ~/.terraformrc (Unix) or %APPDATA%/terraform.rc (Windows)
credentials "${window.location.host}" {
  token = "YOUR_API_KEY_HERE"
}`}
          </Box>
        </Typography>
      </Paper>
    </Container>
  );
};

export default APIKeysPage;
