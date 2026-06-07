package persistence

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool 创建 PostgreSQL 连接池。
func NewPool(ctx context.Context, dsn string) *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("解析数据库 DSN 失败: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("数据库 Ping 失败: %v", err)
	}

	fmt.Println("✅ PostgreSQL 连接成功")
	return pool
}

// RunMigrations 执行幂等建表迁移，多次调用不会报错。
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, migrationSQL); err != nil {
		return fmt.Errorf("执行数据库迁移失败: %w", err)
	}
	log.Println("✅ 数据库迁移完成")
	return nil
}

// migrationSQL 与 docker/init.sql 及 testutil/pg.go 保持结构一致。
// 全部使用 IF NOT EXISTS 保证幂等。
const migrationSQL = `
CREATE TABLE IF NOT EXISTS users (
    email          TEXT PRIMARY KEY,
    password_hash  TEXT NOT NULL,
    nickname       TEXT NOT NULL,
    student_id     TEXT NOT NULL DEFAULT '',
    department     TEXT NOT NULL DEFAULT '',
    tags           TEXT[] NOT NULL DEFAULT '{}',
    avatar         TEXT NOT NULL DEFAULT '',
    contact_info   TEXT NOT NULL DEFAULT '',
    role           TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verify_token   TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS availabilities (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(email) ON DELETE CASCADE,
    date       DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time   TIME NOT NULL,
    CONSTRAINT chk_time_range CHECK (end_time > start_time)
);

CREATE TABLE IF NOT EXISTS appointments (
    id            TEXT PRIMARY KEY,
    mentor_id     TEXT NOT NULL REFERENCES users(email),
    student_id    TEXT NOT NULL REFERENCES users(email),
    time_slot_id  TEXT NOT NULL REFERENCES availabilities(id),
    message       TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'pending',
    reject_reason TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_status CHECK (status IN ('pending', 'accepted', 'rejected'))
);

CREATE TABLE IF NOT EXISTS recruitment_cards (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL UNIQUE REFERENCES users(email) ON DELETE CASCADE,
    skills           TEXT[] NOT NULL DEFAULT '{}',
    target_companies TEXT[] NOT NULL DEFAULT '{}',
    role             VARCHAR(20) NOT NULL DEFAULT 'both',
    experience_years INT NOT NULL DEFAULT 0,
    bio              TEXT NOT NULL DEFAULT '',
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_role CHECK (role IN ('interviewee', 'interviewer', 'both'))
);

CREATE INDEX IF NOT EXISTS idx_avail_user  ON availabilities(user_id);
CREATE INDEX IF NOT EXISTS idx_avail_date  ON availabilities(date);
CREATE INDEX IF NOT EXISTS idx_appt_mentor  ON appointments(mentor_id);
CREATE INDEX IF NOT EXISTS idx_appt_student ON appointments(student_id);
CREATE INDEX IF NOT EXISTS idx_appt_slot    ON appointments(time_slot_id);
CREATE INDEX IF NOT EXISTS idx_appt_status  ON appointments(status);
CREATE INDEX IF NOT EXISTS idx_card_user    ON recruitment_cards(user_id);
CREATE INDEX IF NOT EXISTS idx_card_role    ON recruitment_cards(role);

-- 迁移：为已有数据库添加 role 列（幂等）
DO $$ BEGIN
  ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
`
