import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  ListItemIcon,
  Chip,
  CircularProgress,
  TextField,
  InputAdornment,
  Alert,
  Divider,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  Folder as FolderIcon,
  Code as CodeIcon,
  Tag as TagIcon,
  Search as SearchIcon,
  ExpandMore as ExpandMoreIcon,
  Lock as LockIcon,
  Public as PublicIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material';
import type { SCMRepository, SCMTag, SCMBranch } from '../types/scm';
import apiClient from '../services/api';

interface RepositoryBrowserProps {
  providerId: string;
  onRepositorySelect?: (repository: SCMRepository) => void;
  onTagSelect?: (repository: SCMRepository, tag: SCMTag) => void;
  selectedRepository?: SCMRepository | null;
}

const RepositoryBrowser: React.FC<RepositoryBrowserProps> = ({
  providerId,
  onRepositorySelect,
  onTagSelect,
  selectedRepository,
}) => {
  const [repositories, setRepositories] = useState<SCMRepository[]>([]);
  const [tags, setTags] = useState<SCMTag[]>([]);
  const [branches, setBranches] = useState<SCMBranch[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadingTags, setLoadingTags] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedRepo, setExpandedRepo] = useState<string | null>(null);

  useEffect(() => {
    if (providerId) {
      loadRepositories();
    }
  }, [providerId]);

  useEffect(() => {
    if (selectedRepository && expandedRepo === selectedRepository.full_name) {
      loadTagsAndBranches(selectedRepository);
    }
  }, [selectedRepository, expandedRepo]);

  const loadRepositories = async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await apiClient.listSCMRepositories(providerId);
      const repos = response.repositories || [];
      setRepositories(repos);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load repositories');
      console.error('Error loading repositories:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadTagsAndBranches = async (_repository: SCMRepository) => {
    try {
      setLoadingTags(true);
      setError(null);

      // Mock data - in real implementation, call backend API
      const mockTags: SCMTag[] = [
        {
          name: 'v1.0.0',
          commit_sha: 'abc123def456',
          commit_message: 'Release version 1.0.0',
          created_at: '2024-01-15T10:00:00Z',
          tagger: 'user@example.com',
        },
        {
          name: 'v0.9.0',
          commit_sha: 'def456ghi789',
          commit_message: 'Release version 0.9.0',
          created_at: '2023-12-10T10:00:00Z',
          tagger: 'user@example.com',
        },
      ];

      const mockBranches: SCMBranch[] = [
        {
          name: 'main',
          commit_sha: 'abc123def456',
          protected: true,
        },
        {
          name: 'develop',
          commit_sha: 'xyz789abc123',
          protected: false,
        },
      ];

      setTags(mockTags);
      setBranches(mockBranches);
    } catch (err: any) {
      setError('Failed to load tags and branches');
      console.error('Error loading tags/branches:', err);
    } finally {
      setLoadingTags(false);
    }
  };

  const handleRepositoryClick = (repository: SCMRepository) => {
    if (expandedRepo === repository.full_name) {
      setExpandedRepo(null);
    } else {
      setExpandedRepo(repository.full_name);
      if (onRepositorySelect) {
        onRepositorySelect(repository);
      }
    }
  };

  const handleTagClick = (repository: SCMRepository, tag: SCMTag) => {
    if (onTagSelect) {
      onTagSelect(repository, tag);
    }
  };

  const filteredRepositories = repositories.filter(
    (repo) =>
      repo.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      repo.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="200px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Box mb={2} display="flex" gap={1}>
        <TextField
          fullWidth
          size="small"
          placeholder="Search repositories..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          }}
        />
        <Tooltip title="Refresh">
          <IconButton onClick={loadRepositories} size="small">
            <RefreshIcon />
          </IconButton>
        </Tooltip>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {filteredRepositories.length === 0 ? (
        <Typography variant="body2" color="textSecondary" align="center" sx={{ py: 4 }}>
          No repositories found
        </Typography>
      ) : (
        <Box>
          {filteredRepositories.map((repo) => (
            <Accordion
              key={repo.id}
              expanded={expandedRepo === repo.full_name}
              onChange={() => handleRepositoryClick(repo)}
              sx={{ mb: 1 }}
            >
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Box display="flex" alignItems="center" width="100%">
                  <ListItemIcon sx={{ minWidth: 40 }}>
                    <FolderIcon />
                  </ListItemIcon>
                  <Box flexGrow={1}>
                    <Typography variant="subtitle1">{repo.full_name}</Typography>
                    {repo.description && (
                      <Typography variant="caption" color="textSecondary">
                        {repo.description}
                      </Typography>
                    )}
                  </Box>
                  <Chip
                    icon={repo.private ? <LockIcon /> : <PublicIcon />}
                    label={repo.private ? 'Private' : 'Public'}
                    size="small"
                    sx={{ mr: 2 }}
                  />
                </Box>
              </AccordionSummary>

              <AccordionDetails>
                {loadingTags ? (
                  <Box display="flex" justifyContent="center" py={2}>
                    <CircularProgress size={24} />
                  </Box>
                ) : (
                  <>
                    {/* Tags Section */}
                    <Typography variant="subtitle2" gutterBottom>
                      Tags
                    </Typography>
                    {tags.length === 0 ? (
                      <Typography variant="body2" color="textSecondary" sx={{ py: 1 }}>
                        No tags found
                      </Typography>
                    ) : (
                      <List dense>
                        {tags.map((tag) => (
                          <ListItem
                            key={tag.name}
                            disablePadding
                            secondaryAction={
                              <Chip
                                label={tag.commit_sha.substring(0, 7)}
                                size="small"
                                variant="outlined"
                              />
                            }
                          >
                            <ListItemButton
                              onClick={() => handleTagClick(repo, tag)}
                              sx={{ borderRadius: 1 }}
                            >
                              <ListItemIcon sx={{ minWidth: 36 }}>
                                <TagIcon fontSize="small" />
                              </ListItemIcon>
                              <ListItemText
                                primary={tag.name}
                                secondary={
                                  tag.created_at
                                    ? new Date(tag.created_at).toLocaleDateString()
                                    : undefined
                                }
                              />
                            </ListItemButton>
                          </ListItem>
                        ))}
                      </List>
                    )}

                    <Divider sx={{ my: 2 }} />

                    {/* Branches Section */}
                    <Typography variant="subtitle2" gutterBottom>
                      Branches
                    </Typography>
                    <List dense>
                      {branches.map((branch) => (
                        <ListItem
                          key={branch.name}
                          secondaryAction={
                            branch.protected && (
                              <Chip label="Protected" size="small" color="primary" />
                            )
                          }
                        >
                          <ListItemIcon sx={{ minWidth: 36 }}>
                            <CodeIcon fontSize="small" />
                          </ListItemIcon>
                          <ListItemText
                            primary={branch.name}
                            secondary={branch.commit_sha.substring(0, 7)}
                          />
                        </ListItem>
                      ))}
                    </List>
                  </>
                )}
              </AccordionDetails>
            </Accordion>
          ))}
        </Box>
      )}
    </Box>
  );
};

export default RepositoryBrowser;
