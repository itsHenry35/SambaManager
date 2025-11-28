import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Paper,
  Typography,
  Button,
  Alert,
  CircularProgress,
  Card,
  CardContent,
  Chip,
  TextField,
  Snackbar,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Divider,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from '@mui/material';
import {
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Warning as WarningIcon,
  Refresh as RefreshIcon,
  ExpandMore as ExpandMoreIcon,
  Save as SaveIcon,
  Code as CodeIcon,
  Visibility as VisibilityIcon,
} from '@mui/icons-material';
import { systemAPI } from '../api';
import { handleResp, handleRespWithNotifySuccess } from '../utils/handleResp';
import type { CheckResult, SambaGlobalConfig, SambaHomesConfig } from '../types';

export function Settings() {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [checks, setChecks] = useState<CheckResult[]>([]);
  const [globalConfig, setGlobalConfig] = useState<SambaGlobalConfig>({
    workgroup: '',
    server_string: '',
    security: '',
    passdb_backend: '',
    map_to_guest: '',
    access_based_share_enum: '',
  });
  const [homesConfig, setHomesConfig] = useState<SambaHomesConfig>({
    comment: '',
    browseable: '',
    writable: '',
    valid_users: '',
    force_user: '',
    force_group: '',
    create_mask: '',
    directory_mask: '',
  });
  const [snackbar, setSnackbar] = useState<{
    open: boolean;
    message: string;
    severity: 'success' | 'error' | 'warning' | 'info';
  }>({ open: false, message: '', severity: 'info' });
  const [openRawEditor, setOpenRawEditor] = useState(false);
  const [rawConfigContent, setRawConfigContent] = useState('');
  const [rawConfigPath, setRawConfigPath] = useState('');
  const [openStatusDialog, setOpenStatusDialog] = useState(false);
  const [sambaStatus, setSambaStatus] = useState('');

  const showSnackbar = (message: string, severity: 'success' | 'error' | 'warning' | 'info') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleCloseSnackbar = () => {
    setSnackbar({ ...snackbar, open: false });
  };

  const loadEnvironmentCheck = useCallback(async () => {
    setLoading(true);
    const resp = await systemAPI.checkEnvironment();
    handleResp(
      resp,
      (data) => {
        setChecks(data.checks || []);
      },
      (message) => {
        showSnackbar(`Failed to check environment: ${message}`, 'error');
      }
    );
    setLoading(false);
  }, []);

  const loadSambaConfig = useCallback(async () => {
    const resp = await systemAPI.getSambaConfig();
    handleResp(
      resp,
      (data) => {
        setGlobalConfig(data.global);
        setHomesConfig(data.homes);
      },
      (message) => {
        showSnackbar(`Failed to load Samba config: ${message}`, 'error');
      }
    );
  }, []);

  const loadRawConfig = async () => {
    const resp = await systemAPI.getSambaConfigFile();
    handleResp(
      resp,
      (data) => {
        setRawConfigContent(data.content);
        setRawConfigPath(data.path);
      },
      (message) => {
        showSnackbar(`Failed to load raw config: ${message}`, 'error');
      }
    );
  };

  const handleOpenRawEditor = () => {
    loadRawConfig();
    setOpenRawEditor(true);
  };

  const handleCloseRawEditor = () => {
    setOpenRawEditor(false);
  };

  const handleOpenStatus = async () => {
    setOpenStatusDialog(true);
    const resp = await systemAPI.getSambaStatus();
    handleResp(
      resp,
      (data) => {
        setSambaStatus(data.raw_output);
      },
      (message) => {
        showSnackbar(`Failed to get Samba status: ${message}`, 'error');
      }
    );
  };

  const handleCloseStatus = () => {
    setOpenStatusDialog(false);
  };

  const handleSaveRawConfig = async () => {
    const resp = await systemAPI.updateSambaConfigFile({ content: rawConfigContent });
    handleRespWithNotifySuccess(
      resp,
      () => {
        showSnackbar(t('settings.rawConfigSavedSuccess'), 'success');
        handleCloseRawEditor();
        loadSambaConfig();
      },
      (message) => {
        showSnackbar(`${t('settings.failedToSaveRawConfig')}: ${message}`, 'error');
      }
    );
  };

  const handleSaveConfig = async () => {
    const resp = await systemAPI.updateSambaConfig({
      global: globalConfig,
      homes: homesConfig,
    });
    handleRespWithNotifySuccess(
      resp,
      () => {
        showSnackbar(t('settings.configSavedSuccess'), 'success');
        loadSambaConfig();
      },
      (message) => {
        showSnackbar(`${t('settings.failedToSaveConfig')}: ${message}`, 'error');
      }
    );
  };

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadEnvironmentCheck();
    void loadSambaConfig();
  }, [loadEnvironmentCheck, loadSambaConfig]);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pass':
        return <CheckCircleIcon color="success" />;
      case 'fail':
        return <ErrorIcon color="error" />;
      case 'warning':
        return <WarningIcon color="warning" />;
      default:
        return null;
    }
  };

  const getStatusColor = (status: string): 'success' | 'error' | 'warning' | 'default' => {
    switch (status) {
      case 'pass':
        return 'success';
      case 'fail':
        return 'error';
      case 'warning':
        return 'warning';
      default:
        return 'default';
    }
  };

  return (
    <Box>
      <Typography variant="h5" gutterBottom sx={{ mb: 3 }}>
        {t('settings.title')}
      </Typography>

      {/* Environment Check Section */}
      <Paper sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6">
            {t('settings.environmentCheck')}
          </Typography>
          <Button
            startIcon={loading ? <CircularProgress size={20} /> : <RefreshIcon />}
            onClick={loadEnvironmentCheck}
            disabled={loading}
          >
            {t('settings.refresh')}
          </Button>
        </Box>

        {loading && checks.length === 0 ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        ) : (
          <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 2 }}>
            {checks.map((check) => {
              const checkKey = `settings.checks.${check.id}`;
              const statusKey = check.status; // 'pass', 'fail', or 'warning'
              return (
                <Box key={check.id}>
                  <Card variant="outlined">
                    <CardContent>
                      <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                        {getStatusIcon(check.status)}
                        <Typography variant="h6" sx={{ ml: 1, flexGrow: 1 }}>
                          {t(`${checkKey}.name`)}
                        </Typography>
                        <Chip
                          label={check.status.toUpperCase()}
                          color={getStatusColor(check.status)}
                          size="small"
                        />
                      </Box>
                      <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                        {t(`${checkKey}.description`)}
                      </Typography>
                      <Typography variant="body2" sx={{ mb: 1 }}>
                        {t(`${checkKey}.${statusKey}`)}
                      </Typography>
                      {check.status !== 'pass' && (
                        <Alert severity="info" sx={{ mt: 1 }}>
                          <Typography variant="caption">
                            <strong>{t('settings.fix')}:</strong> {t(`${checkKey}.fix`)}
                          </Typography>
                        </Alert>
                      )}
                    </CardContent>
                  </Card>
                </Box>
              );
            })}
          </Box>
        )}
      </Paper>

      {/* Samba Configuration Section */}
      <Paper sx={{ p: 3 }}>
        <Typography variant="h6" gutterBottom>
          {t('settings.sambaConfig')}
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
          {t('settings.sambaConfigDesc')}
        </Typography>

        {/* Global Section */}
        <Accordion defaultExpanded>
          <AccordionSummary expandIcon={<ExpandMoreIcon />}>
            <Typography variant="subtitle1" fontWeight="bold">
              [global] Section
            </Typography>
          </AccordionSummary>
          <AccordionDetails>
            <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 2 }}>
              <Box>
                <TextField
                  fullWidth
                  label="Workgroup"
                  value={globalConfig.workgroup}
                  onChange={(e) => setGlobalConfig({ ...globalConfig, workgroup: e.target.value })}
                  helperText="Windows workgroup name"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Server String"
                  value={globalConfig.server_string}
                  onChange={(e) => setGlobalConfig({ ...globalConfig, server_string: e.target.value })}
                  helperText="Server description"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Security"
                  value={globalConfig.security}
                  onChange={(e) => setGlobalConfig({ ...globalConfig, security: e.target.value })}
                  helperText="Security mode (usually 'user')"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Passdb Backend"
                  value={globalConfig.passdb_backend}
                  onChange={(e) => setGlobalConfig({ ...globalConfig, passdb_backend: e.target.value })}
                  helperText="Password database backend (e.g., 'tdbsam')"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Map to Guest"
                  value={globalConfig.map_to_guest}
                  onChange={(e) => setGlobalConfig({ ...globalConfig, map_to_guest: e.target.value })}
                  helperText="Guest access policy (e.g., 'never')"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Access Based Share Enum"
                  value={globalConfig.access_based_share_enum}
                  onChange={(e) => setGlobalConfig({ ...globalConfig, access_based_share_enum: e.target.value })}
                  helperText="Must be 'yes' (required)"
                />
              </Box>
            </Box>
          </AccordionDetails>
        </Accordion>

        <Divider sx={{ my: 2 }} />

        {/* Homes Section */}
        <Accordion defaultExpanded>
          <AccordionSummary expandIcon={<ExpandMoreIcon />}>
            <Typography variant="subtitle1" fontWeight="bold">
              [homes] Section
            </Typography>
          </AccordionSummary>
          <AccordionDetails>
            <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 2 }}>
              <Box>
                <TextField
                  fullWidth
                  label="Comment"
                  value={homesConfig.comment}
                  onChange={(e) => setHomesConfig({ ...homesConfig, comment: e.target.value })}
                  helperText="Share description"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Browseable"
                  value={homesConfig.browseable}
                  onChange={(e) => setHomesConfig({ ...homesConfig, browseable: e.target.value })}
                  helperText="yes/no - visible in network browse"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Writable"
                  value={homesConfig.writable}
                  onChange={(e) => setHomesConfig({ ...homesConfig, writable: e.target.value })}
                  helperText="yes/no - allow write access"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Valid Users"
                  value={homesConfig.valid_users}
                  onChange={(e) => setHomesConfig({ ...homesConfig, valid_users: e.target.value })}
                  helperText="e.g., '%S' for owner only"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Force User"
                  value={homesConfig.force_user}
                  onChange={(e) => setHomesConfig({ ...homesConfig, force_user: e.target.value })}
                  helperText="Force all operations as this user"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Force Group"
                  value={homesConfig.force_group}
                  onChange={(e) => setHomesConfig({ ...homesConfig, force_group: e.target.value })}
                  helperText="Force all operations as this group"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Create Mask"
                  value={homesConfig.create_mask}
                  onChange={(e) => setHomesConfig({ ...homesConfig, create_mask: e.target.value })}
                  helperText="File permissions (e.g., '0770')"
                />
              </Box>
              <Box>
                <TextField
                  fullWidth
                  label="Directory Mask"
                  value={homesConfig.directory_mask}
                  onChange={(e) => setHomesConfig({ ...homesConfig, directory_mask: e.target.value })}
                  helperText="Directory permissions (e.g., '0770')"
                />
              </Box>
            </Box>
          </AccordionDetails>
        </Accordion>

        <Box sx={{ mt: 3, display: 'flex', justifyContent: 'space-between', gap: 2 }}>
          <Box sx={{ display: 'flex', gap: 2 }}>
            <Button
              variant="outlined"
              startIcon={<CodeIcon />}
              onClick={handleOpenRawEditor}
            >
              {t('settings.rawEditor')}
            </Button>
            <Button
              variant="outlined"
              startIcon={<VisibilityIcon />}
              onClick={handleOpenStatus}
            >
              {t('settings.viewStatus')}
            </Button>
          </Box>
          <Button
            variant="contained"
            startIcon={<SaveIcon />}
            onClick={handleSaveConfig}
          >
            {t('settings.saveConfig')}
          </Button>
        </Box>
      </Paper>

      {/* Raw Config Editor Dialog */}
      <Dialog open={openRawEditor} onClose={handleCloseRawEditor} maxWidth="md" fullWidth>
        <DialogTitle>{t('settings.rawEditorTitle')}</DialogTitle>
        <DialogContent>
          <Alert severity="warning" sx={{ mb: 2 }}>
            {t('settings.rawEditorDesc')}
          </Alert>
          <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
            {rawConfigPath}
          </Typography>
          <TextField
            fullWidth
            multiline
            rows={20}
            value={rawConfigContent}
            onChange={(e) => setRawConfigContent(e.target.value)}
            variant="outlined"
            sx={{ fontFamily: 'monospace', fontSize: '0.875rem' }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseRawEditor}>{t('common.cancel')}</Button>
          <Button onClick={handleSaveRawConfig} variant="contained" startIcon={<SaveIcon />}>
            {t('settings.saveRawConfig')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Samba Status Dialog */}
      <Dialog open={openStatusDialog} onClose={handleCloseStatus} maxWidth="md" fullWidth>
        <DialogTitle>{t('settings.sambaStatus')}</DialogTitle>
        <DialogContent>
          <Alert severity="info" sx={{ mb: 2 }}>
            {t('settings.sambaStatusDesc')}
          </Alert>
          <TextField
            fullWidth
            multiline
            rows={20}
            value={sambaStatus}
            variant="outlined"
            InputProps={{
              readOnly: true,
              sx: { fontFamily: 'monospace', fontSize: '0.875rem' }
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseStatus}>{t('common.close')}</Button>
          <Button
            onClick={handleOpenStatus}
            variant="contained"
            startIcon={<RefreshIcon />}
          >
            {t('settings.refresh')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert onClose={handleCloseSnackbar} severity={snackbar.severity} sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
