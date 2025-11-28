# SambaManager

A modern web-based Samba management tool built with Go backend and React + shadcn/ui frontend.

## Features

- 🔐 **Authentication**: Secure login with configurable admin credentials
- 👥 **User Management**: Create and delete Samba users (tdbsam) with automatic home directory management
- 📁 **Share Management**: Manage shared directories with read-only/read-write permissions
- 🔄 **Request Queuing**: Built-in queue system to handle concurrent requests safely
- 🎨 **Modern UI**: Clean and responsive interface using React and Material Design
- 🌐 **Internationalization**: Support for English and Chinese languages
- ⚡ **Single Binary**: Frontend embedded in backend using Go embed
- 🚀 **No System Users**: Samba-only users via tdbsam, directories managed by root

## Screenshots

![image-20251128230643668](https://cdn.itshenryz.com/image-20251128230643668.png)

![image-20251128230707264](https://cdn.itshenryz.com/image-20251128230707264.png)

![image-20251128230719885](https://cdn.itshenryz.com/image-20251128230719885.png)

![image-20251128230748220](https://cdn.itshenryz.com/image-20251128230748220.png)

## Architecture

### Backend (Go)

- **Framework**: Gin web framework
- **Queue System**: Custom worker queue for handling concurrent operations
- **Samba Integration**: Direct `smbpasswd` calls for tdbsam user management, `smb.conf` modification
- **Configuration**: YAML-based configuration
- **Frontend**: Embedded using `//go:embed` directive
- **User Management**: Samba-only users (no Linux system users), root-owned directories with 770 permissions

### Frontend (React + TypeScript)

- **UI Library**: Material Design
- **Build Tool**: Vite
- **State Management**: React hooks
- **API Communication**: Fetch API with Basic Authentication
- **i18n**: react-i18next with English and Chinese translations
- **Deployment**: Embedded in Go binary, served by backend

## Prerequisites

- Go 1.21 or later
- Node.js 18 or later (for development)
- Samba installed on the system
- `libnss-extrausers` package installed
- Root/sudo privileges (for user and Samba management)

## Installation

### 1. Install System Dependencies

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

### 2. Configure Extrausers System

SambaManager uses extrausers to store Samba users separately from system users.

Edit `/etc/nsswitch.conf`:

```bash
sudo nano /etc/nsswitch.conf
```

Modify the passwd and group lines to include `extrausers`:

```
passwd:         files extrausers systemd
group:          files extrausers systemd
```

Create extrausers directory structure:

```bash
sudo mkdir -p /var/lib/extrausers
sudo touch /var/lib/extrausers/passwd
sudo touch /var/lib/extrausers/group
sudo touch /var/lib/extrausers/shadow
sudo chmod 600 /var/lib/extrausers/shadow
sudo chmod 644 /var/lib/extrausers/passwd
sudo chmod 644 /var/lib/extrausers/group
```

### 3. Configure Samba with tdbsam

Backup existing configuration:

```bash
sudo cp /etc/samba/smb.conf /etc/samba/smb.conf.backup
```

Edit Samba configuration:

```bash
sudo nano /etc/samba/smb.conf
```

Ensure the `[global]` section contains these settings:

```ini
[global]
   workgroup = WORKGROUP
   server string = Samba Server
   security = user
   passdb backend = tdbsam
   map to guest = never
   
   # Optional performance tuning
   socket options = TCP_NODELAY IPTOS_LOWDELAY
   read raw = yes
   write raw = yes
   max xmit = 65535
   dead time = 15
   getwd cache = yes
```

Add the `[homes]` section for automatic user home directory access:

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

**Configuration Explanation:**

- `passdb backend = tdbsam`: Uses Samba's internal database for users (not system users)
- `[homes]`: Automatically creates a share for each user's home directory
- `valid users = %S`: Only allows the owner to access their home directory
- `force user/group = root`: All file operations run as root (matching actual file ownership)
- `browseable = no`: Home shares don't appear in network browsing
- `create mask = 0770`: New files get rwxrwx--- permissions

Test the configuration:

```bash
sudo testparm
```

If configuration is valid, restart Samba:

```bash
# Ubuntu/Debian
sudo systemctl restart smbd nmbd
sudo systemctl enable smbd nmbd

# CentOS/RHEL  
sudo systemctl restart smb nmb
sudo systemctl enable smb nmb
```

### 4. Create Samba Home Directory

Create the base directory for all Samba users:

```bash
sudo mkdir -p /home/samba
sudo chown root:root /home/samba
sudo chmod 755 /home/samba
```

### 5. Clone the repository

```bash
git clone https://github.com/itsHenry35/SambaManager.git
cd SambaManager
```

### 6. Configure the application

Create a configuration file from the example:

```bash
cp config.yaml.example config.yaml
```

Edit `config.yaml` to set your admin credentials and preferences:

```yaml
admin:
  username: admin
  password: admin123
  
home_dir: /home/samba

samba:
  config_path: /etc/samba/smb.conf
  
server:
  port: 8080
  host: 0.0.0.0
```

### 7. Build and run

#### Quick Build (Production)

Use the build script to create a single binary with embedded frontend:

```bash
./scripts/build-all.sh
```

This will:

1. Build the React frontend with i18n
2. Embed it in the Go binary using `go:embed`
3. Create `samba-manager` executable

Then run:

```bash
sudo ./samba-manager -config config.yaml
```

Access the application at `http://localhost:8080`

#### Development Mode

For development with hot reload:

**Terminal 1 - Backend:**

```bash
go run main.go -config config.yaml
```

**Terminal 2 - Frontend:**

```bash
cd frontend
pnpm install
pnpm run dev
```

Frontend dev server at `http://localhost:5173`, backend at `http://localhost:8080`

#### Manual Build

```bash
# Build frontend
cd frontend
pnpm install
pnpm run build

# Build backend with embedded frontend
cd ..
go build -o samba-manager
```

**Note:** Root privileges required for Samba user management and directory operations.

## Systemd Service Deployment

For production deployment as a system service:

1. Build the application:

```bash
./scripts/build-all.sh
```

2. Copy files to installation directory:

```bash
sudo mkdir -p /opt/samba-manager
sudo cp samba-manager /opt/samba-manager/
sudo cp config.yaml /opt/samba-manager/
```

3. Install and start the service:

```bash
sudo cp scripts/samba-manager.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable samba-manager
sudo systemctl start samba-manager
```

4. Check status:

```bash
sudo systemctl status samba-manager
```

## Automated Builds

This project uses GitHub Actions for automated building. Every push triggers a build workflow that:

- Builds the Go backend
- Builds the React frontend
- Creates a release package with everything bundled

You can download pre-built artifacts from the [Actions tab](../../actions) in the repository.

## Usage

1. **Login**: Access the web interface at `http://localhost:8080`

   - Enter admin credentials from `config.yaml`
   - Click 🌐 icon in header to switch between English and Chinese
2. **User Management**:

   - Click "User Management" / "用户管理" tab
   - Click "Create User" / "创建用户" to add a new Samba user
   - Provide username and password
   - System creates Samba user (tdbsam) and home directory owned by root:root with 770 permissions
   - **Note**: No Linux system user is created, only Samba user in tdbsam
   - Click trash icon to delete user and their home directory
3. **Share Management**:

   - Click "Share Management" / "共享管理" tab
   - Click "Create Share" / "创建共享" to add a new shared directory
   - Provide share name, path, and permissions
   - Toggle "Read Only" / "只读" for read-only shares
   - System automatically updates `/etc/samba/smb.conf` and sets directory to root:root 770
   - Click trash icon to remove share configuration
4. **Language Toggle**:

   - Click the 🌐 (Languages) icon in the header to switch between:
     - **English (en)** - Default
     - **中文 (zh)** - Chinese## Development
