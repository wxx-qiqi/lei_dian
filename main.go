package main

import (
	"database/sql"
	"fmt"
	lei "lei_dian/lei_dian_utils"
	"lei_dian/mysql"
	"log"
	"sync"
	"time"
)

const (
	PackageName    = "cn.damai"
	ClassName      = "cn.damai.homepage.MainActivity"
	maxConcurrency = 5 // 限制最大并发数
)

func main() {
	// 连接数据库
	db, err := mysql.ConnectDB()
	if err != nil {
		log.Fatalf("数据库连接失败: %v\n", err)
	}
	defer db.Close()

	Simulators, err := lei.GetSimulators()
	if err != nil {
		fmt.Println("main获取所有模拟器", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency) // 控制并发数
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
			fmt.Println("======================== 启动app =========================")
			er := lei.RunApp("", simulatorId, PackageName)
			if er != nil {
				log.Printf("启动app 【%s】 失败: %v\n", simulatorId, err)
				<-sem // 释放信号量
				return
			}
			fmt.Println("======================== 等待20秒 =========================")
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
			exists, er := mysql.CheckIMEIExists(db, imei)
			if er != nil {
				log.Printf("mysql检验 【%s】  失败: %v\n", imei, err)
				<-sem // 释放信号量
				return
			}
			if !exists {
				errr := mysql.InsertStoreIMEI(db, simulatorId, imei)
				if errr != nil {
					log.Printf("mysql插入 【%s】  失败: %v\n", imei, err)
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
