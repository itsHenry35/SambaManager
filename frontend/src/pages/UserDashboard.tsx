import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Button,
  Card,
  CardContent,
  Container,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
  Chip,
  FormControlLabel,
  Checkbox,
  AppBar,
  Toolbar,
  Autocomplete,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
  Logout as LogoutIcon,
  VpnKey as VpnKeyIcon,
  Language as LanguageIcon,
} from '@mui/icons-material';
import { userShareAPI, userProfileAPI } from '../api';
import { handleResp, handleRespWithNotifySuccess } from '../utils/handleResp';
import type { ShareResponse, CreateMyShareRequest, UpdateShareRequest, UserResponse, ChangeOwnPasswordRequest } from '../types';

export function UserDashboard() {
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const [shares, setShares] = useState<ShareResponse[]>([]);
  const [openCreateDialog, setOpenCreateDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false);
  const [openPasswordDialog, setOpenPasswordDialog] = useState(false);
  const [selectedShare, setSelectedShare] = useState<ShareResponse | null>(null);
  const [formData, setFormData] = useState<CreateMyShareRequest | UpdateShareRequest>({
    shared_with: [],
    read_only: false,
    comment: '',
  });
  const [passwordData, setPasswordData] = useState<ChangeOwnPasswordRequest>({
    old_password: '',
    new_password: '',
  });
  const [sharedWithUsers, setSharedWithUsers] = useState<string[]>([]);
  const [userSearchQuery, setUserSearchQuery] = useState('');
  const [userSearchResults, setUserSearchResults] = useState<UserResponse[]>([]);
  const [shareName, setShareName] = useState('');

  const currentUsername = localStorage.getItem('username') || '';

  const loadShares = async () => {
    const resp = await userShareAPI.getMyShares();
    handleResp(
      resp,
      (data) => {
        setShares(data || []);
      }
    );
  };

  // Search users for autocomplete
  useEffect(() => {
    if (userSearchQuery.length > 0) {
      const searchUsers = async () => {
        const resp = await userProfileAPI.searchUsers(userSearchQuery);
        handleResp(resp, (data) => {
          setUserSearchResults(data || []);
        });
      };
      void searchUsers();
    } else {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setUserSearchResults([]);
    }
  }, [userSearchQuery]);

  useEffect(() => {
    loadShares();
  }, []);

  const handleOpenCreateDialog = () => {
    setFormData({
      shared_with: [],
      read_only: false,
      comment: '',
      sub_path: '',
    });
    setSharedWithUsers([]);
    setShareName('');
    setOpenCreateDialog(true);
  };

  const handleCloseCreateDialog = () => {
    setOpenCreateDialog(false);
  };

  const handleCreateShare = async () => {
    const resp = await userShareAPI.createMyShare({
      ...(formData as CreateMyShareRequest),
      name: shareName || undefined,
    });
    handleRespWithNotifySuccess(
      resp,
      () => {
        handleCloseCreateDialog();
        loadShares();
      }
    );
  };

  const handleOpenEditDialog = (share: ShareResponse) => {
    setSelectedShare(share);
    setFormData({
      shared_with: share.shared_with,
      read_only: share.read_only,
      comment: share.comment,
      sub_path: share.sub_path || '',
    });
    setSharedWithUsers(share.shared_with);
    setShareName(share.id);
    setOpenEditDialog(true);
  };

  const handleCloseEditDialog = () => {
    setOpenEditDialog(false);
    setSelectedShare(null);
  };

  const handleUpdateShare = async () => {
    if (!selectedShare) return;

    const resp = await userShareAPI.updateMyShare(selectedShare.id, formData as UpdateShareRequest);
    handleRespWithNotifySuccess(
      resp,
      () => {
        handleCloseEditDialog();
        loadShares();
      }
    );
  };

  const handleOpenDeleteDialog = (share: ShareResponse) => {
    setSelectedShare(share);
    setOpenDeleteDialog(true);
  };

  const handleCloseDeleteDialog = () => {
    setOpenDeleteDialog(false);
    setSelectedShare(null);
  };

  const handleDeleteShare = async () => {
    if (!selectedShare) return;

    const resp = await userShareAPI.deleteMyShare(selectedShare.id);
    handleRespWithNotifySuccess(
      resp,
      () => {
        handleCloseDeleteDialog();
        loadShares();
      }
    );
  };

  const handleOpenPasswordDialog = () => {
    setPasswordData({
      old_password: '',
      new_password: '',
    });
    setOpenPasswordDialog(true);
  };

  const handleClosePasswordDialog = () => {
    setOpenPasswordDialog(false);
  };

  const handleChangePassword = async () => {
    const resp = await userProfileAPI.changeOwnPassword(passwordData);
    handleRespWithNotifySuccess(
      resp,
      () => {
        handleClosePasswordDialog();
      }
    );
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    localStorage.removeItem('username');
    navigate('/login');
  };

  const toggleLanguage = () => {
    const newLang = i18n.language === 'zh' ? 'en' : 'zh';
    i18n.changeLanguage(newLang);
  };

  const handleSharedWithChange = (_event: React.SyntheticEvent, newValue: string[]) => {
    setSharedWithUsers(newValue);
    setFormData({ ...formData, shared_with: newValue });
  };

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            {t('userDashboard.title')} - {currentUsername}
          </Typography>
          <IconButton color="inherit" onClick={toggleLanguage}>
            <LanguageIcon />
          </IconButton>
          <Button color="inherit" startIcon={<VpnKeyIcon />} onClick={handleOpenPasswordDialog}>
            {t('userDashboard.changePassword')}
          </Button>
          <Button color="inherit" startIcon={<LogoutIcon />} onClick={handleLogout}>
            {t('common.logout')}
          </Button>
        </Toolbar>
      </AppBar>

      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Card>
          <CardContent>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
              <Typography variant="h5">
                {t('userDashboard.myShares')}
              </Typography>
              <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={handleOpenCreateDialog}
              >
                {t('userDashboard.createShare')}
              </Button>
            </Box>

            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>{t('userDashboard.shareId')}</TableCell>
                    <TableCell>{t('userDashboard.path')}</TableCell>
                    <TableCell>{t('userDashboard.subPath')}</TableCell>
                    <TableCell>{t('userDashboard.sharedWith')}</TableCell>
                    <TableCell>{t('userDashboard.permissions')}</TableCell>
                    <TableCell>{t('userDashboard.comment')}</TableCell>
                    <TableCell>{t('common.actions')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {shares.map((share) => (
                    <TableRow key={share.id}>
                      <TableCell>{share.id}</TableCell>
                      <TableCell>{share.path}</TableCell>
                      <TableCell>
                        {share.sub_path ? (
                          <Chip label={share.sub_path} size="small" color="primary" variant="outlined" icon={<span>📁</span>} />
                        ) : (
                          <Typography variant="caption" color="text.secondary">-</Typography>
                        )}
                      </TableCell>
                      <TableCell>
                        {share.shared_with.map(user => (
                          <Chip key={user} label={user} size="small" sx={{ mr: 0.5 }} />
                        ))}
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={share.read_only ? t('userDashboard.readOnly') : t('userDashboard.readWrite')}
                          color={share.read_only ? 'default' : 'primary'}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>{share.comment}</TableCell>
                      <TableCell>
                        <IconButton size="small" onClick={() => handleOpenEditDialog(share)}>
                          <EditIcon />
                        </IconButton>
                        <IconButton size="small" onClick={() => handleOpenDeleteDialog(share)}>
                          <DeleteIcon />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                  {shares.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={7} align="center">
                        {t('userDashboard.noShares')}
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          </CardContent>
        </Card>
      </Container>
      {/* Create Share Dialog */}
      <Dialog open={openCreateDialog} onClose={handleCloseCreateDialog} maxWidth="sm" fullWidth>
        <DialogTitle>{t('userDashboard.createShare')}</DialogTitle>
        <DialogContent>
          <TextField
            margin="dense"
            label={t('shares.form.name')}
            type="text"
            fullWidth
            variant="outlined"
            value={shareName}
            onChange={(e) => setShareName(e.target.value)}
            placeholder={t('shares.form.namePlaceholder')}
            helperText={t('shares.form.nameHelper')}
          />
          <Autocomplete
            multiple
            freeSolo
            options={userSearchResults.map(u => u.username).filter(u => u !== currentUsername)}
            value={sharedWithUsers}
            onChange={handleSharedWithChange}
            onInputChange={(_e, value) => setUserSearchQuery(value)}
            renderInput={(params) => (
              <TextField
                {...params}
                margin="normal"
                label={t('userDashboard.sharedWith')}
                placeholder={t('userDashboard.sharedWithPlaceholder')}
                helperText={t('userDashboard.sharedWithHelper')}
              />
            )}
          />
          <TextField
            fullWidth
            label={t('userDashboard.subPath')}
            value={formData.sub_path || ''}
            onChange={(e) => setFormData({ ...formData, sub_path: e.target.value })}
            margin="normal"
            placeholder={t('userDashboard.subPathPlaceholder')}
            helperText={t('userDashboard.subPathHelper')}
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={formData.read_only}
                onChange={(e) => setFormData({ ...formData, read_only: e.target.checked })}
              />
            }
            label={t('userDashboard.readOnly')}
          />
          <TextField
            fullWidth
            label={t('userDashboard.comment')}
            value={formData.comment}
            onChange={(e) => setFormData({ ...formData, comment: e.target.value })}
            margin="normal"
            multiline
            rows={2}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseCreateDialog}>{t('common.cancel')}</Button>
          <Button onClick={handleCreateShare} variant="contained">
            {t('common.create')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Edit Share Dialog */}
      <Dialog open={openEditDialog} onClose={handleCloseEditDialog} maxWidth="sm" fullWidth>
        <DialogTitle>{t('userDashboard.editShare')}</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label={t('shares.form.name')}
            value={shareName}
            disabled
            margin="normal"
            helperText={t('shares.form.nameHelperEdit')}
          />
          <Autocomplete
            multiple
            freeSolo
            options={userSearchResults.map(u => u.username).filter(u => u !== currentUsername)}
            value={sharedWithUsers}
            onChange={handleSharedWithChange}
            onInputChange={(_e, value) => setUserSearchQuery(value)}
            renderInput={(params) => (
              <TextField
                {...params}
                margin="normal"
                label={t('userDashboard.sharedWith')}
                placeholder={t('userDashboard.sharedWithPlaceholder')}
                helperText={t('userDashboard.sharedWithHelper')}
              />
            )}
          />
          <TextField
            fullWidth
            label={t('userDashboard.subPath')}
            value={formData.sub_path || ''}
            onChange={(e) => setFormData({ ...formData, sub_path: e.target.value })}
            margin="normal"
            placeholder={t('userDashboard.subPathPlaceholder')}
            helperText={t('userDashboard.subPathHelper')}
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={formData.read_only}
                onChange={(e) => setFormData({ ...formData, read_only: e.target.checked })}
              />
            }
            label={t('userDashboard.readOnly')}
          />
          <TextField
            fullWidth
            label={t('userDashboard.comment')}
            value={formData.comment}
            onChange={(e) => setFormData({ ...formData, comment: e.target.value })}
            margin="normal"
            multiline
            rows={2}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseEditDialog}>{t('common.cancel')}</Button>
          <Button onClick={handleUpdateShare} variant="contained">
            {t('common.save')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete Share Dialog */}
      <Dialog open={openDeleteDialog} onClose={handleCloseDeleteDialog}>
        <DialogTitle>{t('userDashboard.deleteShare')}</DialogTitle>
        <DialogContent>
          <Typography>
            {t('userDashboard.deleteConfirm')} "{selectedShare?.id}"?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDeleteDialog}>{t('common.cancel')}</Button>
          <Button onClick={handleDeleteShare} color="error" variant="contained">
            {t('common.delete')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Change Password Dialog */}
      <Dialog open={openPasswordDialog} onClose={handleClosePasswordDialog} maxWidth="sm" fullWidth>
        <DialogTitle>{t('userDashboard.changePassword')}</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            type="password"
            label={t('userDashboard.oldPassword')}
            value={passwordData.old_password}
            onChange={(e) => setPasswordData({ ...passwordData, old_password: e.target.value })}
            margin="normal"
          />
          <TextField
            fullWidth
            type="password"
            label={t('userDashboard.newPassword')}
            value={passwordData.new_password}
            onChange={(e) => setPasswordData({ ...passwordData, new_password: e.target.value })}
            margin="normal"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClosePasswordDialog}>{t('common.cancel')}</Button>
          <Button onClick={handleChangePassword} variant="contained">
            {t('common.save')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
