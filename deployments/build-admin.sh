#!/bin/bash
# ========================================
# Admin Service Docker 構建腳本
# ========================================

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
IMAGE_NAME="fish-server-admin"
IMAGE_TAG="latest"
CONTAINER_NAME="fish-admin-test"
PORT=6060

# 函數：打印帶顏色的消息
print_message() {
    echo -e "${2}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

print_info() {
    print_message "$1" "${BLUE}"
}

print_success() {
    print_message "$1" "${GREEN}"
}

print_warning() {
    print_message "$1" "${YELLOW}"
}

print_error() {
    print_message "$1" "${RED}"
}

# 函數：清理函數
cleanup() {
    print_info "清理臨時容器..."
    docker rm -f ${CONTAINER_NAME} 2>/dev/null || true
}

# 設置清理陷阱
trap cleanup EXIT

# 檢查 Docker 是否運行
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker 未運行或無法訪問"
        exit 1
    fi
}

# 構建鏡像
build_image() {
    print_info "開始構建 Admin Service Docker 鏡像..."
    
    # 切換到項目根目錄
    cd "$(dirname "$0")/.."
    
    # 構建鏡像
    docker build \
        -f deployments/Dockerfile.admin \
        -t ${IMAGE_NAME}:${IMAGE_TAG} \
        --build-arg BUILDTIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
        .
    
    if [ $? -eq 0 ]; then
        print_success "鏡像構建成功！"
        
        # 顯示鏡像信息
        print_info "鏡像信息："
        docker images ${IMAGE_NAME}:${IMAGE_TAG}
    else
        print_error "鏡像構建失敗！"
        exit 1
    fi
}

# 測試鏡像
test_image() {
    print_info "開始測試 Docker 鏡像..."
    
    # 檢查是否有同名容器在運行
    if docker ps -a --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        print_warning "發現同名容器，正在移除..."
        docker rm -f ${CONTAINER_NAME}
    fi
    
    # 啟動測試容器
    print_info "啟動測試容器..."
    docker run -d \
        --name ${CONTAINER_NAME} \
        -p ${PORT}:6060 \
        -e LOG_LEVEL=debug \
        ${IMAGE_NAME}:${IMAGE_TAG}
    
    # 等待容器啟動
    print_info "等待容器啟動..."
    sleep 5
    
    # 檢查容器狀態
    if ! docker ps --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        print_error "容器啟動失敗！"
        print_info "容器日誌："
        docker logs ${CONTAINER_NAME}
        exit 1
    fi
    
    # 測試健康檢查端點
    print_info "測試健康檢查端點..."
    for i in {1..10}; do
        if curl -s http://localhost:${PORT}/ping >/dev/null 2>&1; then
            print_success "健康檢查通過！"
            break
        fi
        if [ $i -eq 10 ]; then
            print_error "健康檢查失敗！"
            print_info "容器日誌："
            docker logs ${CONTAINER_NAME}
            exit 1
        fi
        print_info "等待服務啟動... (${i}/10)"
        sleep 2
    done
    
    # 測試主要端點
    print_info "測試主要端點..."
    
    # 測試根端點
    if curl -s http://localhost:${PORT}/ | grep -q "Fish Server Admin API"; then
        print_success "根端點測試通過"
    else
        print_warning "根端點測試失敗"
    fi
    
    # 測試健康檢查端點
    if curl -s http://localhost:${PORT}/admin/health | grep -q "healthy"; then
        print_success "健康檢查端點測試通過"
    else
        print_warning "健康檢查端點測試失敗"
    fi
    
    # 顯示容器信息
    print_info "容器信息："
    docker ps --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    
    print_success "所有測試通過！"
}

# 顯示使用說明
show_usage() {
    echo "用法: $0 [OPTIONS] [COMMAND]"
    echo ""
    echo "COMMANDS:"
    echo "  build    - 構建 Docker 鏡像"
    echo "  test     - 測試 Docker 鏡像"
    echo "  all      - 構建並測試鏡像 (默認)"
    echo "  clean    - 清理鏡像和容器"
    echo "  logs     - 查看測試容器日誌"
    echo "  stop     - 停止測試容器"
    echo ""
    echo "OPTIONS:"
    echo "  -t, --tag TAG     設置鏡像標籤 (默認: latest)"
    echo "  -p, --port PORT   設置測試端口 (默認: 6060)"
    echo "  -h, --help        顯示此幫助信息"
}

# 清理鏡像和容器
clean_all() {
    print_info "清理所有相關的容器和鏡像..."
    
    # 停止並移除容器
    docker rm -f ${CONTAINER_NAME} 2>/dev/null || true
    
    # 移除鏡像
    docker rmi ${IMAGE_NAME}:${IMAGE_TAG} 2>/dev/null || true
    
    print_success "清理完成！"
}

# 查看日誌
show_logs() {
    if docker ps -a --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        print_info "顯示容器日誌："
        docker logs -f ${CONTAINER_NAME}
    else
        print_error "測試容器不存在！"
        exit 1
    fi
}

# 停止容器
stop_container() {
    print_info "停止測試容器..."
    docker stop ${CONTAINER_NAME} 2>/dev/null || true
    docker rm ${CONTAINER_NAME} 2>/dev/null || true
    print_success "容器已停止並移除"
}

# 主函數
main() {
    # 解析命令行參數
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--tag)
                IMAGE_TAG="$2"
                shift 2
                ;;
            -p|--port)
                PORT="$2"
                shift 2
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            build)
                COMMAND="build"
                shift
                ;;
            test)
                COMMAND="test"
                shift
                ;;
            all)
                COMMAND="all"
                shift
                ;;
            clean)
                COMMAND="clean"
                shift
                ;;
            logs)
                COMMAND="logs"
                shift
                ;;
            stop)
                COMMAND="stop"
                shift
                ;;
            *)
                print_error "未知參數: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # 默認命令
    COMMAND=${COMMAND:-all}
    
    print_info "開始執行 Admin Service Docker 構建流程..."
    print_info "鏡像名稱: ${IMAGE_NAME}:${IMAGE_TAG}"
    print_info "測試端口: ${PORT}"
    
    # 檢查 Docker
    check_docker
    
    case $COMMAND in
        build)
            build_image
            ;;
        test)
            test_image
            ;;
        all)
            build_image
            test_image
            ;;
        clean)
            clean_all
            ;;
        logs)
            show_logs
            ;;
        stop)
            stop_container
            ;;
    esac
    
    print_success "操作完成！"
}

# 執行主函數
main "$@"