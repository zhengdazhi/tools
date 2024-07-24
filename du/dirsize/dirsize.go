package dirsize

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// DisplayDirSizes 处理目录大小的主要逻辑
func DisplayDirSizes(s, h, help bool, dirs []string) {
	// 处理帮助标志
	if help {
		printHelp()
		return
	}

	switch {
	// 如果没有输入要统计的目录则输出帮助信息
	case len(dirs) == 0:
		printHelp()
	// 遍历一级子目录，并输出可读大小
	case s && h:
		for _, dir := range dirs {
			subDirs, err := getSubDir(dir)
			if err != nil {
				fmt.Printf("Error getting subdirectories %s %v\n", dir, err)
				continue
			}
			for _, subDir := range subDirs {
				subDirPath := filepath.Join(dir, subDir)
				dirSize, err := getSize(subDirPath)
				if err != nil {
					fmt.Printf("Error getting subDirPath %s %v\n", subDirPath, err)
					continue
				}
				fmt.Printf("dir %s size: %s \n", subDir, formatSize(dirSize))
			}
		}
	// 遍历一级子目录大小
	case s:
		for _, dir := range dirs {
			subDirs, err := getSubDir(dir)
			if err != nil {
				fmt.Printf("Error getting subdirectories %s %v \n", dir, err)
				continue
			}
			for _, subDir := range subDirs {
				subDirPath := filepath.Join(dir, subDir)
				dirSize, err := getSize(subDirPath)
				if err != nil {
					fmt.Printf("Error getting subDirPath %s %v\n", subDirPath, err)
					continue
				}
				fmt.Printf("dir %s size: %d \n", subDir, dirSize)
			}
		}
	// 输出可读的大小
	case h:
		for _, dir := range dirs {
			dirSize, err := getSize(dir)
			if err != nil {
				fmt.Printf("Error getting dir %s %v \n", dir, err)
				continue
			}
			fmt.Printf("dir %s size: %s", dir, formatSize(dirSize))
		}
	// 只输入的要统计目录
	case len(dirs) > 0:
		fmt.Println("dir size")
		for _, dir := range dirs {
			dirSize, err := getSize(dir)
			if err != nil {
				fmt.Printf("Error getting dir %s %v \n", dir, err)
				continue
			}
			fmt.Printf("dir %s size: %d", dirSize)
		}
	// 没有输入任何参数打印帮助信息
	default:
		if len(dirs) == 0 {
			printHelp()
			return
		} else {
			fmt.Printf("Directories: %v\n", dirs)
		}
	}
}

// 自定义帮助信息函数
func printHelp() {
	fmt.Println("Help information")
	fmt.Println("Usage:")
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("-%s: %s (default: %v)\n", f.Name, f.Usage, f.Value)
	})
	fmt.Println("Example:")
	fmt.Println("   myprogram -flag \"dirs\"")
}

// 遍历一级子目录
func getSubDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading directory %v: %v\n", dir, err)
		return nil, err
	}
	var subDir []string
	for _, entry := range entries {
		subDir = append(subDir, entry.Name())
	}
	return subDir, nil
}

// 获取目录大小
func getSize(dir string) (int64, error) {
	// 忽略快捷方式
	if filepath.Ext(dir) == ".lnk" {
		return 0, nil
	} else if filepath.Ext(dir) == ".symlink" {
		return 0, nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		return 0, fmt.Errorf("error stating directory %v: %w", dir, err)
	}
	if !info.IsDir() {
		return info.Size(), nil // 如果是文件，返回文件大小
	}

	var dirSize int64
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading directory %v: %v\n", dir, err)
		return 0, err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			size, err := getSize(fullPath)
			if err != nil {
				fmt.Printf("Error getting size for directory %s: %v\n", fullPath, err)
				return 0, err
			}
			dirSize += size
		} else {
			fileInfo, err := os.Stat(fullPath)
			if err != nil {
				fmt.Printf("Error getting file info for %s: %v\n", fullPath, err)
				return 0, err
			}
			dirSize += fileInfo.Size()
		}
	}
	return dirSize, nil
}

func formatSize(size int64) string {
	const (
		_        = iota             // 忽略第一个 iota 值 (0)
		KB int64 = 1 << (10 * iota) // 1024 (2^10)
		MB                          // 1048576 (2^20)
		GB                          // 1073741824 (2^30)
		TB                          // 1099511627776 (2^40)
	)
	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}
