# Quick Start Guide

## Prerequisites Setup

### 1. Install Required Packages

**Ubuntu/Debian:**

```bash
sudo apt-get update
sudo apt-get install samba samba-common-bin libnss-extrausers
```

**CentOS/RHEL:**

```bash
sudo yum install samba samba-common libnss-extrausers
```

**Arch Linux:**

```bash
sudo pacman -S samba
# For extrausers, you may need to install from AUR
```

### 2. Configure NSS for Extrausers

Edit `/etc/nsswitch.conf` and modify the passwd and group lines:

```bash
sudo nano /etc/nsswitch.conf
```

Change:

```
passwd:         files systemd
group:          files systemd
```

To:

```
passwd:         files extrausers systemd
group:          files extrausers systemd
```

Create the extrausers directory:

```bash
sudo mkdir -p /var/lib/extrausers
sudo touch /var/lib/extrausers/passwd
sudo touch /var/lib/extrausers/group
sudo touch /var/lib/extrausers/shadow
sudo chmod 600 /var/lib/extrausers/shadow
sudo chmod 644 /var/lib/extrausers/passwd
sudo chmod 644 /var/lib/extrausers/group
```

### 3. Configure Samba

Backup your original Samba configuration:

```bash
sudo cp /etc/samba/smb.conf /etc/samba/smb.conf.backup
```

Edit Samba configuration:

```bash
sudo nano /etc/samba/smb.conf
```

Ensure the `[global]` section contains:

```ini
[global]
   workgroup = WORKGROUP
   server string = Samba Server
   security = user
   passdb backend = tdbsam
   map to guest = never
```

Add the `[homes]` section for automatic user home directory sharing:

```ini
[homes]
   comment = Home Directories
   browseable = no
   writable = yes
   valid users = %S
   force user = root
   force group = root
   create mask = 0770
   directory mask = 0770
```

**Important:** The `passdb backend = tdbsam` is required for Samba user management without system users. The `[homes]` section allows users to access their own home directories automatically.

Test the configuration:

```bash
sudo testparm
```

Restart Samba service:

```bash
# Ubuntu/Debian
sudo systemctl restart smbd nmbd

# CentOS/RHEL
sudo systemctl restart smb nmb
```

Enable Samba to start on boot:

```bash
# Ubuntu/Debian
sudo systemctl enable smbd nmbd

# CentOS/RHEL
sudo systemctl enable smb nmb
```

### 4. Create Samba Home Directory

Create the base directory for Samba users:

```bash
sudo mkdir -p /home/samba
sudo chown root:root /home/samba
sudo chmod 755 /home/samba
```

## For Development

1. **Setup configuration:**

   ```bash
   cd backend
   cp config.yaml.example config.yaml
   # Edit config.yaml with your settings
   ```
2. **Start backend (terminal 1):**

   ```bash
   cd backend
   go run main.go -config config.yaml
   ```
3. **Start frontend (terminal 2):**

   ```bash
   cd frontend
   pnpm install
   pnpm run dev
   ```
4. **Access the application:**

   - Open `http://localhost:5173`
   - Login with credentials from `config.yaml`

## For Production

1. **Build everything:**

   ```bash
   ./scripts/build-all.sh
   ```
2. **Run the backend:**

   ```bash
   cd backend
   sudo ./samba-manager -config config.yaml
   ```
3. **Access the application:**

   - Open `http://localhost:8080`

## Using Pre-built Binaries

Download the latest build from [GitHub Actions](../../actions):

1. Go to the Actions tab
2. Click on the latest successful build
3. Download the `samba-manager-release` artifact
4. Extract and run:
   ```bash
   tar -xzf samba-manager-release.tar.gz
   cp config.yaml.example config.yaml
   # Edit config.yaml with your settings
   sudo ./samba-manager -config config.yaml
   ```

## Default Credentials

Check your `config.yaml` file for the admin username and password.

**⚠️ IMPORTANT:** Change the default password before deploying to production!

## Troubleshooting

### Permission Denied

The backend needs root privileges to manage system users and Samba:

```bash
sudo ./samba-manager -config config.yaml
```

### Port Already in Use

Change the port in `config.yaml`:

```yaml
server:
  port: 8081  # Change to available port
```

### Samba Not Installed

Install Samba on your system:

```bash
# Ubuntu/Debian
sudo apt-get install samba samba-common-bin

# CentOS/RHEL
sudo yum install samba samba-common

# Arch Linux
sudo pacman -S samba
```

### User Creation Fails

Check that:

1. `/var/lib/extrausers/` directory exists and has correct permissions
2. `/etc/nsswitch.conf` includes `extrausers` in passwd and group lines
3. `libnss-extrausers` package is installed
4. Samba is configured with `passdb backend = tdbsam`

Verify extrausers setup:

```bash
ls -la /var/lib/extrausers/
cat /etc/nsswitch.conf | grep -E "^passwd:|^group:"
```

### Cannot Access Shares

1. Check Samba service is running:

   ```bash
   sudo systemctl status smbd
   ```
2. Verify `[homes]` section exists in `/etc/samba/smb.conf`
3. Test Samba configuration:

   ```bash
   sudo testparm
   ```
4. Check firewall allows Samba ports (139, 445):

   ```bash
   sudo ufw allow samba  # Ubuntu/Debian
   sudo firewall-cmd --permanent --add-service=samba  # CentOS/RHEL
   ```

### Invalid Username Error

If you see "invalid user name" errors, ensure the username:

- Contains only letters, numbers, underscore, and dash
- Is 1-32 characters long
- The application automatically adds `--badname` flag for non-standard usernames

## Usage Tips

### Creating Users

1. Click "User Management" tab
2. Click "Create User" button
3. Enter username and password
4. User's home directory is automatically created at `/home/samba/{username}`

### Creating Shares

1. Click "Share Management" tab
2. Click "Create Share" button
3. Fill in:
   - **Name**: Display name for the share
   - **Path**: Absolute path (e.g., `/data/shared`)
   - **Comment**: Optional description
   - **Read Only**: Check for read-only access

### Deleting Resources

- Click the trash icon next to any user or share to delete it
- Deleting a user also removes their home directory
- Deleting a share only removes the configuration, not the files

## Security Notes

- Always use HTTPS in production (use a reverse proxy like nginx)
- Change default admin password immediately
- Restrict network access to trusted IPs
- Regular backups of `/etc/samba/smb.conf` and user directories
- Keep the system and Samba up to date

## Next Steps

- Configure firewall rules for Samba ports (139, 445)
- Set up automatic backups
- Configure SSL/TLS with reverse proxy
- Add additional Samba share options as needed
