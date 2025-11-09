# 陣型配置管理 API 說明

## 概述

陣型配置系統提供了完整的 REST API 來管理魚群陣型的生成配置，包括難度設置、生成率控制、統計查詢等功能。

## 配置存儲

- **Redis**: 配置自動保存到 Redis，過期時間為 24 小時
- **啟動載入**: 服務啟動時自動從 Redis 載入配置
- **默認配置**: 如果 Redis 中沒有配置，使用默認的普通難度配置

## API 端點

### 1. 獲取當前配置

**GET** `/admin/formations/config`

獲取當前的陣型生成配置。

**Response:**
```json
{
  "success": true,
  "data": {
    "enabled": true,
    "min_interval": "20s",
    "max_interval": "1m0s",
    "base_spawn_chance": 0.3,
    "formation_weights": {
      "v": 0.25,
      "line": 0.20,
      "circle": 0.15,
      "triangle": 0.15,
      "diamond": 0.10,
      "wave": 0.10,
      "spiral": 0.05
    },
    "min_fish_count": 5,
    "max_fish_count": 20,
    "fish_size_preferences": {
      "small": 0.50,
      "medium": 0.35,
      "large": 0.12,
      "boss": 0.03
    },
    "max_concurrent_formations": 3,
    "dynamic_difficulty": true,
    "special_event_multiplier": 1.0
  }
}
```

### 2. 更新配置

**PUT** `/admin/formations/config`

更新陣型配置（支持部分更新）。

**Request Body:**
```json
{
  "enabled": true,
  "min_interval": 15,
  "max_interval": 45,
  "base_spawn_chance": 0.4,
  "fish_size_preferences": {
    "small": 0.30,
    "medium": 0.40,
    "large": 0.25,
    "boss": 0.05
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Formation config updated successfully",
  "data": { /* 更新後的完整配置 */ }
}
```

### 3. 設置難度（快捷方式）

**POST** `/admin/formations/difficulty`

快速設置預定義的難度等級。

**Request Body:**
```json
{
  "difficulty": "hard"
}
```

**可用難度:**
- `easy`: 簡單模式（30-90s 間隔，20% 概率，70% 小型魚）
- `normal`: 普通模式（20-60s 間隔，30% 概率，平衡分佈）
- `hard`: 困難模式（15-45s 間隔，40% 概率，更多大型魚）
- `boss_rush`: Boss 模式（10-30s 間隔，60% 概率，30% Boss 魚）

**Response:**
```json
{
  "success": true,
  "message": "Formation difficulty set to hard",
  "data": { /* 新的配置 */ }
}
```

### 4. 設置生成率

**POST** `/admin/formations/spawn-rate`

調整陣型生成的時間間隔和概率。

**Request Body:**
```json
{
  "min_interval": 20,
  "max_interval": 60,
  "base_chance": 0.35
}
```

**參數說明:**
- `min_interval`: 最小生成間隔（秒）
- `max_interval`: 最大生成間隔（秒）
- `base_chance`: 基礎生成概率（0.0-1.0）

**Response:**
```json
{
  "success": true,
  "message": "Formation spawn rate updated successfully"
}
```

### 5. 啟用/禁用陣型生成

**POST** `/admin/formations/enable?enabled=true`

啟用或禁用陣型生成功能。

**Query Parameters:**
- `enabled`: `true` 或 `false`

**Response:**
```json
{
  "success": true,
  "message": "Formation spawn enabled status updated",
  "enabled": true
}
```

### 6. 觸發特殊事件

**POST** `/admin/formations/trigger-event`

觸發臨時的特殊事件，提高陣型生成率。

**Request Body:**
```json
{
  "multiplier": 2.0,
  "duration": 300
}
```

**參數說明:**
- `multiplier`: 生成率倍數（例如 2.0 表示雙倍生成率）
- `duration`: 持續時間（秒）

**Response:**
```json
{
  "success": true,
  "message": "Special formation event triggered",
  "multiplier": 2.0,
  "duration": 300
}
```

### 7. 獲取統計信息

**GET** `/admin/formations/stats`

獲取陣型生成的統計數據。

**Response:**
```json
{
  "success": true,
  "data": {
    "total_spawned": 145,
    "successful_spawns": 142,
    "failed_spawns": 3,
    "current_formations": 2,
    "last_spawn_time": "2025-11-09T17:30:45Z",
    "success_rate": 0.979
  }
}
```

## 使用示例

### 示例 1: 啟動時設置為困難模式

```bash
curl -X POST http://localhost:8081/admin/formations/difficulty \
  -H "Content-Type: application/json" \
  -d '{"difficulty": "hard"}'
```

### 示例 2: 調整生成率

```bash
curl -X POST http://localhost:8081/admin/formations/spawn-rate \
  -H "Content-Type: application/json" \
  -d '{
    "min_interval": 15,
    "max_interval": 45,
    "base_chance": 0.5
  }'
```

### 示例 3: 觸發雙倍生成事件（5分鐘）

```bash
curl -X POST http://localhost:8081/admin/formations/trigger-event \
  -H "Content-Type: application/json" \
  -d '{
    "multiplier": 2.0,
    "duration": 300
  }'
```

### 示例 4: 自定義配置

```bash
curl -X PUT http://localhost:8081/admin/formations/config \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "min_interval": 25,
    "max_interval": 70,
    "base_spawn_chance": 0.25,
    "max_concurrent_formations": 4,
    "fish_size_preferences": {
      "small": 0.40,
      "medium": 0.35,
      "large": 0.20,
      "boss": 0.05
    }
  }'
```

## 配置參數詳解

### 基礎配置
- **enabled**: 是否啟用陣型生成
- **min_interval**: 最小生成間隔（秒）
- **max_interval**: 最大生成間隔（秒）
- **base_spawn_chance**: 基礎生成概率（0.0-1.0）

### 陣型權重 (formation_weights)
各種陣型類型的生成權重：
- `v`: V 字型
- `line`: 直線型
- `circle`: 圓形
- `triangle`: 三角形
- `diamond`: 菱形
- `wave`: 波浪型
- `spiral`: 螺旋型

### 魚類尺寸偏好 (fish_size_preferences)
各種魚類尺寸的生成概率：
- `small`: 小型魚
- `medium`: 中型魚
- `large`: 大型魚
- `boss`: Boss 魚

### 高級配置
- **max_concurrent_formations**: 最大並發陣型數量
- **dynamic_difficulty**: 是否根據玩家數量動態調整難度
- **special_event_multiplier**: 特殊事件生成倍率

## 注意事項

1. **配置持久化**: 配置會自動保存到 Redis，有效期 24 小時
2. **動態生效**: 配置更新後立即生效，無需重啟服務
3. **安全性**: 生產環境建議添加認證中間件保護 Admin API
4. **驗證**: API 會驗證參數的有效性（如概率範圍、時間間隔等）
5. **統計重置**: 重啟服務後統計數據會重置

## 錯誤處理

API 返回標準的錯誤響應格式：

```json
{
  "success": false,
  "error": "錯誤類型",
  "details": "詳細錯誤信息"
}
```

常見錯誤：
- 400 Bad Request: 參數驗證失敗
- 404 Not Found: 資源不存在
- 500 Internal Server Error: 服務器內部錯誤
