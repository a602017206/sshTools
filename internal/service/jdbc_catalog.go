package service

import (
	"encoding/json"
	"fmt"
	"os"

	"AHaSSHTools/internal/config"
)

type DriverCatalogService struct {
	manifestPath  string
	installedPath string
}

func NewDriverCatalogService(manifestPath, installedPath string) *DriverCatalogService {
	return &DriverCatalogService{manifestPath: manifestPath, installedPath: installedPath}
}

func (s *DriverCatalogService) LoadManifest() (*config.JDBCManifest, error) {
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
