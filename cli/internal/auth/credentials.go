package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nomand-zc/lumin/cli/internal/factory"
	"github.com/nomand-zc/lumin/credentials"
	kirocreds "github.com/nomand-zc/lumin/credentials/kiro"
	"github.com/nomand-zc/lumin/log"
	"github.com/nomand-zc/lumin/pool/taskpool"
	"github.com/nomand-zc/lumin/providers"
)

// LoadCredentials 从文件中加载凭证
func LoadCredentials(providerName string, file string) (credentials.Credential, error) {
	// 读取文件
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("读取凭证文件失败: %w", err)
	}

	var creds credentials.Credential

	switch providerName {
	case "kiro":
		creds = kirocreds.NewCredential(raw)
	default:
		return nil, fmt.Errorf("不支持的 provider: %q，支持的 provider 列表：%v", providerName, factory.SupportedProviders)
	}

	// 验证凭证，但允许过期凭证（ErrExpiresAtExpired 错误不视为失败）
	if err = creds.Validate(); err != nil {
		return creds, fmt.Errorf("验证凭证失败: %w", err)
	}

	return creds, nil
}

// SaveCredentials 将凭证保存到文件
func SaveCredentials(creds credentials.Credential, file string) error {
	if creds == nil {
		return fmt.Errorf("凭证不能为空")
	}
	if err := creds.Validate(); err != nil {
		return fmt.Errorf("验证凭证失败: %w", err)
	}

	// 将刷新后的凭证写回文件
	credsJSON, err := json.MarshalIndent(creds, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化凭证失败: %w", err)
	}
	if err := os.WriteFile(file, credsJSON, 0655); err != nil {
		return fmt.Errorf("写入凭证文件失败: %w", err)
	}

	return nil
}

// GetValidCredentials 获取有效的凭证
func GetValidCredentials(provider providers.Provider, dirPath string) (credentials.Credential, error) {
	// 检查credFile是文件还是目录
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("无法访问凭证路径: %w", err)
	}
	if !fileInfo.IsDir() {
		// 如果是文件，直接使用
		creds, status, err := GetCredentialsFromFile(provider, dirPath)
		if err != nil {
			return nil, err
		}
		if status != credentials.StatusAvailable {
			return nil, fmt.Errorf("凭证不可用，状态: %s", status)
		}
		return creds, nil
	}

	// 使用 taskpool 并发查找第一个有效凭证
	var finalCreds credentials.Credential
	var foundOnce sync.Once
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 如果已经找到有效凭证，跳过剩余文件
		select {
		case <-ctx.Done():
			return filepath.SkipAll
		default:
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			return nil
		}

		wg.Add(1)
		if err := taskpool.DefaultPool().Submit(func() {
			defer wg.Done()
			// 检查是否已找到有效凭证，避免无效网络请求
			select {
			case <-ctx.Done():
				return
			default:
			}
			cred, status, err := GetCredentialsFromFile(provider, path)
			if err != nil || status != credentials.StatusAvailable {
				return
			}
			// 使用 sync.Once 保证只设置一次结果
			foundOnce.Do(func() {
				finalCreds = cred
				cancel()
			})
		}); err != nil {
			wg.Done()
			log.Errorf("提交凭证校验任务失败: %v", err)
		}
		return nil
	})

	// 等待所有已提交的并发任务完成
	wg.Wait()

	if err != nil {
		return nil, err
	}

	return finalCreds, nil
}

// GetCredentialsFromFile 从文件加载凭证并通过 CheckAvailability 校验可用性。
// 如果凭证过期会自动刷新并写回文件。
func GetCredentialsFromFile(provider providers.Provider, file string) (credentials.Credential, credentials.CredentialStatus, error) {
	// 构建凭证
	creds, err := LoadCredentials(provider.Name(), file)
	if err != nil {
		return nil, credentials.StatusInvalidated, fmt.Errorf("构建凭证失败: %w", err)
	}

	// 记录刷新前的过期状态，用于判断是否需要写回文件
	wasExpired := creds.IsExpired()

	// 使用 CheckAvailability 统一进行过期刷新 + 用量校验
	status, err := provider.CheckAvailability(context.Background(), creds)
	if err != nil {
		return creds, status, err
	}

	// 如果之前过期且现在已恢复可用，说明 CheckAvailability 内部做了刷新，写回文件
	if wasExpired && status == credentials.StatusAvailable {
		if err := SaveCredentials(creds, file); err != nil {
			return creds, status, fmt.Errorf("保存刷新后的凭证失败: %w", err)
		}
		log.Infof("凭证刷新成功，已更新到文件: %s", file)
	}

	return creds, status, nil
}
