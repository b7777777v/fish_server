@echo off
setlocal enabledelayedexpansion

REM 設置專案根目錄
for %%I in ("%~dp0..") do set "PROJECT_ROOT=%%~fI"

REM 定義 Protobuf 相關的目錄
set "PROTO_SRC_DIR=%PROJECT_ROOT%\api"
set "GO_OUT_DIR=%PROJECT_ROOT%"
set "PROTO_FILES_DIR=%PROJECT_ROOT%\api\proto\v1"

REM 檢查 protoc 是否存在
where protoc >nul 2>nul
if %errorlevel% neq 0 (
    echo.
    echo ERROR: protoc is not installed or not in your PATH.
    echo Visit: https://grpc.io/docs/protoc-installation/
    echo.
    exit /b 1
)

echo Finding .proto files in %PROTO_FILES_DIR%...

set "PROTO_FILES="
pushd %PROTO_SRC_DIR%

for /r . %%f in (*.proto) do (
    set "filepath=%%f"
    set "relative_path=!filepath:%PROTO_SRC_DIR%\=!"
    
    REM --- 核心修正：將反斜線 \ 替換為斜線 / ---
    set "unix_path=!relative_path:\=/!"
    
    set "PROTO_FILES=!PROTO_FILES! !unix_path!"
)

popd

if not defined PROTO_FILES (
    echo.
    echo WARNING: No .proto files found.
    echo.
    exit /b 0
)

echo Generating for files:%PROTO_FILES%
echo.

REM 執行 protoc 命令
protoc ^
    --proto_path=%PROTO_SRC_DIR% ^
    --go_out=%GO_OUT_DIR% ^
    --go-grpc_out=%GO_OUT_DIR% ^
    %PROTO_FILES%

echo.
echo Protobuf code generated successfully.
echo.

endlocal