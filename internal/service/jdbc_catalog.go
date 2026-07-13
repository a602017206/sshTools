package service

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"AHaSSHTools/internal/config"
)

//go:embed jdbc_builtin_manifest.json
var builtinJDBCManifest []byte

type DriverCatalogService struct {
	manifestPath  string
	installedPath string
}

func NewDriverCatalogService(manifestPath, installedPath string) *DriverCatalogService {
	return &DriverCatalogService{manifestPath: manifestPath, installedPath: installedPath}
}

func (s *DriverCatalogService) LoadManifest() (*config.JDBCManifest, error) {
	if err := s.ensureManifest(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(s.manifestPath)
	if err != nil {
		return nil, fmt.Errorf("读取 JDBC 驱动清单失败: %w", err)
	}
	var manifest config.JDBCManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("解析 JDBC 驱动清单失败: %w", err)
	}
	return &manifest, nil
}

func (s *DriverCatalogService) ensureManifest() error {
	var builtin config.JDBCManifest
	if err := json.Unmarshal(builtinJDBCManifest, &builtin); err != nil {
		return fmt.Errorf("解析内置 JDBC 驱动清单失败: %w", err)
	}

	data, err := os.ReadFile(s.manifestPath)
	if os.IsNotExist(err) {
		return s.writeManifest(&builtin)
	}
	if err != nil {
		return fmt.Errorf("检查 JDBC 驱动清单失败: %w", err)
	}
	var current config.JDBCManifest
	if err := json.Unmarshal(data, &current); err != nil {
		return fmt.Errorf("解析 JDBC 驱动清单失败: %w", err)
	}
	if current.Version >= builtin.Version {
		return nil
	}
	return s.writeManifest(mergeJDBCManifest(builtin, current))
}

func mergeJDBCManifest(builtin, current config.JDBCManifest) *config.JDBCManifest {
	merged := builtin
	currentByID := make(map[string]config.JDBCDriver, len(current.Drivers))
	for _, driver := range current.Drivers {
		currentByID[driver.ID] = driver
	}
	for i := range merged.Drivers {
		currentDriver, ok := currentByID[merged.Drivers[i].ID]
		if !ok {
			continue
		}
		knownProfiles := make(map[string]struct{}, len(merged.Drivers[i].Profiles))
		for _, profile := range merged.Drivers[i].Profiles {
			knownProfiles[profile.ID] = struct{}{}
		}
		for _, profile := range currentDriver.Profiles {
			if _, exists := knownProfiles[profile.ID]; !exists {
				merged.Drivers[i].Profiles = append(merged.Drivers[i].Profiles, profile)
			}
		}
		delete(currentByID, merged.Drivers[i].ID)
	}
	for _, driver := range current.Drivers {
		if _, custom := currentByID[driver.ID]; custom {
			merged.Drivers = append(merged.Drivers, driver)
		}
	}
	return &merged
}

func (s *DriverCatalogService) writeManifest(manifest *config.JDBCManifest) error {
	if err := os.MkdirAll(filepath.Dir(s.manifestPath), 0o700); err != nil {
		return fmt.Errorf("创建 JDBC 驱动清单目录失败: %w", err)
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JDBC 驱动清单失败: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(s.manifestPath), ".jdbc-manifest-*.tmp")
	if err != nil {
		return fmt.Errorf("创建 JDBC 驱动清单临时文件失败: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("写入 JDBC 驱动清单失败: %w", err)
	}
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("设置 JDBC 驱动清单权限失败: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("关闭 JDBC 驱动清单失败: %w", err)
	}
	if err := os.Rename(temporaryPath, s.manifestPath); err != nil {
		return fmt.Errorf("提交 JDBC 驱动清单失败: %w", err)
	}
	return nil
}

func (s *DriverCatalogService) GetRecommendedProfile(driverID string) (*config.JDBCDriver, *config.JDBCDriverProfile, error) {
	manifest, err := s.LoadManifest()
	if err != nil {
		return nil, nil, err
	}
	for i := range manifest.Drivers {
		driver := &manifest.Drivers[i]
		if driver.ID != driverID {
			continue
		}
		for j := range driver.Profiles {
			profile := &driver.Profiles[j]
			if profile.Version == driver.RecommendedVersion || profile.ID == driver.RecommendedVersion {
				return driver, profile, nil
			}
		}
		if len(driver.Profiles) > 0 {
			return driver, &driver.Profiles[0], nil
		}
		return nil, nil, fmt.Errorf("数据库 %s 没有可用 JDBC profile", driverID)
	}
	return nil, nil, fmt.Errorf("未找到 JDBC 驱动: %s", driverID)
}

func (s *DriverCatalogService) GetProfile(driverID, version string) (*config.JDBCDriver, *config.JDBCDriverProfile, error) {
	if version == "" {
		return s.GetRecommendedProfile(driverID)
	}
	manifest, err := s.LoadManifest()
	if err != nil {
		return nil, nil, err
	}
	for i := range manifest.Drivers {
		driver := &manifest.Drivers[i]
		if driver.ID != driverID {
			continue
		}
		for j := range driver.Profiles {
			profile := &driver.Profiles[j]
			if profile.Version == version || profile.ID == version {
				return driver, profile, nil
			}
		}
		return nil, nil, fmt.Errorf("未找到 JDBC 驱动版本: %s %s", driverID, version)
	}
	return nil, nil, fmt.Errorf("未找到 JDBC 驱动: %s", driverID)
}

func (s *DriverCatalogService) ListDriversWithInstallStatus() ([]DriverView, error) {
	manifest, err := s.LoadManifest()
	if err != nil {
		return nil, err
	}
	drivers := make([]DriverView, 0, len(manifest.Drivers))
	for _, driver := range manifest.Drivers {
		profiles := make([]config.JDBCDriverProfile, len(driver.Profiles))
		copy(profiles, driver.Profiles)
		installed := false
		for i := range profiles {
			if s.profileInstalled(driver.ID, profiles[i].Version) {
				profiles[i].Installed = true
				profiles[i].InstallPath = filepath.Join(s.installedPath, driver.ID, profiles[i].Version)
				installed = true
			}
		}
		drivers = append(drivers, DriverView{
			ID:                 driver.ID,
			Name:               driver.Name,
			RecommendedVersion: driver.RecommendedVersion,
			Installed:          installed,
			Profiles:           profiles,
		})
	}
	return drivers, nil
}

func (s *DriverCatalogService) profileInstalled(driverID, version string) bool {
	if s.installedPath == "" || driverID == "" || version == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(s.installedPath, driverID, version))
	return err == nil && info.IsDir()
}
