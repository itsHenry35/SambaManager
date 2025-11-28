import { useNavigate, useLocation, Routes, Route, Navigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  AppBar,
  Box,
  Container,
  IconButton,
  Tab,
  Tabs,
  Toolbar,
  Typography,
  Button,
} from '@mui/material';
import {
  People as PeopleIcon,
  Folder as FolderIcon,
  Settings as SettingsIcon,
  Language as LanguageIcon,
  Logout as LogoutIcon,
} from '@mui/icons-material';
import { UserManagement } from './UserManagement';
import { ShareManagement } from './ShareManagement';
import { Settings } from './Settings';

export function Dashboard() {
  const navigate = useNavigate();
  const location = useLocation();
  const { t, i18n } = useTranslation();

  // Determine active tab based on current path
  const getActiveTab = () => {
    const path = location.pathname;
    if (path.includes('/user-management')) return 0;
    if (path.includes('/share-management')) return 1;
    if (path.includes('/settings')) return 2;
    return 0;
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
    window.location.reload();
  };

  const toggleLanguage = () => {
    const newLang = i18n.language === 'zh' ? 'en' : 'zh';
    i18n.changeLanguage(newLang);
  };

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    const paths = ['/dashboard/user-management', '/dashboard/share-management', '/dashboard/settings'];
    navigate(paths[newValue]);
  };

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'grey.50' }}>
      <AppBar position="static" elevation={1}>
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            {t('app.title')}
          </Typography>
          <IconButton color="inherit" onClick={toggleLanguage} title="Switch Language">
            <LanguageIcon />
          </IconButton>
          <Button
            color="inherit"
            startIcon={<LogoutIcon />}
            onClick={handleLogout}
          >
            {t('app.logout')}
          </Button>
        </Toolbar>
      </AppBar>

      <Container maxWidth="lg" sx={{ mt: 3 }}>
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tabs value={getActiveTab()} onChange={handleTabChange} aria-label="dashboard tabs">
            <Tab
              icon={<PeopleIcon />}
              label={t('tabs.userManagement')}
              iconPosition="start"
            />
            <Tab
              icon={<FolderIcon />}
              label={t('tabs.shareManagement')}
              iconPosition="start"
            />
            <Tab
              icon={<SettingsIcon />}
              label={t('tabs.settings')}
              iconPosition="start"
            />
          </Tabs>
        </Box>

        <Box sx={{ py: 3 }}>
          <Routes>
            <Route path="user-management" element={<UserManagement />} />
            <Route path="share-management" element={<ShareManagement />} />
            <Route path="settings" element={<Settings />} />
            <Route path="/" element={<Navigate to="/dashboard/user-management" replace />} />
          </Routes>
        </Box>
      </Container>
    </Box>
  );
}
