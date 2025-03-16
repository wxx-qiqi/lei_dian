package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

// MySQL 连接信息
const (
	dbUser     = "root"      // 修改为你的 MySQL 用户名
	dbPassword = "root"      // 修改为你的 MySQL 密码
	dbHost     = "127.0.0.1" // MySQL 地址
	dbPort     = "3306"      // MySQL 端口
	dbName     = "ldplayer"  // 数据库名
	tableName  = "imei_store"
)

func init() {
	// 连接数据库
	db, err := getDBConn()
	if err != nil {
		log.Fatalf("数据库连接失败: %v\n", err)
	}
	defer db.Close()

	// 创建数据库和表格（如果不存在）
	err = createDatabaseAndTable(db)
	if err != nil {
		log.Fatalf("数据库和表格创建失败: %v\n", err)
	}
}

// getDBConn 获取数据库连接 生成 MySQL 连接字符串
func getDBConn() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True",
		dbUser, dbPassword, dbHost, dbPort)
	return sql.Open("mysql", dsn)
}

// createDatabaseAndTable 创建数据库及表格
func createDatabaseAndTable(db *sql.DB) error {
	// 创建数据库
	if !isDatabaseExist(db, dbName) {
		_, err := db.Exec("CREATE DATABASE IF NOT EXISTS ldplayer")
		if err != nil {
			return fmt.Errorf("数据库创建失败: %v", err)
		}

		// 使用新创建的数据库
		_, err = db.Exec(fmt.Sprintf("USE %s", dbName))
		if err != nil {
			return fmt.Errorf("切换数据库失败: %v", err)
		}
	}
	if !isTableExist(db, dbName, tableName) {
		// 创建 imei_store 表（如果不存在）
		createTableSQL := `
	CREATE TABLE IF NOT EXISTS imei_store (
		id INT AUTO_INCREMENT PRIMARY KEY,
		imei VARCHAR(15) UNIQUE NOT NULL,
		instance_index VARCHAR(3) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
		_, err := db.Exec(createTableSQL)
		if err != nil {
			return fmt.Errorf("表格创建失败: %v", err)
		}
	}
	return nil
}

// 检查数据库是否存在
func isDatabaseExist(db *sql.DB, dbName string) bool {
	query := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?"
	var name string
	err := db.QueryRow(query, dbName).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false // 数据库不存在
		}
		log.Fatalf("查询数据库是否存在失败: %v\n", err)
	}
	return true // 数据库存在
}

// 检查数据表是否存在
func isTableExist(db *sql.DB, dbName, tableName string) bool {
	query := "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?"
	var name string
	err := db.QueryRow(query, dbName, tableName).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false // 表格不存在
		}
		log.Fatalf("查询数据表是否存在失败: %v\n", err)
	}
	return true // 数据表存在
}

// ConnectDB 连接数据库
func ConnectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	return sql.Open("mysql", dsn)
}

// InsertStoreIMEI 保存IMEI 并存入 MySQL
func InsertStoreIMEI(db *sql.DB, instanceIndex, imei string) error {
	// 插入数据库（如果已存在则重新生成）
	_, err := db.Exec("INSERT INTO imei_store (imei, instance_index) VALUES (?, ?)", imei, instanceIndex)
	if err != nil {
		//if strings.Contains(err.Error(), "Duplicate entry") {
		//	fmt.Printf("IMEI 【%s】 已存在，重新生成...\n", imei)
		//}
		return fmt.Errorf("数据库插入失败: %v\n", err)
	}
	fmt.Printf("模拟器【%s】的{%s}插入成功\n", instanceIndex, imei)
	return nil
}

// CheckIMEIExists 查询 IMEI 是否存在
func CheckIMEIExists(db *sql.DB, imei string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM imei_store WHERE imei = ?"
	err := db.QueryRow(query, imei).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("查询 IMEI 失败: %v", err)
	}
	return count > 0, nil
}
