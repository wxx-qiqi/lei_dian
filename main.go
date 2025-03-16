package main

import (
	"database/sql"
	"fmt"
	"gopkg.in/ini.v1"
	lei "lei_dian/lei_dian_utils"
	"lei_dian/sqlite"
	"log"
	"sync"
	"time"
)

var (
	PackageName    string
	MaxConcurrency int
)

func init() {
	// 读取配置文件
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
		return
	}

	// 获取配置项
	PackageName = cfg.Section("settings").Key("PackageName").String()
	MaxConcurrency, err = cfg.Section("settings").Key("maxConcurrency").Int()
	if err != nil {
		log.Fatalf("读取 MaxConcurrency 失败: %v", err)
		return
	}
}

func main() {
	// 连接数据库
	db, err := sqlite.ConnectDB()
	if err != nil {
		log.Fatalf("连接数据库失败: %v\n", err)
	}
	defer db.Close()

	Simulators, err := lei.GetSimulators()
	if err != nil {
		fmt.Println("main获取所有模拟器", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, MaxConcurrency) // 控制并发数
	for _, simulator := range Simulators {
		if simulator.ID != "0" {
			wg.Add(1)
			go HandleAutoImei(db, simulator, &wg, sem)
		}
	}

	wg.Wait() // 等待所有实例启动完成
	fmt.Println("=== 所有模拟器实例任务完成 ===")
	lei.QuitAll()
}

func HandleAutoImei(db *sql.DB, simulator lei.LDSimulator, wg *sync.WaitGroup, sem chan struct{}) {
	fmt.Println("========== 开始处理HandleAutoImei ===============")
	defer wg.Done()
	count := 0

	sem <- struct{}{}

	simulatorId := simulator.ID
	for {
		count++
		fmt.Printf("=============模拟器【%s】开始{--%d---}次==============\n", simulatorId, count)
		// 修改imei
		err := lei.ModifyAutoImei(simulatorId)
		if err != nil {
			log.Printf("修改 【%s】 imei失败: %v\n", simulatorId, err)
			<-sem // 释放信号量
			return
		}
		if count > 1 {
			time.Sleep(2 * time.Second)
		}

		// 启动模拟器
		fmt.Printf("模拟器 【%s】 开始启动\n", simulatorId)
		err = lei.StartSimulators(simulatorId)
		if err != nil {
			log.Printf("启动模拟器 【%s】 失败: %v\n", simulatorId, err)
			<-sem // 释放信号量
			return
		}

		// 等待模拟器完全启动
		isStart := lei.WaitForBootComplete(simulatorId)
		fmt.Printf("模拟器 【%s】 完全启动\n", simulatorId)
		if isStart {
			time.Sleep(2 * time.Second)
			fmt.Printf("======================== 【%s】启动app =========================\n", simulatorId)
			er := lei.RunApp("", simulatorId, PackageName)
			if er != nil {
				log.Printf("启动app 【%s】 失败: %v\n", simulatorId, err)
				<-sem // 释放信号量
				return
			}
			fmt.Printf("========================【%s】 等待20秒 =========================\n", simulatorId)
			time.Sleep(20 * time.Second)
		}

		// 获取 IMEI
		imei, err := lei.GetPropImei(simulatorId)
		if err != nil {
			log.Printf("获取模拟器 【%s】 IMEI 失败: %v\n", simulatorId, err)
			<-sem // 释放信号量
			return
		}
		fmt.Printf("模拟器 【%s】 获取到 IMEI: %s\n", simulatorId, imei)

		// 存储 IMEI 到数据库
		if lei.IsValidIMEI(imei) {
			exists, er := sqlite.CheckIMEIExists(db, imei)
			if er != nil {
				log.Printf("数据库检验 【%s】  失败: %v\n", imei, err)
				<-sem // 释放信号量
				return
			}
			if !exists {
				errr := sqlite.InsertStoreIMEI(db, simulatorId, imei)
				if errr != nil {
					log.Printf("数据库插入 【%s】  失败: %v\n", imei, err)
					<-sem // 释放信号量
					return
				}
			}
		}

		// 关闭模拟器
		fmt.Printf("=====================关闭模拟器【%s】=========================\n", simulatorId)
		lei.Quit("", simulatorId)

	}

}
