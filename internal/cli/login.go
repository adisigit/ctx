package cli

import (
	"bufio"
	"ctx/internal/auth"
	"ctx/internal/config"
	"ctx/internal/git"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func ApiBase() string {
	if id, err := git.RepoIdentity(); err == nil {
		if cfg, err := config.CompanyForRepo(id); err == nil {
			return cfg.BaseURL
		}
	}
	if cfg, err := config.ActiveCompany(); err == nil {
		return cfg.BaseURL
	}
	return "http://localhost:8080"
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your account and link this device",
	RunE:  runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}
	company, err := ensureCompanyConfig()
	if err != nil {
		return err
	}
	if token, err := config.LoadToken(company.Name); err == nil && token != "" {
		valid, verr := auth.VerifyToken(company.BaseURL, token)
		if verr == nil && valid {
			fmt.Println("Logged in successfully!")
			return nil
		}
		if verr != nil {
			return verr
		}
	}
	fmt.Println("Generating device key...")
	kp, err := auth.GenerateKeyPair()
	if err != nil {
		return err
	}
	if err := kp.SaveToDisk(dir); err != nil {
		return err
	}
	client := auth.NewSessionClient(ApiBase())
	fmt.Println("Requesting login session...")
	sessionID, loginURL, err := client.CreateSession(kp.PublicKeyB64())
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("Open this URL in your browser to login:")
	fmt.Println(loginURL)
	fmt.Println()
	if err := openBrowser(loginURL); err != nil {
		return err
	}
	fmt.Println("Waiting for login...")
	encryptedToken, err := client.PollSession(sessionID, 5*time.Minute)
	if err != nil {
		return err
	}
	plaintext, err := auth.UpenSealedBox(kp.PrivateKey, encryptedToken)
	if err != nil {
		return err
	}
	if err := config.SaveToken(company.Name, string(plaintext)); err != nil {
		return err
	}
	fmt.Println("Logged in successfully!")
	return nil
}

func ensureCompanyConfig() (*config.CompanyConfig, error) {
	repoID, repoErr := git.RepoIdentity()
	if repoErr == nil {
		if cfg, err := config.CompanyForRepo(repoID); err == nil {
			fmt.Printf("Using company from repo link: %s\n", cfg.Name)
			return cfg, nil
		}
	}

	reader := bufio.NewReader(os.Stdin)
	companies, err := config.ListCompanies()

	var selectedCfg *config.CompanyConfig

	if err == nil && len(companies) > 0 {
		fmt.Println("Saved company")
		for i, company := range companies {
			fmt.Printf("%d: %s (%s)\n", i+1, company.Name, company.BaseURL)
		}
		fmt.Printf("[%d] add new company\n", len(companies)+1)
		fmt.Print("Select: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		idx, cerr := strconv.Atoi(choice)
		if cerr != nil || idx < 1 || idx > len(companies)+1 {
			fmt.Println("invalid selection")
			choice, _ = reader.ReadString('\n')
			choice = strings.TrimSpace(choice)
			idx, cerr = strconv.Atoi(choice)
		}
		if idx <= len(companies) {
			selected := companies[idx-1]
			if _, err := config.SetActiveCompany(selected.Name); err != nil {
				return nil, err
			}
			selectedCfg = &selected
		}
		fmt.Println()
	} else {
		fmt.Println("No saved companies.")
	}
	if selectedCfg == nil {
		fmt.Print("Company name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		for name == "" {
			fmt.Println("Name cannot be empty.")
			fmt.Print("Company name: ")
			name, _ = reader.ReadString('\n')
			name = strings.TrimSpace(name)
		}
		fmt.Print("Company URL: ")
		url, _ := reader.ReadString('\n')
		url = strings.TrimSpace(url)
		for url == "" {
			fmt.Println("URL cannot be empty.")
			fmt.Print("Company URL: ")
			url, _ = reader.ReadString('\n')
			url = strings.TrimSpace(url)
		}
		url = strings.TrimSuffix(url, "/")
		cfg := config.CompanyConfig{Name: name, BaseURL: url}
		if err := config.UpsertCompany(cfg); err != nil {
			return nil, err
		}
		fmt.Printf("Config for %s saved.\n\n", name)
		selectedCfg = &cfg
	}
	if repoErr == nil {
		if err := config.LinkRepoToCompany(repoID, selectedCfg.Name); err != nil {
			fmt.Println("Warning: gagal menyimpan repo link:", err)
		}
	}
	return selectedCfg, nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd, args = "xdg-open", []string{url}
	}

	return exec.Command(cmd, args...).Start()
}
