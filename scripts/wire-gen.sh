#!/bin/bash

# 當任何命令失敗時立即退出
set -e

# 專案根目錄
PROJECT_ROOT=$(dirname "$(dirname "$(realpath "$0")")")

# 檢查 wire 是否安裝
if ! command -v wire &> /dev/null; then
    echo "wire command not found. Installing..."
    # 使用 go install 安裝 wire 工具
    (cd "${PROJECT_ROOT}" && go install github.com/google/wire/cmd/wire@latest)
    echo "✅ wire installed."
fi

echo "Running wire to generate dependency injection code..."

# 進入專案根目錄
cd "${PROJECT_ROOT}"

# 執行 wire gen，並傳入 ./... 參數
# `./...` 是 Go 工具鏈的一個模式，代表“當前目錄及其所有子目錄”
# wire 會自動掃描所有子目錄，找到含有 `wire.go` 的地方並生成程式碼
wire gen ./...

echo "Wire code generated successfully."