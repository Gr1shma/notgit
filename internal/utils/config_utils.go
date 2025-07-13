package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

func LoadConfig(useGlobal bool) (*ini.File, string, error) {
	var configPath string

	if useGlobal {
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("HOME"), ".config")
		}

		absConfigDir, err := filepath.Abs(configDir)
		if err != nil {
			return nil, "", fmt.Errorf("error resolving config directory: %w", err)
		}

		globalConfigDir := filepath.Join(absConfigDir, "notgit")
		configPath = filepath.Join(globalConfigDir, "config")

		if err := os.MkdirAll(globalConfigDir, 0o755); err != nil {
			return nil, "", fmt.Errorf("failed to create config directory: %w", err)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			file, err := os.Create(configPath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to create config file: %w", err)
			}
			file.Close()
		}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, "", fmt.Errorf("error getting current working directory: %w", err)
		}

		rootRepoPath, err := FindRepoRoot(cwd)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, "", fmt.Errorf("not inside a notgit repository")
			}
			return nil, "", fmt.Errorf("error locating notgit repository: %w", err)
		}

		configPath = filepath.Join(rootRepoPath, ".notgit", "config")
	}

	configData, err := ini.Load(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load config file %s: %w", configPath, err)
	}

	return configData, configPath, nil
}

type ConfigKeyInfo struct {
	Description string
}

var configSchema = map[string]map[string]ConfigKeyInfo{
	"user": {
		"name": {
			Description: "User's name (e.g., Grishma Dhakal)",
		},
		"email": {
			Description: "User's email address (e.g., grishma@example.com)",
		},
	},
	"core": {
		"editor": {
			Description: "Default text editor (e.g., vim, nvim, nano)",
		},
	},
	"init": {
		"defaultBranch": {
			Description: "Default branch name for new repositories (e.g., main)",
		},
	},
}

func SplitConfigKey(key string) (section string, subkey string, err error) {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid key format: %q (expected section.key)", key)
	}

	section, subkey = parts[0], parts[1]

	if strings.EqualFold(section, ini.DefaultSection) {
		return "", "", fmt.Errorf("section %q is reserved and not allowed", section)
	}

	subkeys, ok := configSchema[section]
	if !ok {
		return "", "", fmt.Errorf("unsupported section: %s", section)
	}

	if _, ok := subkeys[subkey]; !ok {
		return "", "", fmt.Errorf("unsupported key: %s.%s", section, subkey)
	}

	return section, subkey, nil
}

func PrintSupportedConfigKeys() string {
	var sb strings.Builder
	sb.WriteString("\nSupported configuration keys:\n")
	for section, keys := range configSchema {
		for key, info := range keys {
			sb.WriteString(fmt.Sprintf("  %s.%s\t- %s\n", section, key, info.Description))
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

func SaveConfig(cfg *ini.File, path string) error {
	return cfg.SaveTo(path)
}

func GetConfigKeyValue(cfg *ini.File, key string) (string, error) {
	section, subkey, err := SplitConfigKey(key)
	if err != nil {
		return "", err
	}
	val := cfg.Section(section).Key(subkey).String()
	if val == "" {
		return "", fmt.Errorf("key %s not set", key)
	}
	return val, nil
}

func SetConfigKeyValue(cfg *ini.File, path, key, value string) error {
	section, subkey, err := SplitConfigKey(key)
	if err != nil {
		return err
	}

	cfg.Section(section).Key(subkey).SetValue(value)

	return SaveConfig(cfg, path)
}

func UnsetConfigKey(cfg *ini.File, path, key string) (bool, error) {
	section, subkey, err := SplitConfigKey(key)
	if err != nil {
		return false, err
	}

	sec := cfg.Section(section)
	if sec == nil || !sec.HasKey(subkey) {
		return false, fmt.Errorf("key %s is not set", key)
	}

	sec.DeleteKey(subkey)

	// Remove section if empty and not default
	if len(sec.Keys()) == 0 && sec.Name() != ini.DefaultSection {
		cfg.DeleteSection(section)
	}

	if err := SaveConfig(cfg, path); err != nil {
		return false, err
	}

	return true, nil
}

func PrintAllConfig(cfg *ini.File) {
	for _, section := range cfg.Sections() {
		if section.Name() == ini.DefaultSection && len(section.Keys()) == 0 {
			continue
		}
		for _, key := range section.Keys() {
			if section.Name() == ini.DefaultSection {
				fmt.Printf("%s = %s\n", key.Name(), key.Value())
			} else {
				fmt.Printf("%s.%s = %s\n", section.Name(), key.Name(), key.Value())
			}
		}
	}
}
