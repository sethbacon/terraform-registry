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
  FormGroup,
  FormControlLabel,
  Checkbox,
  Chip,
  Tooltip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import {
  Delete as DeleteIcon,
  Add as AddIcon,
  ContentCopy as CopyIcon,
  Visibility as VisibilityIcon,
  VisibilityOff as VisibilityOffIcon,
  Info as InfoIcon,
} from '@mui/icons-material';
import api from '../../services/api';
import { APIKey, UserMembership } from '../../types';
import { REGISTRY_HOST } from '../../config';
import { useAuth } from '../../contexts/AuthContext';
import { AVAILABLE_SCOPES } from '../../types/rbac';

const APIKeysPage: React.FC = () => {
  const { allowedScopes, roleTemplate, user } = useAuth();
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [memberships, setMemberships] = useState<UserMembership[]>([]);
  const [membershipsLoading, setMembershipsLoading] = useState(true);

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
    scopes: [] as string[],
  });

  // Check if user has admin scope (which grants all permissions)
  const hasAdminScope = allowedScopes.includes('admin');

  // Get available scopes for this user
  const availableScopes = hasAdminScope
    ? AVAILABLE_SCOPES.map((s) => s.value)
    : allowedScopes;

  useEffect(() => {
    loadAPIKeys();
    loadMemberships();
  }, [user?.id]);

  const loadMemberships = async () => {
    if (!user?.id) return;
    try {
      setMembershipsLoading(true);
      // Use self-access endpoint that doesn't require users:read scope
      const userMemberships = await api.getCurrentUserMemberships();
      setMemberships(userMemberships);
    } catch (err) {
      console.error('Failed to load memberships:', err);
      setMemberships([]);
    } finally {
      setMembershipsLoading(false);
    }
  };

  const loadAPIKeys = async () => {
    try {
      setLoading(true);
      setError(null);
      const keys = await api.listAPIKeys();
      // Ensure keys is always an array
      setApiKeys(Array.isArray(keys) ? keys : []);
    } catch (err) {
      console.error('Failed to load API keys:', err);
      setApiKeys([]);
      setError('Failed to load API keys. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = () => {
    // Reset state when opening dialog
    setNewKeyValue(null);
    // Default to common read scopes that the user has access to
    const defaultScopes = ['modules:read', 'providers:read'].filter((s) =>
      availableScopes.includes(s)
    );
    // Default to first organization membership
    const defaultOrgId = memberships.length > 0 ? memberships[0].organization_id : '';
    setFormData({
      name: '',
      description: '',
      organization_id: defaultOrgId,
      scopes: defaultScopes,
    });
    setError(null);
    setOpenDialog(true);
  };

  const handleScopeToggle = (scope: string) => {
    setFormData((prev) => {
      const newScopes = prev.scopes.includes(scope)
        ? prev.scopes.filter((s) => s !== scope)
        : [...prev.scopes, scope];
      return { ...prev, scopes: newScopes };
    });
  };

  const getScopeInfo = (scopeValue: string) => {
    return AVAILABLE_SCOPES.find((s) => s.value === scopeValue) || {
      value: scopeValue,
      label: scopeValue,
      description: '',
    };
  };

  const handleCloseDialog = () => {
    // Just close the dialog - don't clear state yet
    setOpenDialog(false);
  };

  const handleCreateAPIKey = async () => {
    try {
      setError(null);
      // Use selected organization or first membership
      const orgId = formData.organization_id || (memberships.length > 0 ? memberships[0].organization_id : '');
      if (!orgId) {
        setError('You must be a member of an organization to create API keys.');
        return;
      }
      const response = await api.createAPIKey({
        name: formData.name,
        organization_id: orgId,
        description: formData.description || undefined,
        scopes: formData.scopes,
      });
      setNewKeyValue(response.key);
      await loadAPIKeys();
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

      {!membershipsLoading && memberships.length === 0 && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          You are not a member of any organization. Contact an administrator to add you to an organization before creating API keys.
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
                  <TableCell>Created By</TableCell>
                  <TableCell>Last Used</TableCell>
                  <TableCell>Created</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {apiKeys.map((apiKey) => (
                  <TableRow key={apiKey.id}>
                      <TableCell>
                        <Typography fontWeight="medium">{apiKey.name || '-'}</Typography>
                      </TableCell>
                      <TableCell>{apiKey.description || '-'}</TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Typography
                            variant="body2"
                            sx={{ fontFamily: 'monospace', fontSize: '0.875rem' }}
                          >
                            {showKeys.has(apiKey.id)
                              ? (apiKey.key_prefix ? apiKey.key_prefix + '...' : '-')
                              : (apiKey.key_prefix ? maskKey(apiKey.key_prefix + '...') : '-')}
                          </Typography>
                          {apiKey.key_prefix && (
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
                          )}
                        </Box>
                      </TableCell>
                      <TableCell>{apiKey.user_name || '-'}</TableCell>
                      <TableCell>
                        {apiKey.last_used_at && !isNaN(Date.parse(apiKey.last_used_at))
                          ? new Date(apiKey.last_used_at).toLocaleDateString()
                          : 'Never'}
                      </TableCell>
                      <TableCell>
                        {apiKey.created_at && !isNaN(Date.parse(apiKey.created_at))
                          ? new Date(apiKey.created_at).toLocaleDateString()
                          : '-'}
                      </TableCell>
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
      <Dialog 
        open={openDialog} 
        onClose={newKeyValue ? undefined : handleCloseDialog}
        disableEscapeKeyDown={!!newKeyValue}
        maxWidth="sm" 
        fullWidth
      >
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
              {!roleTemplate && (
                <Alert severity="warning" icon={<InfoIcon />}>
                  You don't have a role template assigned. Contact an administrator to assign a role
                  before creating API keys.
                </Alert>
              )}
              {roleTemplate && (
                <Alert severity="info" icon={<InfoIcon />}>
                  Your role: <strong>{roleTemplate.display_name}</strong>. You can only create API keys
                  with scopes that match your role permissions.
                </Alert>
              )}
              {memberships.length === 0 && (
                <Alert severity="error">
                  You must be a member of an organization to create API keys.
                </Alert>
              )}
              {memberships.length > 0 && (
                <FormControl fullWidth>
                  <InputLabel>Organization</InputLabel>
                  <Select
                    value={formData.organization_id}
                    onChange={(e) => setFormData({ ...formData, organization_id: e.target.value })}
                    label="Organization"
                  >
                    {memberships.map((m) => (
                      <MenuItem key={m.organization_id} value={m.organization_id}>
                        {m.organization_name} {m.role_template_display_name && `(${m.role_template_display_name})`}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              )}
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
              <Box>
                <Typography variant="subtitle2" gutterBottom>
                  Scopes
                </Typography>
                <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                  Select the permissions for this API key. You can only select scopes within your role's permissions.
                </Typography>
                {availableScopes.length === 0 ? (
                  <Alert severity="error">
                    No scopes available. Please contact an administrator to assign you a role.
                  </Alert>
                ) : (
                  <FormGroup>
                    {availableScopes.map((scope) => {
                      const info = getScopeInfo(scope);
                      return (
                        <Tooltip key={scope} title={info.description} placement="right">
                          <FormControlLabel
                            control={
                              <Checkbox
                                checked={formData.scopes.includes(scope)}
                                onChange={() => handleScopeToggle(scope)}
                                size="small"
                              />
                            }
                            label={
                              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <Typography variant="body2">{info.label}</Typography>
                                <Chip
                                  label={scope}
                                  size="small"
                                  sx={{
                                    height: 20,
                                    fontSize: '0.7rem',
                                    backgroundColor:
                                      scope === 'admin'
                                        ? 'error.light'
                                        : scope.includes(':write') || scope.includes(':manage')
                                        ? 'warning.light'
                                        : 'success.light',
                                  }}
                                />
                              </Box>
                            }
                          />
                        </Tooltip>
                      );
                    })}
                  </FormGroup>
                )}
                {formData.scopes.length === 0 && availableScopes.length > 0 && (
                  <Typography variant="caption" color="error">
                    Please select at least one scope
                  </Typography>
                )}
              </Box>
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
              <Button
                onClick={handleCreateAPIKey}
                variant="contained"
                disabled={
                  !formData.name ||
                  formData.scopes.length === 0 ||
                  availableScopes.length === 0 ||
                  memberships.length === 0
                }
              >
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
              backgroundColor: (theme) => theme.palette.mode === 'dark' ? '#2d2d2d' : '#f5f5f5',
              color: (theme) => theme.palette.mode === 'dark' ? '#e6e6e6' : '#1e1e1e',
              borderRadius: 1,
              overflow: 'auto',
              fontSize: '0.875rem',
            }}
          >
            {`# ~/.terraformrc (Unix) or %APPDATA%/terraform.rc (Windows)
credentials "${REGISTRY_HOST}" {
  token = "YOUR_API_KEY_HERE"
}`}
          </Box>
        </Typography>
      </Paper>
    </Container>
  );
};

export default APIKeysPage;
