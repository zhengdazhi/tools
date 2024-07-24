package main

import (
	"flag"

	"du/dirsize" // 导入你创建的包，替换为实际的模块路径
)

func main() {
	var s bool
	var h bool
	var help bool

	// 定义标志
	flag.BoolVar(&s, "s", false, "display only a total for each argument")
	flag.BoolVar(&h, "h", false, "print sizes in human readable format (e.g., 1K 234M 2G)")
	flag.BoolVar(&help, "help", false, "show help information")
	// 解析标志
	flag.Parse()

	// 获取非标志参数
	dirs := flag.Args()

	// 调用包中的功能函数
	dirsize.DisplayDirSizes(s, h, help, dirs)
}
