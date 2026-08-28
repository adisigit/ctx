package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type repoLink struct {
	RepoID  string `json:"repo_id"`
	Company string `json:"company"`
}

type repoLinkStore struct {
	Links []repoLink `json:"links"`
}

func repoLinkPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "repo_links.json"), nil
}

func loadRepoLinks() (repoLinkStore, error) {
	var store repoLinkStore
	path, err := repoLinkPath()
	if err != nil {
		return store, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return store, err
	}
	err = json.Unmarshal(data, &store)
	if err != nil {
		return store, err
	}
	return store, nil
}

func saveRepoLinks(store repoLinkStore) error {
	dir, _ := Dir()
	os.MkdirAll(dir, 0700)
	path, err := repoLinkPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func CompanyForRepo(repoID string) (*CompanyConfig, error) {
	links, err := loadRepoLinks()
	if err != nil {
		return nil, err
	}
	var companyName string
	for _, link := range links.Links {
		if link.RepoID == repoID {
			companyName = link.Company
			break
		}
	}
	if companyName == "" {
		return nil, os.ErrNotExist
	}
	store, err := loadCompanyStore()
	if err != nil {
		return nil, err
	}
	for _, company := range store.Companies {
		if company.Name == companyName {
			return &company, nil
		}
	}
	return nil, os.ErrNotExist
}

func LinkRepoToCompany(repoID, companyName string) error {
	links, _ := loadRepoLinks()
	for i, link := range links.Links {
		if link.RepoID == repoID {
			links.Links[i].Company = companyName
			return saveRepoLinks(links)
		}
	}
	links.Links = append(links.Links, repoLink{RepoID: repoID, Company: companyName})
	return saveRepoLinks(links)
}
