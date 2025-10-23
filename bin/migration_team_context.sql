-- Team Context Migration Script
-- 为 Token、Channel、Log 表添加上下文支持字段
-- 1. Token 表添加 owner_type 和 owner_id 字段
ALTER TABLE tokens
ADD COLUMN owner_type VARCHAR(10) DEFAULT 'user'
AFTER user_id;
ALTER TABLE tokens
ADD COLUMN owner_id INT DEFAULT 0
AFTER owner_type;
ALTER TABLE tokens
ADD INDEX idx_tokens_owner (owner_type, owner_id);
-- 2. Channel 表添加 owner_type 和 owner_id 字段
ALTER TABLE channels
ADD COLUMN owner_type VARCHAR(10) DEFAULT 'user'
AFTER user_id;
ALTER TABLE channels
ADD COLUMN owner_id INT DEFAULT 0
AFTER owner_type;
ALTER TABLE channels
ADD INDEX idx_channels_owner (owner_type, owner_id);
-- 3. Log 表添加 team_id 字段
ALTER TABLE logs
ADD COLUMN team_id INT DEFAULT 0
AFTER user_id;
ALTER TABLE logs
ADD INDEX idx_logs_team (team_id);
-- 4. 数据迁移：将现有数据设置为用户上下文
UPDATE tokens
SET owner_type = 'user',
    owner_id = user_id
WHERE owner_type IS NULL
    OR owner_type = '';
UPDATE channels
SET owner_type = 'user',
    owner_id = user_id
WHERE owner_type IS NULL
    OR owner_type = '';
-- 5. 添加外键约束（可选，根据需要启用）
-- ALTER TABLE tokens ADD CONSTRAINT fk_tokens_owner_user FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;
-- ALTER TABLE tokens ADD CONSTRAINT fk_tokens_owner_team FOREIGN KEY (owner_id) REFERENCES teams(id) ON DELETE CASCADE;
-- ALTER TABLE channels ADD CONSTRAINT fk_channels_owner_user FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;
-- ALTER TABLE channels ADD CONSTRAINT fk_channels_owner_team FOREIGN KEY (owner_id) REFERENCES teams(id) ON DELETE CASCADE;
-- ALTER TABLE logs ADD CONSTRAINT fk_logs_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL;