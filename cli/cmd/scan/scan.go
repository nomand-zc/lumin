package scan

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/spf13/cobra"

	"github.com/nomand-zc/lumin-client/cli/internal/auth"
	"github.com/nomand-zc/lumin-client/cli/internal/factory"
	"github.com/nomand-zc/lumin-client/credentials"
	"github.com/nomand-zc/lumin-client/log"
	"github.com/nomand-zc/lumin-client/pool/taskpool"
	"github.com/nomand-zc/lumin-client/providers"
)

var defaultCredScanner credScanner

// credScanner 持有 scan 命令的参数
type credScanner struct {
	srcDir       string
	destDir      string
	providerName string
	provider     providers.Provider
}

// CMD 返回 scan 子命令
func CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "扫描并校验凭证文件",
		Long:  `从指定目录下递归扫描所有 JSON 凭证文件，校验凭证是否有效，并将凭证文件按状态移动到对应目录。`,
	}

	// 注册子命令
	cmd.AddCommand(
		defaultCredScanner.cmd(),
	)

	return cmd
}

func (s *credScanner) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "校验凭证文件并按状态分类移动",
		Long: `从指定目录下递归扫描所有 JSON 凭证文件，校验凭证是否有效，
并根据校验结果将凭证文件移动到目标目录下的对应子目录：
  - enable:  凭证有效且未触发限流
  - limit:   凭证有效但触发了限流（临时不可用）
  - disable: 凭证永久失效（GetUsageStats 返回错误）

校验规则：
  1. 通过 Validate() 方法校验凭证文件格式是否正确
  2. 通过 GetUsageStats() 获取凭证使用情况，返回 error 则视为永久失效
  3. 如果 usage 触发了限流，则视为临时不可用
  4. 如果 usage 未触发限流，则视为有效凭证

示例：
  provider-client scan check --src /path/from/creds --dest /path/to/output --provider kiro`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.run()
		},
	}

	cmd.Flags().StringVarP(&s.srcDir, "src", "s", "", "凭证文件所在的源目录（必填）")
	cmd.Flags().StringVarP(&s.destDir, "dest", "d", "", "分类后凭证文件的目标目录（必填）")
	cmd.Flags().StringVarP(&s.providerName, "provider", "p", "kiro", fmt.Sprintf("provider 名称，支持：%v", factory.SupportedProviders))
	_ = cmd.MarkFlagRequired("src")
	_ = cmd.MarkFlagRequired("dest")

	return cmd
}

// credStatus 凭证状态
type credStatus string

const (
	statusEnable  credStatus = "enable"
	statusLimit   credStatus = "limit"
	statusDisable credStatus = "disable"
)

// run 执行扫描校验逻辑
func (s *credScanner) run() error {
	// 初始化 provider
	var err error
	s.provider, err = factory.NewProvider(s.providerName)
	if err != nil {
		return err
	}

	// 创建目标子目录
	if err := ensureDirs(s.destDir); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 统计计数（使用 atomic 保证并发安全）
	var enableCount, limitCount, disableCount, invalidCount atomic.Int64
	var mu sync.Mutex // 保护文件拷贝操作
	var wg sync.WaitGroup

	// 递归扫描所有 JSON 文件
	err = filepath.Walk(s.srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() ||
			!strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			return nil
		}

		fileName := info.Name()
		wg.Add(1)
		if err := taskpool.DefaultPool.Submit(func() {
			defer wg.Done()

			status, err := s.checkCredential(path)
			if err != nil {
				log.Warnf("\n凭证文件 %q 格式校验失败: %v，跳过", path, err)
				invalidCount.Add(1)
				return
			}

			// 根据状态拷贝文件（加锁保护文件写入）
			destPath := filepath.Join(s.destDir, string(status), fileName)
			mu.Lock()
			copyErr := copyFile(path, destPath)
			mu.Unlock()
			if copyErr != nil {
				log.Warnf("\n拷贝凭证文件 %q 到 %q 失败: %v", path, destPath, copyErr)
				return
			}

			switch status {
			case statusEnable:
				enableCount.Add(1)
				log.Infof("\n[有效] %s -> %s", path, destPath)
			case statusLimit:
				limitCount.Add(1)
				log.Infof("\n[限流] %s -> %s", path, destPath)
			case statusDisable:
				disableCount.Add(1)
				log.Infof("\n[失效] %s -> %s", path, destPath)
			}
		}); err != nil {
			wg.Done()
			log.Errorf("提交任务失败: %v", err)
			invalidCount.Add(1)
		}

		return nil
	})

	// 等待所有并发任务完成
	wg.Wait()

	if err != nil {
		return fmt.Errorf("扫描目录失败: %w", err)
	}

	total := enableCount.Load() + limitCount.Load() + disableCount.Load() + invalidCount.Load()
	log.Infof("\n扫描完成！总凭证数量: %d, 有效: %d, 限流: %d, 失效: %d, 格式无效: %d",
		total, enableCount.Load(), limitCount.Load(), disableCount.Load(), invalidCount.Load())

	return nil
}

// checkCredential 校验单个凭证文件，返回凭证状态
func (s *credScanner) checkCredential(filePath string) (credStatus, error) {
	// 加载凭证文件
	creds, err := auth.LoadCredentials(s.provider.Name(), filePath)
	if err != nil {
		return "", fmt.Errorf("加载凭证失败: %w", err)
	}

	// 统一校验凭证可用性
	status, err := auth.CheckAvailability(context.Background(), s.provider, creds)
	if err != nil {
		return statusDisable, nil
	}

	switch status {
	case credentials.StatusAvailable:
		// 如果凭证被刷新了，写回文件
		_ = auth.SaveCredentials(creds, filePath)
		return statusEnable, nil
	case credentials.StatusUsageLimited:
		return statusLimit, nil
	default:
		// StatusInvalidated, StatusBanned, StatusExpired, StatusReauthRequired 等均视为失效
		return statusDisable, nil
	}
}

// ensureDirs 确保目标目录及其子目录存在，如果子目录已存在且有文件则先清空
func ensureDirs(destDir string) error {
	dirs := []string{
		filepath.Join(destDir, string(statusEnable)),
		filepath.Join(destDir, string(statusLimit)),
		filepath.Join(destDir, string(statusDisable)),
	}
	for _, dir := range dirs {
		// 如果子目录已存在，先删除（连同其中的文件一起清除）
		if _, err := os.Stat(dir); err == nil {
			log.Infof("\n子目录 %q 已存在，清空旧文件...", dir)
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("删除目录 %q 失败: %w", dir, err)
			}
		}
		// 重新创建空目录
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %q 失败: %w", dir, err)
		}
	}
	return nil
}

// copyFile 拷贝文件
func copyFile(srcFile, destFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	return dst.Sync()
}
