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
