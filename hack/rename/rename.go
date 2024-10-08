package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const frameworkModuleName = "leafdev.top/ecosystem/billing"

func main() {
	// 输入新的 go.mod module
	var newModName string
	fmt.Printf("Enter new module name(eg: github.com/<your-username>/<your-project-name>): ")
	_, err := fmt.Scanln(&newModName)
	if err != nil {
		fmt.Printf("Unable get new module name: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Printf("Do you want to setup the project to %s? (y/n)", newModName)
	var answer string
	_, err = fmt.Scanln(&answer)
	if err != nil {
		fmt.Printf("Error reading user input: %v\n", err)
		os.Exit(1)
	}
	if answer != "y" {
		fmt.Printf("Aborting setup.\n")
	}

	// 修改 go.mod 文件中的 module 名称
	err = replaceInFile("../../go.mod", frameworkModuleName, newModName)
	if err != nil {
		fmt.Printf("Error replacing module name in go.mod: %v\n", err)
		os.Exit(1)
	}

	// 读取 go.mod 中的 module 名称
	modName, err := getModuleName("../../go.mod")
	if err != nil {
		fmt.Printf("Error reading go.mod: %v\n", err)
		os.Exit(1)
	}

	if modName == frameworkModuleName {
		fmt.Printf("Please update go.mod module to a different name.\n")
		os.Exit(1)
	}

	fmt.Printf("Module name found: %s\n", modName)
	// 遍历当前文件夹（排除 vendor、setup.go 和版本控制文件夹）
	err = filepath.Walk("../../", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 条件排除
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}
		if !info.IsDir() && info.Name() == "setup.go" {
			return nil
		}

		// 处理文件
		if !info.IsDir() {
			err := replaceInFile(path, `"`+frameworkModuleName, fmt.Sprintf(`"%s`, modName))
			if err != nil {
				fmt.Printf("Error replacing in file %s: %v\n", path, err)
			}
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	}

	// run go mod tidy
	fmt.Println("Running go mod tidy...")
	var aCmd = exec.Command("go", "mod", "tidy")
	if err := aCmd.Run(); err != nil {
		fmt.Printf("Error running go mod tidy: %v\n", err)
	}

}

// 读取 go.mod 文件中的 module 名称
func getModuleName(modFilePath string) (string, error) {
	file, err := os.Open(modFilePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("module name not found in go.mod")
}

// 在文件中替换指定的字符串
func replaceInFile(filePath string, old string, new string) error {
	input, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	output := strings.ReplaceAll(string(input), old, new)
	if err = os.WriteFile(filePath, []byte(output), 0666); err != nil {
		return err
	}

	return nil
}
