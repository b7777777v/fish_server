#!/bin/bash

# 當任何命令失敗時立即退出
set -e

# 專案根目錄 (腳本所在的上一層目錄)
PROJECT_ROOT=$(dirname "$(dirname "$(realpath "$0")")")

# 定義 Protobuf 相關的目錄
PROTO_SRC_DIR="${PROJECT_ROOT}/api"
PROTO_FILES_DIR="${PROJECT_ROOT}/api/proto/v1"
GO_OUT_DIR="${PROJECT_ROOT}" # 輸出到專案根目錄，protoc 會根據 go_package 自動創建路徑

# 檢查 protoc 是否安裝
if ! command -v protoc &> /dev/null; then
    echo " protoc is not installed. Please install protobuf compiler."
    echo "   Visit: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# 檢查 protoc-gen-go 和 protoc-gen-go-grpc 是否安裝
# 通常透過 `go install` 安裝
if ! command -v protoc-gen-go &> /dev/null || ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo " Go protobuf plugins not found. Installing..."
    # 確保在專案目錄下執行 go install，以便它們被添加到 GOBIN
    (cd "${PROJECT_ROOT}" && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest)
    (cd "${PROJECT_ROOT}" && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest)
    echo " Go protobuf plugins installed."
fi

echo " Finding .proto files in ${PROTO_FILES_DIR}..."

# 查找所有 .proto 文件
PROTO_FILES=$(find "${PROTO_FILES_DIR}" -name "*.proto")

if [ -z "$PROTO_FILES" ]; then
    echo " No .proto files found. Exiting."
    exit 0
fi

echo " Generating Go code from .proto files..."

# 執行 protoc 命令
# --proto_path (-I): 指定 import 的搜索路徑。我們設為 api/，這樣在 proto 文件中可以 import "proto/v1/common.proto"
# --go_out: 指定 Go 原始碼的輸出目錄。設為專案根目錄，protoc 會根據 go_package 創建路徑。
# --go-grpc_out: 指定 gRPC 服務代碼的輸出目錄。
protoc \
    --proto_path="${PROTO_SRC_DIR}" \
    --go_out="${GO_OUT_DIR}" \
    --go-grpc_out="${GO_OUT_DIR}" \
    ${PROTO_FILES}

echo " Protobuf code generated successfully."
echo "   Output directory: ${PROJECT_ROOT}/pkg/pb/v1"