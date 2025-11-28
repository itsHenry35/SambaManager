import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Button,
  Card,
  CardContent,
  Container,
  IconButton,
  TextField,
  Typography,
  Alert,
} from '@mui/material';
import { Language as LanguageIcon } from '@mui/icons-material';
import { authAPI } from '../api';
import { handleRespWithoutAuthAndNotify } from '../utils/handleResp';

export function Login() {
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  // Check if user is already logged in
  useEffect(() => {
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('role');

    if (token && role) {
      // Redirect based on role
      if (role === 'admin') {
        navigate('/dashboard/user-management', { replace: true });
      } else {
        navigate('/user-dashboard', { replace: true });
      }
    }
  }, [navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    const resp = await authAPI.login({ username, password });
    handleRespWithoutAuthAndNotify(
      resp,
      (data) => {
        // Save token and role to localStorage
        localStorage.setItem('token', data.token);
        localStorage.setItem('role', data.role);
        localStorage.setItem('username', data.username);

        // Navigate based on role
        if (data.role === 'admin') {
          navigate('/dashboard/user-management');
        } else {
          navigate('/user-dashboard');
        }
      },
      (message) => {
        setError(message);
      }
    );
    setLoading(false);
  };

  const toggleLanguage = () => {
    const newLang = i18n.language === 'zh' ? 'en' : 'zh';
    i18n.changeLanguage(newLang);
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: 'grey.100',
        position: 'relative',
      }}
    >
      <IconButton
        onClick={toggleLanguage}
        sx={{
          position: 'absolute',
          top: 16,
          right: 16,
        }}
        title="Switch Language"
      >
        <LanguageIcon />
      </IconButton>
      
      <Container maxWidth="sm">
        <Card elevation={3}>
          <CardContent sx={{ p: 4 }}>
            <Typography variant="h4" component="h1" gutterBottom align="center">
              {t('login.title')}
            </Typography>
            <Typography variant="body2" color="text.secondary" gutterBottom align="center" sx={{ mb: 3 }}>
              {t('login.subtitle')}
            </Typography>
            
            <form onSubmit={handleSubmit}>
              <TextField
                fullWidth
                label={t('login.username')}
                variant="outlined"
                margin="normal"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                disabled={loading}
              />
              <TextField
                fullWidth
                label={t('login.password')}
                type="password"
                variant="outlined"
                margin="normal"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={loading}
              />
              
              {error && (
                <Alert severity="error" sx={{ mt: 2 }}>
                  {error}
                </Alert>
              )}
              
              <Button
                type="submit"
                fullWidth
                variant="contained"
                size="large"
                disabled={loading}
                sx={{ mt: 3 }}
              >
                {loading ? t('login.loggingIn') : t('login.loginButton')}
              </Button>
            </form>
          </CardContent>
        </Card>
      </Container>
    </Box>
  );
}
