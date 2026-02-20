#!/bin/bash
# SynBridge — M0 Server Setup Script
# Target: AlmaLinux 9 (dnf), IONOS VPS 87.106.213.239
# Run as root on a fresh VPS.
# Security review: Codex
#
# What this does:
#   1. System update + base packages
#   2. Create non-root 'synbridge' user
#   3. SSH hardening (key-only, no root login, no password)
#   4. Firewall (firewalld — AlmaLinux 9 default, not ufw)
#   5. fail2ban
#   6. Go 1.23 (via official tarball)
#   7. PostgreSQL 16 (via pgdg repo)
#   8. nginx
#   9. certbot / Let's Encrypt
#   10. App directory + systemd service skeleton
#
# Usage:
#   curl -O https://raw.githubusercontent.com/BioAILogic/agentbridge/main/scripts/setup.sh
#   chmod +x setup.sh
#   ./setup.sh
#
# After running:
#   - SSH back in as 'synbridge' user with your key
#   - Run certbot to get TLS certificate (after DNS is live)
#   - Deploy the Go binary to /opt/synbridge/bin/

set -euo pipefail

# --- Config ---
APP_USER="synbridge"
APP_DIR="/opt/synbridge"
GO_VERSION="1.23.6"
DOMAIN="synbridge.eu"

# Colours for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[setup]${NC} $*"; }
warn() { echo -e "${YELLOW}[warn]${NC} $*"; }
err()  { echo -e "${RED}[error]${NC} $*" >&2; exit 1; }

[[ $EUID -ne 0 ]] && err "Run as root"

# --- 1. System update ---
log "Updating system packages..."
dnf update -y
dnf install -y \
    curl wget tar git \
    vim-minimal \
    policycoreutils-python-utils \
    bash-completion

# --- 2. Create app user ---
log "Creating user: $APP_USER"
if ! id "$APP_USER" &>/dev/null; then
    useradd -m -s /bin/bash "$APP_USER"
fi

# SSH authorized_keys — paste your public key here OR the script will print a reminder
APP_HOME="/home/$APP_USER"
mkdir -p "$APP_HOME/.ssh"
chmod 700 "$APP_HOME/.ssh"
touch "$APP_HOME/.ssh/authorized_keys"
chmod 600 "$APP_HOME/.ssh/authorized_keys"
chown -R "$APP_USER:$APP_USER" "$APP_HOME/.ssh"

if [[ ! -s "$APP_HOME/.ssh/authorized_keys" ]]; then
    warn "No SSH key in $APP_HOME/.ssh/authorized_keys"
    warn "Add your public key before hardening SSH or you will be locked out!"
    warn "Run: echo 'YOUR_PUBLIC_KEY' >> $APP_HOME/.ssh/authorized_keys"
fi

# --- 3. SSH hardening ---
log "Hardening SSH..."
SSHD_CONFIG="/etc/ssh/sshd_config"
cp "$SSHD_CONFIG" "${SSHD_CONFIG}.bak.$(date +%Y%m%d%H%M%S)"

# Apply hardening settings
sed -i 's/^#*PermitRootLogin.*/PermitRootLogin no/' "$SSHD_CONFIG"
sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication no/' "$SSHD_CONFIG"
sed -i 's/^#*PubkeyAuthentication.*/PubkeyAuthentication yes/' "$SSHD_CONFIG"
sed -i 's/^#*X11Forwarding.*/X11Forwarding no/' "$SSHD_CONFIG"

# Ensure settings exist if not already present
grep -q "^PermitRootLogin"       "$SSHD_CONFIG" || echo "PermitRootLogin no"       >> "$SSHD_CONFIG"
grep -q "^PasswordAuthentication" "$SSHD_CONFIG" || echo "PasswordAuthentication no" >> "$SSHD_CONFIG"

systemctl reload sshd
warn "SSH hardened. Root login and password auth disabled."
warn "Ensure '$APP_USER' has a valid authorized_keys before logging out!"

# --- 4. Firewall (firewalld — AlmaLinux 9 default) ---
log "Configuring firewall (firewalld)..."
dnf install -y firewalld
systemctl enable --now firewalld

# Reset to only SSH + HTTP + HTTPS
firewall-cmd --set-default-zone=drop
firewall-cmd --zone=drop --add-service=ssh --permanent
firewall-cmd --zone=drop --add-service=http --permanent
firewall-cmd --zone=drop --add-service=https --permanent
firewall-cmd --reload

log "Firewall: drop default zone, allowing ssh/http/https only"

# --- 5. fail2ban ---
log "Installing fail2ban..."
dnf install -y epel-release
dnf install -y fail2ban

cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime  = 3600
findtime = 600
maxretry = 5
backend  = systemd

[sshd]
enabled = true
port    = ssh
EOF

systemctl enable --now fail2ban
log "fail2ban enabled (SSH: 5 retries, 1h ban)"

# --- 6. Go ---
log "Installing Go $GO_VERSION..."
GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="https://go.dev/dl/${GO_TARBALL}"
GO_SHA256_URL="${GO_URL}.sha256"

cd /tmp
curl -LO "$GO_URL"

# Verify checksum
EXPECTED_SHA=$(curl -sL "$GO_SHA256_URL")
ACTUAL_SHA=$(sha256sum "$GO_TARBALL" | awk '{print $1}')
if [[ "$EXPECTED_SHA" != "$ACTUAL_SHA" ]]; then
    err "Go tarball checksum mismatch. Expected: $EXPECTED_SHA Got: $ACTUAL_SHA"
fi
log "Go checksum verified."

rm -rf /usr/local/go
tar -C /usr/local -xzf "$GO_TARBALL"
rm "$GO_TARBALL"

# PATH for all users
cat > /etc/profile.d/go.sh << 'EOF'
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/home/synbridge/go
EOF

# PATH for synbridge user's shell
echo 'export PATH=$PATH:/usr/local/go/bin' >> "$APP_HOME/.bash_profile"
echo 'export GOPATH=$HOME/go'              >> "$APP_HOME/.bash_profile"

log "Go installed: $(/usr/local/go/bin/go version)"

# --- 7. PostgreSQL 16 ---
log "Installing PostgreSQL 16..."
dnf install -y "https://download.postgresql.org/pub/repos/yum/reporpms/EL-9-x86_64/pgdg-redhat-repo-latest.noarch.rpm"
dnf -qy module disable postgresql   # disable AppStream module to avoid conflict
dnf install -y postgresql16-server postgresql16-contrib

/usr/pgsql-16/bin/postgresql-16-setup initdb
systemctl enable --now postgresql-16

# Create database + user
log "Creating PostgreSQL database..."
sudo -u postgres /usr/pgsql-16/bin/psql << 'PSQL'
CREATE USER synbridge WITH PASSWORD 'changeme_in_production';
CREATE DATABASE synbridge OWNER synbridge;
\q
PSQL

warn "PostgreSQL: default password is 'changeme_in_production' — change it!"
warn "Edit /var/lib/pgsql/16/data/pg_hba.conf if remote access is needed."

# Add pgsql-16 to PATH
echo 'export PATH=$PATH:/usr/pgsql-16/bin' >> "$APP_HOME/.bash_profile"
cat >> /etc/profile.d/go.sh << 'EOF'
export PATH=$PATH:/usr/pgsql-16/bin
EOF

# --- 8. nginx ---
log "Installing nginx..."
dnf install -y nginx
systemctl enable nginx

# Basic nginx config — reverse proxy to Go app on :8080
cat > /etc/nginx/conf.d/synbridge.conf << EOF
server {
    listen 80;
    server_name ${DOMAIN} www.${DOMAIN};

    # Let's Encrypt ACME challenge
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    # Redirect to HTTPS once cert is obtained
    # Uncomment after running certbot:
    # return 301 https://\$host\$request_uri;

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header   Host             \$host;
        proxy_set_header   X-Real-IP        \$remote_addr;
        proxy_set_header   X-Forwarded-For  \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
    }
}
EOF

# Remove default config
rm -f /etc/nginx/conf.d/default.conf

# ACME challenge directory
mkdir -p /var/www/certbot
chown nginx:nginx /var/www/certbot

# SELinux: allow nginx to connect to Go backend
setsebool -P httpd_can_network_connect 1

systemctl start nginx
log "nginx configured for $DOMAIN → :8080"

# --- 9. certbot ---
log "Installing certbot..."
dnf install -y certbot python3-certbot-nginx

cat << EOF

${YELLOW}[certbot]${NC} TLS certificate not yet issued.

Once DNS for $DOMAIN points to this server (87.106.213.239), run:

    certbot --nginx -d $DOMAIN -d www.$DOMAIN

Then uncomment the 'return 301' redirect in /etc/nginx/conf.d/synbridge.conf.

Auto-renewal timer is already installed by certbot package.
EOF

# --- 10. App directory + systemd service ---
log "Creating app directory: $APP_DIR"
mkdir -p "$APP_DIR/bin"
mkdir -p "$APP_DIR/logs"
chown -R "$APP_USER:$APP_USER" "$APP_DIR"

cat > /etc/systemd/system/synbridge.service << EOF
[Unit]
Description=SynBridge Forum
After=network.target postgresql-16.service
Requires=postgresql-16.service

[Service]
Type=simple
User=$APP_USER
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/bin/synbridge
Restart=on-failure
RestartSec=5s
StandardOutput=append:$APP_DIR/logs/synbridge.log
StandardError=append:$APP_DIR/logs/synbridge.log

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=$APP_DIR

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
log "systemd service registered (not started — no binary yet)"

# --- Done ---
cat << EOF

${GREEN}=== M0 Setup Complete ===${NC}

What was installed:
  - Go ${GO_VERSION}         → /usr/local/go/bin/go
  - PostgreSQL 16    → systemd: postgresql-16
  - nginx            → systemd: nginx
  - certbot          → manual step required (see above)
  - fail2ban         → systemd: fail2ban
  - firewalld        → ports 22, 80, 443 open

App user:    $APP_USER
App dir:     $APP_DIR
Service:     synbridge.service (not started)
DB name:     synbridge
DB user:     synbridge
DB password: changeme_in_production ← CHANGE THIS

${YELLOW}Required manual steps:${NC}
  1. Add SSH public key for $APP_USER:
     echo 'YOUR_PUBLIC_KEY' >> $APP_HOME/.ssh/authorized_keys

  2. Change PostgreSQL password:
     sudo -u postgres psql -c "ALTER USER synbridge PASSWORD 'your_strong_password';"

  3. Point DNS for $DOMAIN to 87.106.213.239

  4. Get TLS certificate:
     certbot --nginx -d $DOMAIN -d www.$DOMAIN

  5. Build and deploy Go binary to $APP_DIR/bin/synbridge

  6. Start service:
     systemctl start synbridge

${GREEN}M0 complete. M1 (skeleton app) can begin.${NC}
EOF
