-- TODO: 大廳模組 - 創建公告表
-- 此表儲存遊戲公告和活動訊息

CREATE TABLE IF NOT EXISTS announcements (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL, -- 公告標題
    content TEXT NOT NULL, -- 公告內容
    priority INT NOT NULL DEFAULT 0, -- 優先級（數字越大越重要）
    is_active BOOLEAN NOT NULL DEFAULT TRUE, -- 是否啟用
    start_time TIMESTAMP WITH TIME ZONE, -- 公告開始顯示時間（NULL 表示立即顯示）
    end_time TIMESTAMP WITH TIME ZONE, -- 公告結束顯示時間（NULL 表示永久顯示）
    created_by BIGINT, -- 建立者 ID（可關聯到 users 表，但不強制）
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 創建索引
CREATE INDEX idx_announcements_is_active ON announcements(is_active);
CREATE INDEX idx_announcements_priority ON announcements(priority DESC);
CREATE INDEX idx_announcements_time_range ON announcements(start_time, end_time);
CREATE INDEX idx_announcements_created_at ON announcements(created_at DESC);

-- 創建更新時間觸發器
CREATE OR REPLACE FUNCTION update_announcements_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_announcements_updated_at
    BEFORE UPDATE ON announcements
    FOR EACH ROW
    EXECUTE FUNCTION update_announcements_updated_at();

-- 插入示例公告
INSERT INTO announcements (title, content, priority, is_active) VALUES
('歡迎來到 Fish Server', '歡迎來到捕魚遊戲！每日登入可獲得免費金幣，快來體驗吧！', 100, TRUE),
('新魚種上線', '全新魚種「黃金鯨魚」已經上線，擊殺可獲得超高獎勵！', 80, TRUE);
