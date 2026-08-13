#!/usr/bin/env bash
# =============================================================================
# Accessible Path - deploy script for Beget VPS (Ubuntu 22.04/24.04)
#
# Usage:
#   ./deploy/deploy.sh setup      - install docker + swap (first time, needs root)
#   ./deploy/deploy.sh up         - generate keys (if missing), build and start stack
#   ./deploy/deploy.sh restart    - rebuild changed images and restart stack
#   ./deploy/deploy.sh down       - stop the stack
#   ./deploy/deploy.sh logs [svc] - follow logs (default: all)
#   ./deploy/deploy.sh ssl-init   - obtain Let's Encrypt certificate (DOMAIN set)
#   ./deploy/deploy.sh ssl-renew  - renew certificates (run from cron)
#   ./deploy/deploy.sh status     - container health summary
#
# Requires: docker + compose plugin, .env.prod (copy of .env.prod.example)
# =============================================================================
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

ENV_FILE=".env.prod"
COMPOSE_FILE="docker-compose.prod.yml"
COMPOSE_CMD=(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE")

log()  { echo -e "\033[1;32m[deploy]\033[0m $*"; }
warn() { echo -e "\033[1;33m[deploy]\033[0m $*"; }
die()  { echo -e "\033[1;31m[deploy]\033[0m $*" >&2; exit 1; }

require_env_file() {
  [ -f "$ENV_FILE" ] || die "Missing $ENV_FILE. Run: cp $ENV_FILE.example $ENV_FILE && edit it"
}

# -----------------------------------------------------------------------------
# First-time server setup (needs root/sudo)
# -----------------------------------------------------------------------------
cmd_setup() {
  if [ "$(id -u)" -ne 0 ]; then
    die "Run 'setup' with sudo: sudo ./deploy/deploy.sh setup"
  fi

  log "Adding swap 2G (protects the 2GB RAM VPS during builds)"
  if ! swapon --show | grep -q swapfile; then
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile >/dev/null
    swapon /swapfile
    grep -q '/swapfile' /etc/fstab || echo '/swapfile none swap sw 0 0' >> /etc/fstab
  else
    log "Swap already present, skipping"
  fi

  log "Installing docker"
  if ! command -v docker >/dev/null 2>&1; then
    curl -fsSL https://get.docker.com | sh
  else
    log "Docker already installed"
  fi

  systemctl enable --now docker
  log "Docker compose plugin check"
  docker compose version || die "docker compose plugin missing (install docker-compose-plugin)"
  log "Setup done. Now: cp .env.prod.example .env.prod, edit it, then ./deploy/deploy.sh up"
}

# -----------------------------------------------------------------------------
# JWT keys: services/auth-service/keys/{private,public}.pem
# -----------------------------------------------------------------------------
ensure_jwt_keys() {
  local dir="services/auth-service/keys"
  if [ -f "$dir/private.pem" ] && [ -f "$dir/public.pem" ]; then
    log "JWT keys already exist in $dir"
    return
  fi
  mkdir -p "$dir"
  log "Generating JWT RSA keypair in $dir"
  openssl genrsa -out "$dir/private.pem" 2048 >/dev/null 2>&1
  openssl rsa -in "$dir/private.pem" -pubout -out "$dir/public.pem" >/dev/null 2>&1
  chmod 600 "$dir/private.pem"
}

# -----------------------------------------------------------------------------
# Build + start
# -----------------------------------------------------------------------------
cmd_up() {
  require_env_file
  ensure_jwt_keys

  log "Building images (first build takes a while on 1 CPU)"
  "${COMPOSE_CMD[@]}" build --parallel

  log "Starting stack"
  "${COMPOSE_CMD[@]}" up -d --remove-orphans

  log "Waiting for nginx (all services healthy)..."
  for i in $(seq 1 60); do
    if docker exec accessible-path-prod-nginx pidof nginx >/dev/null 2>&1; then
      log "Stack is up. Nginx: http://$(hostname -I 2>/dev/null | awk '{print $1}' || echo '<vps-ip>')"
      if grep -q '^DOMAIN=.\+' "$ENV_FILE" 2>/dev/null; then
        log "Domain set - run SSL: ./deploy/deploy.sh ssl-init"
      fi
      exit 0
    fi
    sleep 5
  done
  warn "Nginx not healthy yet. Check: ${COMPOSE_CMD[*]} logs nginx"
}

cmd_restart() {
  require_env_file
  "${COMPOSE_CMD[@]}" build --parallel
  "${COMPOSE_CMD[@]}" up -d --remove-orphans
  log "Restarted"
}

cmd_down() {
  require_env_file
  "${COMPOSE_CMD[@]}" down
}

cmd_logs() {
  require_env_file
  "${COMPOSE_CMD[@]}" logs -f "${1:-}"
}

cmd_status() {
  require_env_file
  "${COMPOSE_CMD[@]}" ps
}

# -----------------------------------------------------------------------------
# Let's Encrypt via certbot (webroot through nginx on port 80)
# Certificates land in infra/nginx/ssl/live/accessible-path/ (fixed cert-name)
# -----------------------------------------------------------------------------
require_domain() {
  require_env_file
  DOMAIN="$(grep '^DOMAIN=.\+' "$ENV_FILE" | head -1 | cut -d= -f2- || true)"
  CERTBOT_EMAIL="$(grep '^CERTBOT_EMAIL=.\+' "$ENV_FILE" | head -1 | cut -d= -f2- || true)"
  [ -n "$DOMAIN" ] || die "DOMAIN is empty in $ENV_FILE"
  [ -n "$CERTBOT_EMAIL" ] || die "CERTBOT_EMAIL is empty in $ENV_FILE"
}

certbot_cmd() {
  docker run --rm \
    -v "$ROOT_DIR/infra/nginx/ssl:/etc/letsencrypt" \
    -v "$ROOT_DIR/infra/nginx/certbot:/var/www/certbot" \
    certbot/certbot "$@"
}

cmd_ssl_init() {
  require_domain
  mkdir -p infra/nginx/ssl infra/nginx/certbot
  log "Obtaining certificate for $DOMAIN"
  certbot_cmd certonly --webroot -w /var/www/certbot \
    --cert-name accessible-path \
    -d "$DOMAIN" -m "$CERTBOT_EMAIL" \
    --agree-tos --no-eff-email --non-interactive
  log "Reloading nginx"
  docker exec accessible-path-prod-nginx nginx -s reload
  log "HTTPS is live: https://$DOMAIN"
}

cmd_ssl_renew() {
  require_domain
  log "Renewing certificates"
  certbot_cmd renew --webroot -w /var/www/certbot --non-interactive
  log "Reloading nginx"
  docker exec accessible-path-prod-nginx nginx -s reload || true
}

# -----------------------------------------------------------------------------
case "${1:-}" in
  setup)      cmd_setup ;;
  up)         cmd_up ;;
  restart)    cmd_restart ;;
  down)       cmd_down ;;
  logs)       cmd_logs "${2:-}" ;;
  status)     cmd_status ;;
  ssl-init)   cmd_ssl_init ;;
  ssl-renew)  cmd_ssl_renew ;;
  *)
    echo "Usage: $0 {setup|up|restart|down|logs [svc]|status|ssl-init|ssl-renew}"
    exit 1
    ;;
esac
