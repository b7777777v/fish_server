#!/bin/bash
# ========================================
# 環境管理腳本
# ========================================

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 函數：打印帶顏色的消息
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 顯示使用說明
show_usage() {
    echo "用法: $0 [ENVIRONMENT] [COMMAND]"
    echo ""
    echo "ENVIRONMENTS:"
    echo "  dev       - 開發環境 (啟用 pprof, 詳細日誌)"
    echo "  staging   - 預發布環境 (關閉 pprof, 中等安全性)"
    echo "  prod      - 生產環境 (最高安全性, 最佳性能)"
    echo ""
    echo "COMMANDS:"
    echo "  build     - 構建指定環境的鏡像"
    echo "  up        - 啟動指定環境"
    echo "  down      - 停止指定環境"
    echo "  restart   - 重啟指定環境"
    echo "  logs      - 查看指定環境日誌"
    echo "  status    - 查看指定環境狀態"
    echo "  clean     - 清理指定環境"
    echo ""
    echo "範例:"
    echo "  $0 dev up          # 啟動開發環境"
    echo "  $0 staging build   # 構建 staging 鏡像"
    echo "  $0 prod status     # 查看生產環境狀態"
}

# 驗證環境參數
validate_environment() {
    case $1 in
        dev|staging|prod)
            return 0
            ;;
        *)
            print_error "無效的環境: $1"
            show_usage
            exit 1
            ;;
    esac
}

# 檢查必要的文件
check_prerequisites() {
    local env=$1
    
    print_info "檢查 $env 環境的必要文件..."
    
    # 檢查 docker-compose 文件
    if [ ! -f "docker-compose.$env.yml" ]; then
        print_error "找不到 docker-compose.$env.yml"
        exit 1
    fi
    
    # 檢查配置文件
    if [ ! -f "config-docker.$env.yaml" ]; then
        print_error "找不到 config-docker.$env.yaml"
        exit 1
    fi
    
    # 檢查環境變量文件
    if [ ! -f ".env.$env" ]; then
        print_warning "找不到 .env.$env，將使用默認配置"
    fi
    
    print_success "必要文件檢查完成"
}

# 構建鏡像
build_image() {
    local env=$1
    print_info "構建 $env 環境的 Docker 鏡像..."
    
    # 設置鏡像標籤
    local image_tag="fish-server-admin:$env"
    if [ "$env" = "dev" ]; then
        image_tag="fish-server-admin:latest"
    fi
    
    # 構建鏡像
    docker build \
        -f Dockerfile.admin \
        -t $image_tag \
        --build-arg ENVIRONMENT=$env \
        --build-arg BUILDTIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
        ..
    
    print_success "鏡像構建完成: $image_tag"
}

# 啟動環境
start_environment() {
    local env=$1
    print_info "啟動 $env 環境..."
    
    # 載入環境變量
    if [ -f ".env.$env" ]; then
        export $(cat .env.$env | grep -v '^#' | xargs)
    fi
    
    # 啟動服務
    docker-compose -f docker-compose.$env.yml --env-file .env.$env up -d
    
    print_success "$env 環境啟動成功"
    print_info "可用端點:"
    
    case $env in
        dev)
            echo "  - Admin API: http://localhost:6060"
            echo "  - Environment Info: http://localhost:6060/admin/env"
            echo "  - Pprof (已啟用): http://localhost:6060/debug/pprof/"
            ;;
        staging)
            echo "  - Admin API: http://localhost:6061"
            echo "  - Environment Info: http://localhost:6061/admin/env"
            echo "  - Pprof: 已關閉 (訪問 /debug/pprof/disabled 查看原因)"
            ;;
        prod)
            echo "  - Admin API: http://localhost:6062"
            echo "  - Environment Info: http://localhost:6062/admin/env"
            echo "  - Pprof: 已關閉 (生產環境)"
            ;;
    esac
}

# 停止環境
stop_environment() {
    local env=$1
    print_info "停止 $env 環境..."
    
    docker-compose -f docker-compose.$env.yml down
    
    print_success "$env 環境已停止"
}

# 重啟環境
restart_environment() {
    local env=$1
    print_info "重啟 $env 環境..."
    
    stop_environment $env
    sleep 2
    start_environment $env
}

# 查看日誌
show_logs() {
    local env=$1
    print_info "顯示 $env 環境日誌..."
    
    docker-compose -f docker-compose.$env.yml logs -f admin
}

# 查看狀態
show_status() {
    local env=$1
    print_info "$env 環境狀態:"
    
    docker-compose -f docker-compose.$env.yml ps
    
    # 測試健康狀態
    local port
    case $env in
        dev) port=6060 ;;
        staging) port=6061 ;;
        prod) port=6062 ;;
    esac
    
    if curl -s http://localhost:$port/ping >/dev/null 2>&1; then
        print_success "服務健康檢查通過"
    else
        print_warning "服務健康檢查失敗"
    fi
}

# 清理環境
clean_environment() {
    local env=$1
    print_info "清理 $env 環境..."
    
    # 停止並移除容器
    docker-compose -f docker-compose.$env.yml down -v
    
    # 移除鏡像
    local image_tag="fish-server-admin:$env"
    if [ "$env" = "dev" ]; then
        image_tag="fish-server-admin:latest"
    fi
    
    docker rmi $image_tag 2>/dev/null || true
    
    print_success "$env 環境清理完成"
}

# 主函數
main() {
    if [ $# -lt 1 ]; then
        show_usage
        exit 1
    fi
    
    local env=$1
    local command=${2:-up}
    
    # 特殊處理 help
    if [ "$env" = "help" ] || [ "$env" = "-h" ] || [ "$env" = "--help" ]; then
        show_usage
        exit 0
    fi
    
    # 驗證環境
    validate_environment $env
    
    # 切換到腳本目錄
    cd "$(dirname "$0")"
    
    # 檢查必要文件
    check_prerequisites $env
    
    # 執行命令
    case $command in
        build)
            build_image $env
            ;;
        up|start)
            start_environment $env
            ;;
        down|stop)
            stop_environment $env
            ;;
        restart)
            restart_environment $env
            ;;
        logs)
            show_logs $env
            ;;
        status)
            show_status $env
            ;;
        clean)
            clean_environment $env
            ;;
        *)
            print_error "未知命令: $command"
            show_usage
            exit 1
            ;;
    esac
}

# 執行主函數
main "$@"