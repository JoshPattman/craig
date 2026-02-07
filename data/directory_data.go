package data

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/JoshPattman/react"
	"github.com/adrg/frontmatter"
)

func NewDirectoryData(root string) *DirectoryData {
	return &DirectoryData{root}
}

type DirectoryData struct {
	root string
}

func (dd *DirectoryData) GetScratchPad() ScratchPad {
	return &fileScratchPad{
		path.Join(dd.root, "scratchpad.txt"),
	}
}

type fileScratchPad struct {
	filepath string
}

func (s *fileScratchPad) Content() (string, error) {
	content, err := os.ReadFile(s.filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (s *fileScratchPad) Rewrite(oldText, newText string) error {
	content, err := os.ReadFile(s.filepath)
	if err != nil {
		return err
	}
	n := strings.Count(string(content), oldText)
	if n == 0 {
		return ErrOldTextNotFound
	}
	if n > 1 {
		return ErrOldTextAmbiguous
	}
	newContent := strings.ReplaceAll(string(content), oldText, newText)
	err = os.WriteFile(s.filepath, []byte(newContent), os.ModePerm)
	return nil
}

func (dd *DirectoryData) GetSkillset() Skillset {
	return &mdcDirectorySkillset{
		path.Join(dd.root, "skills"),
	}
}

type mdcDirectorySkillset struct {
	root string
}

func (m *mdcDirectorySkillset) List() ([]react.Skill, error) {
	return loadMDCSkills(m.root)
}

func loadMDCSkills(dir string) ([]react.Skill, error) {
	var skills []react.Skill

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".mdc" {
			return nil
		}

		skill, err := loadMDCSkill(path)
		if err != nil {
			return err
		}

		skills = append(skills, skill)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return skills, nil
}

type mdcFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Always      bool   `yaml:"always"`
}

func loadMDCSkill(skillPath string) (react.Skill, error) {
	f, err := os.Open(skillPath)
	if err != nil {
		return react.Skill{}, err
	}
	defer f.Close()
	var meta mdcFrontmatter
	body, err := frontmatter.Parse(f, &meta)
	if err != nil {
		return react.Skill{}, err
	}
	if meta.Always {
		meta.Description = ""
	}
	return react.Skill{
		Key:     meta.Name,
		When:    meta.Description,
		Content: string(body),
	}, nil
}

func (dd *DirectoryData) Init(fileSystem fs.FS, rootDirName string) error {
	err := os.RemoveAll(dd.root)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dd.root, 0755)
	if err != nil {
		return err
	}
	srcDir := rootDirName
	return fs.WalkDir(fileSystem, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(dd.root, relPath)

		if d.IsDir() {
			return os.MkdirAll(outPath, 0755)
		}

		data, err := fs.ReadFile(fileSystem, path)
		if err != nil {
			return err
		}

		return os.WriteFile(outPath, data, 0644)
	})
}

func (dd *DirectoryData) AgentModel() (ModelSetup, error) {
	return loadModelSetup(path.Join(dd.root, "models", "agent.json"))
}

func (dd *DirectoryData) FilterModel() (ModelSetup, error) {
	return loadModelSetup(path.Join(dd.root, "models", "filter.json"))
}

func (dd *DirectoryData) Personality() (string, error) {
	data, err := os.ReadFile(path.Join(dd.root, "personality.txt"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func loadModelSetup(fp string) (ModelSetup, error) {
	f, err := os.Open(fp)
	if err != nil {
		return ModelSetup{}, err
	}
	defer f.Close()
	var result ModelSetup
	err = json.NewDecoder(f).Decode(&result)
	if err != nil {
		return ModelSetup{}, err
	}
	return result, nil
}

type mcpConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Enabled bool              `json:"enabled"`
}

func (dd *DirectoryData) EnabledTools() ([]react.Tool, error) {
	var allTools []react.Tool

	// Directory where MCP configs live
	configDir := filepath.Join(dd.root, "mcp")

	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read mcp directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only load .json files
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(configDir, entry.Name())

		// Read file
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Parse JSON config
		var cfg mcpConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Skip disabled servers
		if !cfg.Enabled {
			continue
		}

		// Connect MCP
		mcp, err := connectMCP(cfg.URL, cfg.Headers)
		if err != nil {
			return nil, fmt.Errorf("failed to connect MCP %s: %w", cfg.URL, err)
		}

		// Convert MCP tools
		mcpTools, err := createToolsFromMCP(mcp)
		if err != nil {
			return nil, fmt.Errorf("failed creating tools from MCP %s: %w", cfg.URL, err)
		}

		// Add to global list
		allTools = append(allTools, mcpTools...)
	}

	return allTools, nil
}
