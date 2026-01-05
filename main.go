package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	configFileName = "config.mtime"
	timeLayout     = "20060102150405" // yyyyMMddHHmmss 格式
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "log":
		if len(os.Args) < 3 || len(os.Args) > 4 {
			fmt.Fprintln(os.Stderr, "用法: mtstamp log <目录路径> [保存目录]")
			fmt.Fprintln(os.Stderr, "  目录路径: 要遍历的目录（必须使用绝对路径）")
			fmt.Fprintln(os.Stderr, "  保存目录: config.mtime 的保存目录（可选，必须是绝对路径，不传则保存在运行命令所在的目录）")
			os.Exit(1)
		}
		dir := os.Args[2]
		var saveDir string
		if len(os.Args) == 4 {
			saveDir = os.Args[3]
		}
		if err := runMtimeLog(dir, saveDir); err != nil {
			fmt.Fprintln(os.Stderr, "log 错误:", err)
			os.Exit(1)
		}
	case "back":
		if len(os.Args) < 3 || len(os.Args) > 4 {
			fmt.Fprintln(os.Stderr, "用法: mtstamp back <目录路径> [config.mtime所在目录]")
			fmt.Fprintln(os.Stderr, "  目录路径: 要恢复文件时间的目录（必须使用绝对路径）")
			fmt.Fprintln(os.Stderr, "  config.mtime所在目录: config.mtime 文件所在的目录（可选，必须是绝对路径）")
			fmt.Fprintln(os.Stderr, "                       如果不传，则使用运行命令时所在的目录（当前工作目录）")
			os.Exit(1)
		}
		dir := os.Args[2]
		var configDir string
		if len(os.Args) == 4 {
			configDir = os.Args[3]
		}
		if err := runMtimeBack(dir, configDir); err != nil {
			fmt.Fprintln(os.Stderr, "back 错误:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintln(os.Stderr, "未知命令:", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("mtstamp - 记录和恢复文件修改时间的工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  mtstamp log <目录路径> [保存目录]")
	fmt.Println("      遍历指定目录下所有文件，记录每个文件的相对路径和修改时间，生成 config.mtime。")
	fmt.Println("      目录路径: 要遍历的目录，必须使用绝对路径。")
	fmt.Println("      保存目录: config.mtime 的保存目录（可选），必须使用绝对路径。")
	fmt.Println("               如果不指定保存目录，则保存在运行命令时所在的目录（当前工作目录）。")
	fmt.Println()
	fmt.Println("  mtstamp back <目录路径> [config.mtime所在目录]")
	fmt.Println("      根据 config.mtime 中记录的内容，恢复目录下所有文件的修改时间。")
	fmt.Println("      目录路径: 要恢复文件时间的目录，必须使用绝对路径。")
	fmt.Println("      config.mtime所在目录: config.mtime 文件所在的目录（可选），必须使用绝对路径。")
	fmt.Println("                          如果不指定，则使用运行命令时所在的目录（当前工作目录）下的 config.mtime。")
}

// runMtimeLog 遍历目录，生成 config.mtime
//
// 参数:
//   root: 要遍历的目录路径（必须是绝对路径）
//   saveDir: config.mtime 的保存目录（可选，必须是绝对路径）
//           如果为空字符串，则保存在运行命令时所在的目录（当前工作目录）
//
// config.mtime 格式：每行一个文件
//   <相对路径>\t<yyyyMMddHHmmss格式的时间>
//   其中相对路径是相对于 root 目录的路径
//   时间格式示例：20231225143045 表示 2023年12月25日14点30分45秒
func runMtimeLog(root, saveDir string) error {
	// 验证要遍历的目录路径必须是绝对路径
	if !filepath.IsAbs(root) {
		return fmt.Errorf("目录路径必须是绝对路径: %s", root)
	}

	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是目录", root)
	}

	// 确定 config.mtime 的保存位置
	var configPath string
	if saveDir == "" {
		// 如果不指定保存目录，使用当前工作目录
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("获取当前工作目录失败: %w", err)
		}
		configPath = filepath.Join(wd, configFileName)
	} else {
		// 如果指定了保存目录，验证必须是绝对路径
		if !filepath.IsAbs(saveDir) {
			return fmt.Errorf("保存目录必须是绝对路径: %s", saveDir)
		}
		// 确保保存目录存在
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			return fmt.Errorf("创建保存目录失败: %w", err)
		}
		configPath = filepath.Join(saveDir, configFileName)
	}
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		// 跳过我们自己生成的 config.mtime（无论它保存在哪里）
		if filepath.Clean(path) == filepath.Clean(configPath) {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}
		modTime := fi.ModTime()
		timeStr := modTime.Format(timeLayout)

		line := fmt.Sprintf("%s\t%s\n", filepath.ToSlash(rel), timeStr)
		if _, err := writer.WriteString(line); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("已生成配置文件: %s\n", configPath)
	return nil
}

// runMtimeBack 根据 config.mtime 恢复文件的修改时间
//
// 参数:
//   root: 要恢复文件时间的目录路径（必须是绝对路径）
//   configDir: config.mtime 文件所在的目录（可选，必须是绝对路径）
//              如果为空字符串，则使用运行命令时所在的目录（当前工作目录）
func runMtimeBack(root, configDir string) error {
	// 验证要恢复的目录路径必须是绝对路径
	if !filepath.IsAbs(root) {
		return fmt.Errorf("目录路径必须是绝对路径: %s", root)
	}

	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是目录", root)
	}

	// 确定 config.mtime 文件的路径
	var configPath string
	if configDir == "" {
		// 如果不指定 config.mtime 所在目录，使用当前工作目录
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("获取当前工作目录失败: %w", err)
		}
		configPath = filepath.Join(wd, configFileName)
	} else {
		// 如果指定了 config.mtime 所在目录，验证必须是绝对路径
		if !filepath.IsAbs(configDir) {
			return fmt.Errorf("config.mtime所在目录必须是绝对路径: %s", configDir)
		}
		configPath = filepath.Join(configDir, configFileName)
	}

	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	restored := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		relPath, timeStr, err := parseConfigLine(line)
		if err != nil {
			return fmt.Errorf("解析配置文件第 %d 行失败: %w", lineNum, err)
		}

		// 解析时间字符串为 time.Time
		// 使用 ParseInLocation 指定本地时区，确保与 Format 时使用的时区一致
		tm, err := time.ParseInLocation(timeLayout, timeStr, time.Local)
		if err != nil {
			return fmt.Errorf("解析配置文件第 %d 行的时间格式失败: %w", lineNum, err)
		}

		// 解析文件路径：config.mtime 中记录的路径可以是相对路径或绝对路径
		// 如果是绝对路径，直接使用；如果是相对路径，则相对于 root 目录
		var fullPath string
		if filepath.IsAbs(relPath) {
			fullPath = relPath
		} else {
			// 将相对路径转换为相对于 root 的完整路径
			// filepath.FromSlash 将斜杠转换为当前平台的分隔符
			fullPath = filepath.Join(root, filepath.FromSlash(relPath))
		}

		if _, err := os.Stat(fullPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// 如果文件不存在，给出提示但不视为致命错误
				fmt.Fprintf(os.Stderr, "警告: 文件不存在，跳过: %s\n", fullPath)
				continue
			}
			return err
		}
		// 将文件的访问时间（atime）和修改时间（mtime）都设置为记录的时间
		// os.Chtimes 在 Windows 和 Linux 平台都支持
		if err := os.Chtimes(fullPath, tm, tm); err != nil {
			return fmt.Errorf("设置文件时间失败 %s: %w", fullPath, err)
		}
		restored++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Printf("已恢复 %d 个文件的修改时间\n", restored)
	return nil
}

// parseConfigLine 解析 config.mtime 文件中的一行
//
// 格式: <路径>\t<yyyyMMddHHmmss格式的时间>
// 路径可以是相对路径（相对于 root 目录）或绝对路径
// 时间格式示例：20231225143045 表示 2023年12月25日14点30分45秒
// 返回: 路径字符串、时间字符串、错误信息
func parseConfigLine(line string) (relPath string, timeStr string, err error) {
	fields := strings.Split(line, "\t")
	if len(fields) != 2 {
		return "", "", fmt.Errorf("行格式不正确，应为: <路径>\\t<时间>，实际: %q", line)
	}

	relPath = strings.TrimSpace(fields[0])
	if relPath == "" {
		return "", "", fmt.Errorf("路径为空")
	}

	timeStr = strings.TrimSpace(fields[1])
	if timeStr == "" {
		return "", "", fmt.Errorf("时间为空")
	}

	// 验证时间格式是否正确（应该是14位数字）
	if len(timeStr) != 14 {
		return "", "", fmt.Errorf("时间格式不正确，应为14位数字（yyyyMMddHHmmss），实际长度: %d，内容: %q", len(timeStr), timeStr)
	}

	// 验证是否全为数字（time.Parse会进一步验证格式的正确性）
	for _, r := range timeStr {
		if r < '0' || r > '9' {
			return "", "", fmt.Errorf("时间格式不正确，应全为数字（yyyyMMddHHmmss），实际: %q", timeStr)
		}
	}

	return relPath, timeStr, nil
}


