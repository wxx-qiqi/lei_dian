package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	_ "modernc.org/sqlite" // 使用纯 Go 版本的 SQLite
)

// SQLite 数据库文件
const (
	dbFile    = "imei.db"    // SQLite 数据库文件名
	tableName = "imei_store" // IMEI 存储表
)

func init() {
	db, err := ConnectDB()
	if err != nil {
		log.Fatalf("数据库连接失败: %v\n", err)
	}
	defer db.Close()

	// 创建表格（如果不存在）
	err = createTable(db)
	if err != nil {
		log.Fatalf("创建表失败: %v\n", err)
	}
}

// ConnectDB 连接 SQLite 数据库
func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbFile) // SQLite 连接
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}
	return db, nil
}

// createTable 创建 IMEI 表
func createTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS imei_store (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		imei TEXT UNIQUE NOT NULL,
		instance_index TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("表创建失败: %v", err)
	}
	return nil
}

// InsertStoreIMEI 插入 IMEI，避免高并发冲突
func InsertStoreIMEI(db *sql.DB, instanceIndex, imei string) error {
	// 事务插入，防止并发冲突
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("事务启动失败: %v\n", err)
	}

	_, err = tx.Exec("INSERT OR IGNORE INTO imei_store (imei, instance_index) VALUES (?, ?)", imei, instanceIndex)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("数据库插入失败: %v\n", err)
	}

	tx.Commit()
	fmt.Printf("模拟器【%s】的 IMEI【%s】 插入成功\n", instanceIndex, imei)
	return nil
}

// CheckIMEIExists 查询 IMEI 是否已存在
func CheckIMEIExists(db *sql.DB, imei string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM imei_store WHERE imei = ?", imei).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("查询数据库失败: %v\n", err)
	}
	return count > 0, nil
}
