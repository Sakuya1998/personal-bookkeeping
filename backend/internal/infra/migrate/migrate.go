package migrate

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Run 执行数据库迁移。
// dir: migrations 目录路径（绝对或相对 migrate 包）
// dsn: postgres DSN
func Run(dir, dsn string) error {
	srcURL := fmt.Sprintf("file://%s", dir)
	m, err := migrate.New(srcURL, dsn)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// Down 回滚所有迁移。
func Down(dir, dsn string) error {
	srcURL := fmt.Sprintf("file://%s", dir)
	m, err := migrate.New(srcURL, dsn)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}
