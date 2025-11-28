import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  IconButton,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
  TextField,
  Typography,
  Collapse,
  Alert,
} from '@mui/material';
import {
  Delete as DeleteIcon,
  PersonAdd as PersonAddIcon,
  Lock as LockIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
  FolderOff as FolderOffIcon,
} from '@mui/icons-material';
import { userAPI } from '../api';
import { handleResp, handleRespWithNotifySuccess } from '../utils/handleResp';
import type { UserResponse, OrphanedDirectory, DeleteUserRequest } from '../types';

export function UserManagement() {
  const { t } = useTranslation();
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [orphanedDirs, setOrphanedDirs] = useState<OrphanedDirectory[]>([]);
  const [loading, setLoading] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [openPasswordDialog, setOpenPasswordDialog] = useState(false);
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false);
  const [newUsername, setNewUsername] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [selectedUser, setSelectedUser] = useState('');
  const [changePasswordValue, setChangePasswordValue] = useState('');
  const [deleteHomeDir, setDeleteHomeDir] = useState(true);
  const [showOrphaned, setShowOrphaned] = useState(false);

  // Pagination and search
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(20);
  const [search, setSearch] = useState('');
  const [total, setTotal] = useState(0);

  const loadUsers = async () => {
    const resp = await userAPI.getUsers(page + 1, rowsPerPage, search);
    handleResp(resp, (data, pagination) => {
      setUsers(data || []);
      setTotal(pagination?.total || 0);
    });
  };

  const loadOrphanedDirs = async () => {
    const resp = await userAPI.getOrphanedDirectories();
    handleResp(resp, (data) => {
      setOrphanedDirs(data || []);
    });
  };

  useEffect(() => {
    loadUsers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, rowsPerPage, search]);

  useEffect(() => {
    if (showOrphaned) {
      loadOrphanedDirs();
    }
  }, [showOrphaned]);

  const handleCreateUser = async () => {
    setLoading(true);
    const resp = await userAPI.createUser({
      username: newUsername,
      password: newPassword,
    });
    handleRespWithNotifySuccess(resp, () => {
      setNewUsername('');
      setNewPassword('');
      setOpenDialog(false);
      loadUsers();
    });
    setLoading(false);
  };

  const handleOpenDeleteDialog = (username: string) => {
    setSelectedUser(username);
    setDeleteHomeDir(true);
    setOpenDeleteDialog(true);
  };

  const handleDeleteUser = async () => {
    setLoading(true);
    const deleteReq: DeleteUserRequest = {
      delete_home_dir: deleteHomeDir,
    };
    const resp = await userAPI.deleteUser(selectedUser, deleteReq);
    handleRespWithNotifySuccess(resp, () => {
      setOpenDeleteDialog(false);
      loadUsers();
      if (showOrphaned) {
        loadOrphanedDirs();
      }
    });
    setLoading(false);
  };

  const handleDeleteOrphanedDir = async (dirName: string) => {
    if (!confirm(t('users.deleteOrphanedConfirm', `Delete orphaned directory "${dirName}"?`))) {
      return;
    }
    const resp = await userAPI.deleteOrphanedDirectory(dirName);
    handleRespWithNotifySuccess(resp, () => {
      loadOrphanedDirs();
    });
  };

  const handleOpenPasswordDialog = (username: string) => {
    setSelectedUser(username);
    setChangePasswordValue('');
    setOpenPasswordDialog(true);
  };

  const handleChangePassword = async () => {
    setLoading(true);
    const resp = await userAPI.changePassword(selectedUser, {
      password: changePasswordValue,
    });
    handleRespWithNotifySuccess(resp, () => {
      setOpenPasswordDialog(false);
      setChangePasswordValue('');
    });
    setLoading(false);
  };

  const handleChangePage = (_event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const handleSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(event.target.value);
    setPage(0); // Reset to first page on search
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  return (
    <Box>
      <Card>
        <CardHeader
          title={t('users.title')}
          subheader={t('users.subtitle')}
          action={
            <Button
              variant="contained"
              startIcon={<PersonAddIcon />}
              onClick={() => setOpenDialog(true)}
            >
              {t('users.createUser')}
            </Button>
          }
        />
        <CardContent>
          {/* Search bar */}
          <TextField
            fullWidth
            label={t('common.search')}
            variant="outlined"
            value={search}
            onChange={handleSearchChange}
            sx={{ mb: 2 }}
            placeholder={t('users.searchPlaceholder')}
          />

          {/* Users table */}
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>{t('users.form.username')}</TableCell>
                  <TableCell>{t('users.homeDir')}</TableCell>
                  <TableCell align="right">{t('common.actions')}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {users.map((user) => (
                  <TableRow key={user.username}>
                    <TableCell>{user.username}</TableCell>
                    <TableCell>{user.home_dir}</TableCell>
                    <TableCell align="right">
                      <IconButton
                        size="small"
                        onClick={() => handleOpenPasswordDialog(user.username)}
                        title={t('users.changePassword')}
                      >
                        <LockIcon />
                      </IconButton>
                      <IconButton
                        size="small"
                        onClick={() => handleOpenDeleteDialog(user.username)}
                        title={t('common.delete')}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
                {users.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={3} align="center">
                      {search
                        ? t('users.noSearchResults')
                        : t('users.noUsers')}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>

          {/* Pagination */}
          <TablePagination
            component="div"
            count={total}
            page={page}
            onPageChange={handleChangePage}
            rowsPerPage={rowsPerPage}
            onRowsPerPageChange={handleChangeRowsPerPage}
            rowsPerPageOptions={[10, 20, 50, 100]}
            labelRowsPerPage={t('common.rowsPerPage')}
          />

          {/* Orphaned directories section */}
          <Box sx={{ mt: 3 }}>
            <Button
              onClick={() => setShowOrphaned(!showOrphaned)}
              startIcon={showOrphaned ? <ExpandLessIcon /> : <ExpandMoreIcon />}
              endIcon={<FolderOffIcon />}
            >
              {t('users.orphanedDirs')}
              {orphanedDirs.length > 0 && (
                <Chip label={orphanedDirs.length} size="small" sx={{ ml: 1 }} color="warning" />
              )}
            </Button>
            <Collapse in={showOrphaned}>
              <Alert severity="info" sx={{ mt: 2, mb: 2 }}>
                {t('users.orphanedDirsInfo')}
              </Alert>
              {orphanedDirs.length === 0 ? (
                <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                  {t('users.noOrphanedDirs')}
                </Typography>
              ) : (
                <TableContainer component={Paper} sx={{ mt: 2 }}>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>{t('users.dirName')}</TableCell>
                        <TableCell>{t('users.dirPath')}</TableCell>
                        <TableCell>{t('users.dirSize')}</TableCell>
                        <TableCell align="right">{t('common.actions')}</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {orphanedDirs.map((dir) => (
                        <TableRow key={dir.name}>
                          <TableCell>{dir.name}</TableCell>
                          <TableCell>{dir.path}</TableCell>
                          <TableCell>{formatBytes(dir.size)}</TableCell>
                          <TableCell align="right">
                            <IconButton
                              size="small"
                              onClick={() => handleDeleteOrphanedDir(dir.name)}
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
            </Collapse>
          </Box>
        </CardContent>
      </Card>

      {/* Create User Dialog */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t('users.createUser')}</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label={t('users.form.username')}
            fullWidth
            value={newUsername}
            onChange={(e) => setNewUsername(e.target.value)}
          />
          <TextField
            margin="dense"
            label={t('users.form.password')}
            type="password"
            fullWidth
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>{t('users.form.cancel')}</Button>
          <Button onClick={handleCreateUser} variant="contained" disabled={loading}>
            {t('users.form.create')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete User Dialog */}
      <Dialog open={openDeleteDialog} onClose={() => setOpenDeleteDialog(false)}>
        <DialogTitle>{t('users.deleteUser')}</DialogTitle>
        <DialogContent>
          <Typography>
            {t('users.deleteUserConfirm', `Delete user "${selectedUser}"?`)}
          </Typography>
          <FormControlLabel
            control={
              <Checkbox
                checked={deleteHomeDir}
                onChange={(e) => setDeleteHomeDir(e.target.checked)}
              />
            }
            label={t('users.deleteHomeDir')}
            sx={{ mt: 2 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDeleteDialog(false)}>{t('common.cancel')}</Button>
          <Button onClick={handleDeleteUser} color="error" variant="contained" disabled={loading}>
            {t('common.delete')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Change Password Dialog */}
      <Dialog open={openPasswordDialog} onClose={() => setOpenPasswordDialog(false)}>
        <DialogTitle>{t('users.changePassword')}</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label={t('users.form.newPassword')}
            type="password"
            fullWidth
            value={changePasswordValue}
            onChange={(e) => setChangePasswordValue(e.target.value)}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenPasswordDialog(false)}>{t('common.cancel')}</Button>
          <Button onClick={handleChangePassword} variant="contained" disabled={loading}>
            {t('users.form.change')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
