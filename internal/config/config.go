package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ctx"), nil
}

type CompanyConfig struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
}

type companyStore struct {
	Companies []CompanyConfig `json:"companies"`
	Active    string          `json:"active"`
}

func companyStorePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "company.json"), nil
}

func loadCompanyStore() (companyStore, error) {
	var store companyStore
	path, err := companyStorePath()
	if err != nil {
		return store, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return store, err
	}
	if err := json.Unmarshal(data, &store); err != nil {
		return store, err
	}
	return store, nil
}

func saveCompanyStore(store companyStore) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path, err := companyStorePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func ListCompanies() ([]CompanyConfig, error) {
	store, err := loadCompanyStore()
	if err != nil {
		return nil, err
	}
	return store.Companies, nil
}

func ActiveCompany() (*CompanyConfig, error) {
	store, err := loadCompanyStore()
	if err != nil {
		return nil, err
	}
	for _, company := range store.Companies {
		if company.Name == store.Active {
			return &company, nil
		}
	}
	return nil, os.ErrNotExist
}

func UpsertCompany(cfg CompanyConfig) error {
	store, err := loadCompanyStore()
	if err != nil {
		store = companyStore{}
	}
	replaced := false
	for i, company := range store.Companies {
		if company.Name == cfg.Name {
			store.Companies[i] = cfg
			replaced = true
		}
	}
	if !replaced {
		store.Companies = append(store.Companies, cfg)
	}
	if cfg.Name != "" {
		store.Active = cfg.Name
	}
	return saveCompanyStore(store)
}

func SetActiveCompany(name string) (*CompanyConfig, error) {
	store, err := loadCompanyStore()
	if err != nil {
		return nil, err
	}
	for _, company := range store.Companies {
		if company.Name == name {
			store.Active = name
			return &company, saveCompanyStore(store)
		}
	}
	return nil, os.ErrNotExist
}

func tokenPath(companyName string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	safe := sanitizeFileName(companyName)
	return filepath.Join(dir, "token_"+safe), nil
}

func SaveToken(companyName, token string) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path, err := tokenPath(companyName)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(token), 0600)
}

func LoadToken(companyName string) (string, error) {
	path, err := tokenPath(companyName)
	if err != nil {
		return "", err
	}
	token, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func ClearToken(companyName string) error {
	path, err := tokenPath(companyName)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func sanitizeFileName(name string) string {
	out := make([]byte, 0, len(name))
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			out = append(out, byte(r))
		} else {
			out = append(out, '_')
		}
	}
	return string(out)
}
