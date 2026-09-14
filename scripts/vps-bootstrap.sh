#!/usr/bin/env bash
# Prepare an Ubuntu/Debian OVH VPS for the GATI GitHub Actions deploy user.
# Run locally on the VPS as root. This script does not install GATI or handle
# Cloudflare/SMTP/VAPID secrets.
set -euo pipefail

if [[ "$(id -u)" -ne 0 ]]; then
  printf 'Run this script as root on the VPS.\n' >&2
  exit 1
fi

if [[ ! -r /etc/os-release ]]; then
  printf 'Cannot identify the VPS operating system.\n' >&2
  exit 1
fi
# shellcheck disable=SC1091
source /etc/os-release
case "${ID:-}" in
  ubuntu|debian) ;;
  *) printf 'Supported systems: Ubuntu or Debian (detected %s).\n' "${ID:-unknown}" >&2; exit 1 ;;
esac

deploy_user="${GATI_DEPLOY_USER:-gati-deploy}"
if [[ ! "$deploy_user" =~ ^[a-z_][a-z0-9_-]{0,31}$ ]]; then
  printf 'Invalid GATI_DEPLOY_USER.\n' >&2
  exit 1
fi

printf 'Preparing %s for GATI deployment user %s.\n' "${PRETTY_NAME:-$ID}" "$deploy_user"
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y --no-install-recommends ca-certificates curl gnupg openssh-server python3 tar gzip xz-utils

# Docker's repository supplies the engine and Compose plugin. The key is stored
# in a root-readable keyring rather than installed into the legacy apt keychain.
install -d -m 0755 /etc/apt/keyrings
curl --fail --location --proto '=https' --tlsv1.2 \
  "https://download.docker.com/linux/${ID}/gpg" \
  | gpg --dearmor --yes -o /etc/apt/keyrings/docker.gpg
chmod 0644 /etc/apt/keyrings/docker.gpg
arch="$(dpkg --print-architecture)"
codename="${VERSION_CODENAME:-}"
if [[ -z "$codename" ]]; then
  codename="$(. /etc/os-release; printf '%s' "${VERSION_CODENAME:-}")"
fi
if [[ -z "$codename" ]]; then
  printf 'Could not determine the Docker repository codename.\n' >&2
  exit 1
fi
printf 'deb [arch=%s signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/%s %s stable\n' \
  "$arch" "$ID" "$codename" > /etc/apt/sources.list.d/docker.list
apt-get update
apt-get install -y --no-install-recommends docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
systemctl enable --now docker

if ! getent group docker >/dev/null; then
  groupadd --system docker
fi
if ! id "$deploy_user" >/dev/null 2>&1; then
  useradd --create-home --shell /bin/bash --comment 'GATI deployment user' "$deploy_user"
fi
usermod --append --groups docker "$deploy_user"

install -d -m 0755 -o "$deploy_user" -g "$deploy_user" /opt/gati
install -d -m 0700 -o "$deploy_user" -g "$deploy_user" /opt/gati/releases
install -d -m 0700 -o "$deploy_user" -g "$deploy_user" /opt/gati/audits
install -d -m 0700 /opt/gati/releases/.gati-production-runtime
chown "$deploy_user":"$deploy_user" /opt/gati/releases/.gati-production-runtime

ssh_dir="$(getent passwd "$deploy_user" | cut -d: -f6)/.ssh"
install -d -m 0700 -o "$deploy_user" -g "$deploy_user" "$ssh_dir"
authorized_keys="$ssh_dir/authorized_keys"
touch "$authorized_keys"
chown "$deploy_user":"$deploy_user" "$authorized_keys"
chmod 0600 "$authorized_keys"
if [[ -z "${GATI_SSH_PUBLIC_KEY:-}" ]]; then
  printf 'Paste the deployment public key (one line), then press Enter:\n'
  read -r GATI_SSH_PUBLIC_KEY
fi
if [[ ! "$GATI_SSH_PUBLIC_KEY" =~ ^(ssh-ed25519|ssh-rsa|ecdsa-sha2-nistp(256|384|521))[[:space:]]+[A-Za-z0-9+/]+={0,3}([[:space:]]+.*)?$ ]]; then
  printf 'The supplied value does not look like an OpenSSH public key.\n' >&2
  exit 1
fi
if ! grep -Fqx -- "$GATI_SSH_PUBLIC_KEY" "$authorized_keys"; then
  printf '%s\n' "$GATI_SSH_PUBLIC_KEY" >> "$authorized_keys"
fi

# The deployment preflight rejects active swap. Preserve a dated backup and
# comment only swap entries so a reboot cannot silently re-enable them.
if [[ -s /proc/swaps && "$(wc -l < /proc/swaps)" -gt 1 ]]; then
  swapoff -a
  backup="/etc/fstab.gati-backup.$(date -u +%Y%m%dT%H%M%SZ)"
  cp -a /etc/fstab "$backup"
  awk '
    /^[[:space:]]*#/ { print; next }
    NF >= 3 && $3 == "swap" { print "# GATI disabled swap: " $0; next }
    { print }
  ' /etc/fstab > /etc/fstab.gati-new
  chmod 0644 /etc/fstab.gati-new
  mv /etc/fstab.gati-new /etc/fstab
  printf 'Swap disabled; original fstab saved as %s.\n' "$backup"
else
  printf 'No active swap detected.\n'
fi

# Prevent core dumps from retaining process memory on disk.
install -d -m 0755 /etc/systemd/coredump.conf.d /etc/security/limits.d
cat > /etc/systemd/coredump.conf.d/gati.conf <<'EOF'
[Coredump]
Storage=none
ProcessSizeMax=0
EOF
cat > /etc/security/limits.d/gati.conf <<'EOF'
* hard core 0
EOF
systemctl daemon-reload

systemctl enable --now ssh 2>/dev/null || systemctl enable --now sshd 2>/dev/null || true
printf '\nVPS bootstrap completed.\n'
printf 'Deployment user: %s\n' "$deploy_user"
printf 'Docker: %s\n' "$(docker --version)"
printf 'Compose: %s\n' "$(docker compose version)"
printf 'GATI directories: /opt/gati/releases and /opt/gati/audits\n'
printf 'Next: configure GitHub SSH secrets, Cloudflare named tunnel, and the private runtime files.\n'
