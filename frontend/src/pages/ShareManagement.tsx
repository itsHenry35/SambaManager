import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  FormControlLabel,
  IconButton,
  List,
  ListItem,
  ListItemText,
  Switch,
  TextField,
  Typography,
  Alert,
  Autocomplete,
} from '@mui/material';
import {
  Delete as DeleteIcon,
  CreateNewFolder as CreateNewFolderIcon,
  Edit as EditIcon,
} from '@mui/icons-material';
import { userAPI, shareAPI } from '../api';
import { handleResp, handleRespWithNotifySuccess } from '../utils/handleResp';
import type { UserResponse, ShareResponse } from '../types';

export function ShareManagement() {
  const { t } = useTranslation();
  const [shares, setShares] = useState<ShareResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [currentShareId, setCurrentShareId] = useState('');
  const [shareName, setShareName] = useState('');
  const [selectedOwner, setSelectedOwner] = useState('');
  const [selectedUsers, setSelectedUsers] = useState<string[]>([]);
  const [readOnly, setReadOnly] = useState(false);
  const [comment, setComment] = useState('');
  const [subPath, setSubPath] = useState('');
  const [error, setError] = useState('');
  const [ownerSearchQuery, setOwnerSearchQuery] = useState('');
  const [sharedWithSearchQuery, setSharedWithSearchQuery] = useState('');
  const [ownerSearchResults, setOwnerSearchResults] = useState<UserResponse[]>([]);
  const [sharedWithSearchResults, setSharedWithSearchResults] = useState<UserResponse[]>([]);

  const loadShares = async () => {
    const resp = await shareAPI.getShares();
    handleResp(resp, (data) => {
      setShares(data || []);
    });
  };

  // Search users for owner autocomplete
  useEffect(() => {
    if (ownerSearchQuery.length > 0) {
      const searchUsers = async () => {
        const resp = await userAPI.searchUsers(ownerSearchQuery);
        handleResp(resp, (data) => {
          setOwnerSearchResults(data || []);
        });
      };
      void searchUsers();
    } else {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setOwnerSearchResults([]);
    }
  }, [ownerSearchQuery]);

  // Search users for shared with autocomplete
  useEffect(() => {
    if (sharedWithSearchQuery.length > 0) {
      const searchUsers = async () => {
        const resp = await userAPI.searchUsers(sharedWithSearchQuery);
        handleResp(resp, (data) => {
          setSharedWithSearchResults(data || []);
        });
      };
      void searchUsers();
    } else {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setSharedWithSearchResults([]);
    }
  }, [sharedWithSearchQuery]);

  useEffect(() => {
    loadShares();
  }, []);

  const handleCreateShare = async () => {
    setError('');
    setLoading(true);

    const resp = await shareAPI.createShare({
      name: shareName || undefined,
      owner: selectedOwner,
      shared_with: selectedUsers,
      read_only: readOnly,
      comment: comment,
      sub_path: subPath || undefined,
    });

    handleRespWithNotifySuccess(
      resp,
      () => {
        handleCloseDialog();
        loadShares();
      },
      (message) => {
        setError(message);
      }
    );

    setLoading(false);
  };

  const handleUpdateShare = async () => {
    setError('');
    setLoading(true);

    const resp = await shareAPI.updateShare(currentShareId, {
      shared_with: selectedUsers,
      read_only: readOnly,
      comment: comment,
      sub_path: subPath || undefined,
    });

    handleRespWithNotifySuccess(
      resp,
      () => {
        handleCloseDialog();
        loadShares();
      },
      (message) => {
        setError(message);
      }
    );

    setLoading(false);
  };

  const handleEditShare = (share: ShareResponse) => {
    setEditMode(true);
    setCurrentShareId(share.id);
    setSelectedOwner(share.owner);
    setSelectedUsers(share.shared_with);
    setReadOnly(share.read_only);
    setComment(share.comment || '');
    setSubPath(share.sub_path || '');
    setShareName(share.id);
    setOpenDialog(true);
  };

  const handleDeleteShare = async (shareId: string) => {
    if (!confirm(t('shares.deleteConfirm', { name: shareId }))) {
      return;
    }

    setLoading(true);
    const resp = await shareAPI.deleteShare(shareId);
    handleRespWithNotifySuccess(
      resp,
      () => {
        loadShares();
      }
    );
    setLoading(false);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setEditMode(false);
    setCurrentShareId('');
    setShareName('');
    setSelectedOwner('');
    setSelectedUsers([]);
    setReadOnly(false);
    setComment('');
    setSubPath('');
    setError('');
  };

  const handleOpenCreateDialog = () => {
    setEditMode(false);
    setOpenDialog(true);
  };

  return (
    <Box>
      <Card>
        <CardHeader
          title={t('shares.title')}
          subheader={t('shares.subtitle')}
          action={
            <Button
              variant="contained"
              startIcon={<CreateNewFolderIcon />}
              onClick={handleOpenCreateDialog}
            >
              {t('shares.createShare')}
            </Button>
          }
        />
        <CardContent>
          {shares.length === 0 ? (
            <Typography variant="body2" color="text.secondary" align="center" sx={{ py: 4 }}>
              {t('shares.noShares')}
            </Typography>
          ) : (
            <List>
              {shares.map((share, index) => (
                <Box key={share.id}>
                  {index > 0 && <Divider />}
                  <ListItem
                    secondaryAction={
                      <Box>
                        <IconButton
                          edge="end"
                          aria-label="edit"
                          onClick={() => handleEditShare(share)}
                          disabled={loading}
                          sx={{ mr: 1 }}
                        >
                          <EditIcon />
                        </IconButton>
                        <IconButton
                          edge="end"
                          aria-label="delete"
                          onClick={() => handleDeleteShare(share.id)}
                          disabled={loading}
                          color="error"
                        >
                          <DeleteIcon />
                        </IconButton>
                      </Box>
                    }
                  >
                    <ListItemText
                      primary={
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
                          <Chip label={share.id} size="small" color="secondary" sx={{ fontWeight: 'bold' }} />
                          <Typography variant="body1" sx={{ fontWeight: 'bold' }}>
                            {share.owner}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            →
                          </Typography>
                          {share.shared_with.map((user) => (
                            <Chip key={user} label={user} size="small" variant="outlined" />
                          ))}
                          <Chip
                            label={share.read_only ? t('shares.readOnly') : t('shares.readWrite')}
                            size="small"
                            color={share.read_only ? 'default' : 'primary'}
                          />
                        </Box>
                      }
                      secondary={
                        <>
                          <Typography variant="body2" component="span">
                            {share.path}
                          </Typography>
                          {share.sub_path && (
                            <>
                              <br />
                              <Typography variant="caption" component="span" color="primary">
                                📁 {t('shares.subdirectory')}: {share.sub_path}
                              </Typography>
                            </>
                          )}
                          {share.comment && (
                            <>
                              <br />
                              <Typography variant="caption" component="span" color="text.secondary">
                                {share.comment}
                              </Typography>
                            </>
                          )}
                        </>
                      }
                    />
                  </ListItem>
                </Box>
              ))}
            </List>
          )}
        </CardContent>
      </Card>

      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>
          {editMode ? t('shares.editShare') : t('shares.createShare')}
        </DialogTitle>
        <DialogContent>
          <TextField
            margin="dense"
            label={t('shares.form.name')}
            type="text"
            fullWidth
            variant="outlined"
            value={shareName}
            onChange={(e) => setShareName(e.target.value)}
            disabled={loading || editMode}
            placeholder={t('shares.form.namePlaceholder')}
            helperText={editMode ? t('shares.form.nameHelperEdit') : t('shares.form.nameHelper')}
          />

          <Autocomplete
            freeSolo
            options={ownerSearchResults.map(u => u.username)}
            value={selectedOwner}
            onChange={(_e, newValue) => {
              setSelectedOwner(newValue || '');
              setSelectedUsers(prev => prev.filter(u => u !== newValue));
            }}
            onInputChange={(_e, value) => setOwnerSearchQuery(value)}
            disabled={loading || editMode}
            renderInput={(params) => (
              <TextField
                {...params}
                margin="normal"
                label={t('shares.form.owner')}
                placeholder={t('shares.form.ownerPlaceholder')}
                helperText={t('shares.form.ownerHelper')}
              />
            )}
          />

          <Autocomplete
            multiple
            freeSolo
            options={sharedWithSearchResults.map(u => u.username).filter(u => u !== selectedOwner)}
            value={selectedUsers}
            onChange={(_e, newValue) => setSelectedUsers(newValue)}
            onInputChange={(_e, value) => setSharedWithSearchQuery(value)}
            disabled={loading || !selectedOwner}
            renderInput={(params) => (
              <TextField
                {...params}
                margin="normal"
                label={t('shares.form.sharedWith')}
                placeholder={t('shares.form.sharedWithPlaceholder')}
                helperText={t('shares.form.sharedWithHelper')}
              />
            )}
          />

          <TextField
            margin="dense"
            label={t('shares.form.subPath')}
            type="text"
            fullWidth
            variant="outlined"
            value={subPath}
            onChange={(e) => setSubPath(e.target.value)}
            disabled={loading}
            placeholder={t('shares.form.subPathPlaceholder')}
            helperText={t('shares.form.subPathHelper')}
          />

          <TextField
            margin="dense"
            label={t('shares.form.comment')}
            type="text"
            fullWidth
            variant="outlined"
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            disabled={loading}
          />

          <FormControlLabel
            control={
              <Switch
                checked={readOnly}
                onChange={(e) => setReadOnly(e.target.checked)}
                disabled={loading}
              />
            }
            label={t('shares.form.readOnly')}
            sx={{ mt: 2 }}
          />

          {error && (
            <Alert severity="error" sx={{ mt: 2 }}>
              {error}
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog} disabled={loading}>
            {t('shares.form.cancel')}
          </Button>
          <Button
            onClick={editMode ? handleUpdateShare : handleCreateShare}
            variant="contained"
            disabled={loading || !selectedOwner || selectedUsers.length === 0}
          >
            {loading ? t('shares.form.creating') : editMode ? t('shares.form.update') : t('shares.form.create')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
