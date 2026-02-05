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
  Chip,
  CircularProgress,
  Alert,
  Stack,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
} from '@mui/material';
import {
  Edit as EditIcon,
  Delete as DeleteIcon,
  Add as AddIcon,
  People as PeopleIcon,
} from '@mui/icons-material';
import api from '../../services/api';
import { Organization } from '../../types';

const OrganizationsPage: React.FC = () => {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Dialog state
  const [openDialog, setOpenDialog] = useState(false);
  const [editingOrg, setEditingOrg] = useState<Organization | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [orgToDelete, setOrgToDelete] = useState<Organization | null>(null);
  const [membersDialogOpen, setMembersDialogOpen] = useState(false);
  const [selectedOrg, setSelectedOrg] = useState<Organization | null>(null);

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    display_name: '',
  });

  useEffect(() => {
    loadOrganizations();
  }, []);

  const loadOrganizations = async () => {
    try {
      setLoading(true);
      setError(null);
      const orgs = await api.listOrganizations();
      setOrganizations(orgs || []);
    } catch (err) {
      console.error('Failed to load organizations:', err);
      // In dev mode, just show empty list instead of error
      setOrganizations([]);
      if (!import.meta.env.DEV) {
        setError('Failed to load organizations. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (org?: Organization) => {
    if (org) {
      setEditingOrg(org);
      setFormData({
        name: org.name,
        display_name: org.display_name || '',
      });
    } else {
      setEditingOrg(null);
      setFormData({
        name: '',
        display_name: '',
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setEditingOrg(null);
    setError(null);
  };

  const handleSaveOrganization = async () => {
    try {
      setError(null);
      if (editingOrg) {
        await api.updateOrganization(editingOrg.id, {
          display_name: formData.display_name,
        });
      } else {
        await api.createOrganization({
          name: formData.name,
          display_name: formData.display_name,
        });
      }
      handleCloseDialog();
      loadOrganizations();
    } catch (err: any) {
      console.error('Failed to save organization:', err);
      console.error('Error response:', err.response);
      console.error('Error data:', err.response?.data);
      const errorMessage = err.response?.data?.error || err.message || 'Failed to save organization. Please try again.';
      setError(errorMessage);
    }
  };

  const handleDeleteClick = (org: Organization) => {
    setOrgToDelete(org);
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!orgToDelete) return;

    try {
      setError(null);
      await api.deleteOrganization(orgToDelete.id);
      setDeleteDialogOpen(false);
      setOrgToDelete(null);
      loadOrganizations();
    } catch (err: any) {
      console.error('Failed to delete organization:', err);
      setError(err.response?.data?.error || 'Failed to delete organization. Please try again.');
    }
  };

  const handleViewMembers = (org: Organization) => {
    setSelectedOrg(org);
    setMembersDialogOpen(true);
  };

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box>
          <Typography variant="h4" gutterBottom>
            Organizations
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Manage organizations and their members
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => handleOpenDialog()}
        >
          Add Organization
        </Button>
      </Box>

      {error && !import.meta.env.DEV && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Organizations Table */}
      <Paper>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
            <CircularProgress />
          </Box>
        ) : organizations.length === 0 ? (
          <Box sx={{ textAlign: 'center', py: 8 }}>
            <Typography color="text.secondary">No organizations found</Typography>
            <Button
              variant="outlined"
              startIcon={<AddIcon />}
              onClick={() => handleOpenDialog()}
              sx={{ mt: 2 }}
            >
              Create First Organization
            </Button>
          </Box>
        ) : (
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Name</TableCell>
                  <TableCell>Display Name</TableCell>
                  <TableCell>Members</TableCell>
                  <TableCell>Created</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {organizations.map((org) => (
                  <TableRow key={org.id}>
                    <TableCell>
                      <Typography fontWeight="medium">{org.name}</Typography>
                    </TableCell>
                    <TableCell>{org.display_name || '-'}</TableCell>
                    <TableCell>
                      <Button
                        size="small"
                        startIcon={<PeopleIcon />}
                        onClick={() => handleViewMembers(org)}
                      >
                        View Members
                      </Button>
                    </TableCell>
                    <TableCell>{new Date(org.created_at).toLocaleDateString()}</TableCell>
                    <TableCell align="right">
                      <IconButton
                        size="small"
                        onClick={() => handleOpenDialog(org)}
                        color="primary"
                      >
                        <EditIcon />
                      </IconButton>
                      <IconButton
                        size="small"
                        onClick={() => handleDeleteClick(org)}
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

      {/* Add/Edit Organization Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>{editingOrg ? 'Edit Organization' : 'Add Organization'}</DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ mt: 2 }}>
            <TextField
              label="Name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              required
              fullWidth
              helperText="Organization name (e.g., myorg)"
            />
            <TextField
              label="Display Name"
              value={formData.display_name}
              onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
              required
              multiline
              rows={3}
              fullWidth
              helperText="Display name for the organization"
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          <Button 
            onClick={handleSaveOrganization} 
            variant="contained"
            disabled={!formData.name.trim() || !formData.display_name.trim()}
          >
            {editingOrg ? 'Save' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>Delete Organization</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete organization "{orgToDelete?.name}"? This action cannot
            be undone.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleDeleteConfirm} color="error" variant="contained">
            Delete
          </Button>
        </DialogActions>
      </Dialog>

      {/* Members Dialog */}
      <Dialog
        open={membersDialogOpen}
        onClose={() => setMembersDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Organization Members - {selectedOrg?.name}</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            View and manage members of this organization
          </Typography>
          <Paper variant="outlined">
            <List>
              <ListItem>
                <ListItemText
                  primary="Member management"
                  secondary="Member management functionality can be added here"
                />
              </ListItem>
            </List>
          </Paper>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setMembersDialogOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default OrganizationsPage;
