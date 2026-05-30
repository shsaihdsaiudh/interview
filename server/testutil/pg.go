// Package testutil 提供集成测试所需的 PostgreSQL 容器管理等基础设施。
// 只在 `go test -tags=integration` 或默认（非 -short）时使用。
package testutil

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer 封装一个 testcontainers PostgreSQL 实例。
type PostgresContainer struct {
	container *postgres.PostgresContainer
	URL       string
}

// StartPostgres 启动一个 PostgreSQL 容器并执行 init.sql 建表。
// 调用方应在测试结束时调用 c.Terminate(ctx)。
func StartPostgres(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("interview_platform"),
		postgres.WithUsername("interview"),
		postgres.WithPassword("interview123"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("启动 PostgreSQL 容器失败: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("获取连接串失败: %w", err)
	}

	// 建表
	if err := runMigrations(ctx, connStr); err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("执行迁移失败: %w", err)
	}

	log.Printf("🧪 测试 PostgreSQL 就绪: %s", connStr)

	return &PostgresContainer{
		container: pgContainer,
		URL:       connStr,
	}, nil
}

// Terminate 销毁容器。
func (c *PostgresContainer) Terminate(ctx context.Context) error {
	return c.container.Terminate(ctx)
}

// NewPool 从容器连接串创建 pgxpool。
func (c *PostgresContainer) NewPool(ctx context.Context) *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(c.URL)
	if err != nil {
		panic(fmt.Sprintf("解析连接串失败: %v", err))
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		panic(fmt.Sprintf("创建连接池失败: %v", err))
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		panic(fmt.Sprintf("Ping 失败: %v", err))
	}

	return pool
}

// runMigrations 执行建表 SQL。
func runMigrations(ctx context.Context, connStr string) error {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, migrationSQL)
	return err
}

// migrationSQL 来自 docker/init.sql
const migrationSQL = `
CREATE TABLE users (
    email          TEXT PRIMARY KEY,
    password_hash  TEXT NOT NULL,
    nickname       TEXT NOT NULL,
    student_id     TEXT NOT NULL DEFAULT '',
    department     TEXT NOT NULL DEFAULT '',
    tags           TEXT[] NOT NULL DEFAULT '{}',
    avatar         TEXT NOT NULL DEFAULT '',
    contact_info   TEXT NOT NULL DEFAULT '',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verify_token   TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE availabilities (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(email) ON DELETE CASCADE,
    date       DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time   TIME NOT NULL,
    CONSTRAINT chk_time_range CHECK (end_time > start_time)
);

CREATE TABLE appointments (
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

CREATE INDEX idx_avail_user  ON availabilities(user_id);
CREATE INDEX idx_avail_date  ON availabilities(date);
CREATE INDEX idx_appt_mentor  ON appointments(mentor_id);
CREATE INDEX idx_appt_student ON appointments(student_id);
CREATE INDEX idx_appt_slot    ON appointments(time_slot_id);
CREATE INDEX idx_appt_status  ON appointments(status);
`
