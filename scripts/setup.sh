#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# Personal Bookkeeping — 项目初始化脚本
# 用法: bash scripts/setup.sh
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; }
info() { echo -e "${CYAN}[i]${NC} $1"; }

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║    Personal Bookkeeping — 初始化脚本        ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════╝${NC}"
echo ""

# =============================================================================
# 1. 环境检查
# =============================================================================
echo -e "${YELLOW}[1/6] 环境检查${NC}"

# Git
if ! command -v git &>/dev/null; then
  err "git 未安装"
  exit 1
fi
log "git: $(git --version | head -1)"

# Docker
if ! command -v docker &>/dev/null; then
  err "docker 未安装"
  exit 1
fi
log "docker: $(docker --version | head -1)"

# Docker Compose
if docker compose version &>/dev/null; then
  COMPOSE="docker compose"
  log "docker compose: $(docker compose version | head -1)"
elif command -v docker-compose &>/dev/null; then
  COMPOSE="docker-compose"
  log "docker-compose: $(docker-compose --version | head -1)"
else
  err "docker compose 未安装"
  exit 1
fi

# =============================================================================
# 2. 配置文件
# =============================================================================
echo ""
echo -e "${YELLOW}[2/6] 配置文件${NC}"

CONFIG_FILE="backend/config.yaml"
if [ -f "$CONFIG_FILE" ]; then
  log "配置文件已存在: $CONFIG_FILE"
else
  warn "未找到 $CONFIG_FILE，将创建默认配置"
  cat > "$CONFIG_FILE" << 'EOF'
server:
  port: "8000"

db:
  host: "localhost"
  port: "5432"
  user: "bookkeeper"
  password: "bookkeeper_dev"
  name: "bookkeeping"
  sslmode: "disable"

jwt:
  secret: "change-this-in-production"
  expire_minutes: 10080

cors:
  origins: "http://localhost:5173,http://localhost:3000"

log:
  target: "file"
  dir: "logs"
  info: "app.log"
  warn: "warn.log"
  error: "error.log"
  max_size: 100
  max_age: 30
  max_backups: 10
  compress: true

cache:
  type: "redis"
  ttl: 300

queue:
  enabled: false
  type: "redis"
  workers: 5
  max_retries: 3

exchange_rate:
  provider: "exchangerate-api"
  api_key: ""
  base: "CNY"

ocr:
  provider: "paddleocr"
  endpoint: "http://paddleocr:9000"
EOF
  log "默认配置已创建: $CONFIG_FILE"
fi

# =============================================================================
# 3. JWT 密钥生成
# =============================================================================
echo ""
echo -e "${YELLOW}[3/6] 安全设置${NC}"

if grep -q "change-this-in-production" "$CONFIG_FILE" 2>/dev/null; then
  NEW_SECRET=$(openssl rand -hex 32 2>/dev/null || uuidgen 2>/dev/null || echo "dev-$(date +%s)")
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/change-this-in-production/$NEW_SECRET/" "$CONFIG_FILE"
  else
    sed -i "s/change-this-in-production/$NEW_SECRET/" "$CONFIG_FILE"
  fi
  log "JWT 密钥已生成"
else
  log "JWT 密钥已配置"
fi

# =============================================================================
# 4. 汇率 API Key（可选）
# =============================================================================
echo ""
echo -e "${YELLOW}[4/6] 汇率配置${NC}"

if grep -q 'api_key: ""' "$CONFIG_FILE" 2>/dev/null; then
  read -r -p "$(echo -e "${YELLOW}?${NC} 输入 ExchangeRate-API key（留空跳过，汇率自动更新将不启用）: ")" API_KEY
  if [ -n "$API_KEY" ]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' "s/api_key: \"\"/api_key: \"$API_KEY\"/" "$CONFIG_FILE"
    else
      sed -i "s/api_key: \"\"/api_key: \"$API_KEY\"/" "$CONFIG_FILE"
    fi
    log "汇率 API Key 已配置"
  else
    warn "跳过，汇率自动更新将不启用（不影响手动录入）"
  fi
else
  log "汇率 API Key 已配置"
fi

# =============================================================================
# 5. Docker 启动
# =============================================================================
echo ""
echo -e "${YELLOW}[5/6] 启动服务${NC}"

# 检查端口占用
check_port() {
  if command -v ss &>/dev/null; then
    ss -tln "sport = :$1" 2>/dev/null | grep -q LISTEN && return 0
  elif command -v lsof &>/dev/null; then
    lsof -i ":$1" &>/dev/null && return 0
  fi
  return 1
}

for port in 5432 6379 8000 9000 3000; do
  if check_port "$port"; then
    warn "端口 $port 已被占用 — 请检查是否有其他服务在运行"
  fi
done

info "拉取镜像并启动服务..."
$COMPOSE up -d --wait 2>&1 | while IFS= read -r line; do echo "  $line"; done
log "所有服务已启动"

# =============================================================================
# 6. 验证
# =============================================================================
echo ""
echo -e "${YELLOW}[6/6] 健康检查${NC}"

sleep 3

# 等待后端就绪
MAX_RETRY=20
RETRY=0
while [ $RETRY -lt $MAX_RETRY ]; do
  if curl -sf http://localhost:8000/api/v1/health >/dev/null 2>&1; then
    break
  fi
  RETRY=$((RETRY + 1))
  sleep 2
done

if [ $RETRY -eq $MAX_RETRY ]; then
  warn "后端健康检查超时，请检查 docker compose logs backend"
else
  log "后端 API 健康检查通过"
fi

# PaddleOCR 检查
if curl -sf http://localhost:9000/health >/dev/null 2>&1; then
  log "PaddleOCR 服务正常"
else
  warn "PaddleOCR 未响应（部分版本不支持 /health，不影响使用）"
fi

# =============================================================================
# 完成
# =============================================================================
echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║       初始化完成!                           ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  前端:      ${CYAN}http://localhost:3000${NC}"
echo -e "  API:       ${CYAN}http://localhost:8000/api/v1${NC}"
echo -e "  Swagger:   ${CYAN}http://localhost:8000/swagger/index.html${NC}"
echo -e "  Metrics:   ${CYAN}http://localhost:8000/metrics${NC}"
echo -e "  PostgreSQL:${CYAN} localhost:5432${NC}"
echo -e "  Redis:     ${CYAN}localhost:6379${NC}"
echo -e "  PaddleOCR: ${CYAN}localhost:9000${NC}"
echo ""
info "常用命令:"
echo -e "  docker compose logs -f   查看日志"
echo -e "  docker compose down      停止服务"
echo -e "  make test                运行测试"
echo -e "  make lint                代码检查"
echo ""
warn "首次启动需等待 PaddleOCR 模型下载（约 1-2 分钟）"
warn "生产环境请修改 config.yaml 中的 JWT 密钥和数据库密码"
echo ""
