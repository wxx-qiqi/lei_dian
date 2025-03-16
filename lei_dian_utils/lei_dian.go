package lei_dian_utils

import (
	"bytes"
	"fmt"
	"lei_dian/tools"
	"os/exec"
	"strings"
	"time"
)

const (
	LdPath      = `D:\leidian\LDPlayer9\ldconsole.exe`
	MaxWaitTime = 60 // 最大等待时间 120 秒
	Interval    = 2  // 每 2 秒检查一次
	NotStarted  = "0"
	Starting    = "1"
	Running     = "2"
)

// https://www.ldmnq.com/forum/30.html

// LDSimulator 定义模拟器实例结构体
type LDSimulator struct {
	ID            string `json:"id"`             // 索引
	Name          string `json:"name"`           // 标题
	TopWindow     string `json:"top_window"`     // 顶层窗口句柄
	BindWindow    string `json:"bind_window"`    // 绑定窗口句柄
	AndroidStatus string `json:"android_status"` // 是否进入 Android（0=未进入，1=已进入,2=进入中）
	ProcessPID    string `json:"process_pid"`    // 进程 PID
	VBoxPID       string `json:"vbox_pid"`       // VBox 进程 PID
	Width         string `json:"width"`          // 屏幕宽
	Height        string `json:"height"`         // 屏幕高
	DPI           string `json:"dpi"`            // dpi（像素密度）
}

// GetSimulators 获取已有的模拟器实例列表
func GetSimulators() ([]LDSimulator, error) {
	cmd := exec.Command(LdPath, "list2")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	// 转换 GBK -> UTF-8
	utf8Output, err := tools.GBKToUTF8(output)
	if err != nil {
		return nil, err
	}

	// 按行解析数据
	lines := strings.Split(strings.TrimSpace(string(utf8Output)), "\n")
	var Simulators []LDSimulator

	for _, line := range lines {
		fields := strings.Split(line, ",") // 按逗号分割
		if len(fields) >= 10 {             // 确保字段数量正确
			Simulator := LDSimulator{
				ID:            fields[0],
				Name:          fields[1],
				TopWindow:     fields[2],
				BindWindow:    fields[3],
				AndroidStatus: fields[4],
				ProcessPID:    fields[5],
				VBoxPID:       fields[6],
				Width:         fields[7],
				Height:        fields[8],
				DPI:           fields[9],
			}
			Simulators = append(Simulators, Simulator)
		}
	}

	return Simulators, nil
}

// CreateSimulators 创建新的模拟器实例
func CreateSimulators(name string) error {
	fmt.Printf("实例 %s 不存在，正在创建...\n", name)
	cmd := exec.Command(LdPath, "add", "--name", name)
	return cmd.Run()
}

// CopySimulator 复制模拟器 from参数既可以是名字也可以是索引，判断规则为如果全数字就认为是索引，否则是名字
func CopySimulator() {

}

// RemoveSimulator 删除模拟器
func RemoveSimulator() {

}

// StartSimulators 启动多个雷电模拟器实例
func StartSimulators(id string) error {
	fmt.Printf("正在启动实例 【%s】...\n", id)
	cmd := exec.Command(LdPath, "launch", "--index", id)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("启动实例 【%s】 失败: %v\n", id, err)
	} else {
		fmt.Printf("雷电模拟器实例 【%s】 已启动\n", id)
	}
	return err
}

// ContainsSimulators 判断实例是否在已存在的列表中
func ContainsSimulators(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// 属性设置部分 调用modify需要在模拟器启动前，不然可能不生效

// ModifyAutoImei 随机生成Imei 启动前修改imei否者不生效
func ModifyAutoImei(id string) error {
	fmt.Printf("正在生成实例 【%s】 的imei\n", id)
	cmd := exec.Command(LdPath, "modify", "--index", id, "--imei", "auto")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("实例 【%s】 的imei生成失败: %v\n", id, err)
	} else {
		fmt.Printf("实例 【%s】 的imei生成成功\n", id)
	}
	return err
}

// GetPropImei 获取imei
func GetPropImei(id string) (string, error) {
	fmt.Printf("正在获取实例 【%s】 的imei\n", id)
	cmd := exec.Command(LdPath, "getprop", "--index", id, "--key", "phone.imei")
	// 捕获命令输出
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("获取 IMEI 失败: %v", err)
	}

	// 解析输出，去除换行和空格
	imei := strings.TrimSpace(out.String())
	if imei == "" {
		return "", fmt.Errorf("未能获取到 IMEI，可能模拟器未启动")
	}

	return imei, nil
}

func IsValidIMEI(imei string) bool {
	// 1. IMEI 必须是 15 位数字
	if len(imei) != 15 {
		fmt.Printf("无效 IMEI (长度错误): %s\n", imei)
		return false
	}

	// 2. 过滤包含 "adb.exe" 的错误信息
	if strings.Contains(imei, "adb.exe") {
		fmt.Printf("无效 IMEI (ADB 错误信息): %s\n", imei)
		return false
	}

	return true
}

// SetPropImei 设置imei
//func SetPropImei(id string) (string, error) {
//	fmt.Printf("正在获取实例 【%s】 的imei\n", id)
//	cmd := exec.Command(LdPath, "setprop",
//		"--index", id,
//		"--key", "phone.imei",
//		"--value", "auto")
//	// 捕获命令输出
//	var out bytes.Buffer
//	cmd.Stdout = &out
//	err := cmd.Run()
//	if err != nil {
//		return "", fmt.Errorf("设置 IMEI 失败: %v", err)
//	}
//
//	// 解析输出，去除换行和空格
//	imei := strings.TrimSpace(out.String())
//	if imei == "" {
//		return "", fmt.Errorf("未能设置 IMEI，可能模拟器未启动")
//	}
//
//	return imei, nil
//}

// RebootSimulator 重启模拟器
func RebootSimulator(name, id string) error {
	fmt.Printf("正在重启实例 【%s】 的imei\n", id)
	var cmd *exec.Cmd
	if name != "" {
		cmd = exec.Command(LdPath, "reboot", "--name", name)
	} else {
		cmd = exec.Command(LdPath, "reboot", "--index", id)
	}
	err := cmd.Run()
	if err != nil {
		fmt.Printf("实例 【%s】 重启失败: %v\n", id, err)
	} else {
		fmt.Printf("实例 【%s】 重启成功\n", id)
	}
	return err
}

// QuitAll 退出所有开着的模拟器
func QuitAll() {
	fmt.Printf("正在退出所有开着的模拟器\n")
	cmd := exec.Command(LdPath, "quitall")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("退出所有模拟器失败: %v\n", err)
	} else {
		fmt.Printf("退出所有模拟器成功\n")
	}
}

// Quit 退出对应的模拟器
func Quit(name, id string) {
	var cmd *exec.Cmd
	if name != "" {
		cmd = exec.Command(LdPath, "quit", "--name", name)
	} else {
		cmd = exec.Command(LdPath, "quit", "--index", id)
	}
	err := cmd.Run()
	if err != nil {
		fmt.Printf("退出模拟器失败: %v\n", err)
	} else {
		fmt.Printf("退出模拟器成功\n")
	}
}

func getByIdSimulators(instanceID string) (LDSimulator, error) {
	sis, err := GetSimulators()
	if err == nil {
		for _, si := range sis {
			if si.ID == instanceID {
				return si, nil
			}
		}
	}
	fmt.Printf("获取所有模拟器失败: %v\n", err)
	return LDSimulator{}, err
}

// RunApp 启动模拟器后才能 启动app
func RunApp(name, id, packagename string) error {
	var cmd *exec.Cmd
	if name != "" {
		cmd = exec.Command(LdPath, "runapp", "--name", name, "--packagename", packagename)
	} else {
		cmd = exec.Command(LdPath, "runapp", "--index", id, "--packagename", packagename)
	}
	err := cmd.Run()
	if err != nil {
		fmt.Printf("启动app失败: %v\n", err)
		return err
	}
	fmt.Printf("启动app成功\n")
	return nil
}

// WaitForBootComplete 检查模拟器是否完全启动
func WaitForBootComplete(instanceID string) bool {
	for i := 0; i < MaxWaitTime/Interval; i++ {
		// 执行 adb 命令检查 boot_completed 状态
		Simulators, err := getByIdSimulators(instanceID)
		if err == nil && Simulators.AndroidStatus == Starting {
			fmt.Printf("模拟器 【%s】 已完全启动\n", instanceID)
			return true
		}

		fmt.Printf("模拟器 【%s】 启动中... 等待 %d 秒\n", instanceID, Interval)
		time.Sleep(time.Duration(Interval) * time.Second)
	}

	fmt.Printf("模拟器 【%s】 启动超时，无法获取 IMEI\n", instanceID)
	return false
}
