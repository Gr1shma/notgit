package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

func LoadConfig() (*ini.File, string, error) {
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

	repoConfigPath := filepath.Join(rootRepoPath, ".notgit", "config")

	configRawContent, err := ReadFile(repoConfigPath)
	if err != nil {
		return nil, "", fmt.Errorf("error reading .notgit/config: %w", err)
	}

	configData, err := ini.Load(configRawContent)
	if err != nil {
		return nil, "", fmt.Errorf("invalid format in .notgit/config: %w", err)
	}

	return configData, repoConfigPath, nil
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

func SplitKey(key string) (section string, subkey string, err error) {
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
