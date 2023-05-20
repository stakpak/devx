package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/browser"
	"gopkg.in/yaml.v3"
)

const DEVX_CLOUD_ENDPOINT = "https://hub.stakpak.dev"

type ServerConfig struct {
	Disable  bool
	Tenant   string
	Endpoint string
}

type ConfigFile map[string]Config

type Config struct {
	Default        bool       `yaml:"default,omitempty"`
	Endpoint       *string    `yaml:"endpoint,omitempty"`
	Username       *string    `yaml:"username,omitempty"`
	Password       *string    `yaml:"password,omitempty"`
	Token          *string    `yaml:"token,omitempty"`
	TokenExpiresOn *time.Time `yaml:"tokenExpiresOn,omitempty"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   uint   `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
	TokenType   string `json:"token_type"`
}

func IsLoggedIn(server ServerConfig) bool {
	var cfg Config

	if server.Disable {
		return false
	}

	if server.Tenant == "" {
		tenant, _, err := loadDefaultConfig()
		if err != nil {
			return false
		}
		server.Tenant = tenant
	}

	cfgFile, err := loadConfig()
	if err != nil {
		return false
	}

	if _, ok := cfgFile[server.Tenant]; ok {
		cfg = cfgFile[server.Tenant]
	}

	if cfg.TokenExpiresOn != nil && cfg.TokenExpiresOn.After(time.Now()) {
		return true
	}

	log.Info("Found Hub configurations but without a valid token, proceeding offline\nTry again after logging in using: devx login --tenant <your tenant>")

	return false
}

func GetToken(server ServerConfig) (*string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	if _, ok := cfg[server.Tenant]; !ok {
		err = Login(server)
		if err != nil {
			return nil, err
		}
	}
	return cfg[server.Tenant].Token, nil
}

func GetDefaultToken() (string, *string, error) {
	tenant, cfg, err := loadDefaultConfig()
	if err != nil {
		return "", nil, err
	}
	if tenant == "" {
		return "", nil, nil
	}

	return tenant, cfg.Token, nil
}

func Clear(server ServerConfig) error {
	return deleteConfig()
}

func Info(server ServerConfig) error {
	var cfg Config

	if server.Tenant == "" {
		tenant, _, err := loadDefaultConfig()
		if err != nil {
			return err
		}
		server.Tenant = tenant
	}

	cfgFile, err := loadConfig()
	if err != nil {
		return err
	}

	if _, ok := cfgFile[server.Tenant]; ok {
		cfg = cfgFile[server.Tenant]
	}

	log.Infof(`Auth info:
	Tenant: %s
	Server endpoint: %s
	Token valid until: %s
	`,
		server.Tenant,
		*cfg.Endpoint,
		cfg.TokenExpiresOn,
	)

	return nil
}

func Login(server ServerConfig) error {
	if server.Endpoint == "" {
		server.Endpoint = DEVX_CLOUD_ENDPOINT
	}

	if server.Tenant == "" {
		return fmt.Errorf("--tenant <tenant name> is required")
	}

	cfgFile, err := loadConfig()
	if err != nil {
		return err
	}

	cfg := Config{}
	if _, ok := cfgFile[server.Tenant]; ok {
		cfg = cfgFile[server.Tenant]
	}

	if len(cfgFile) == 0 {
		cfg.Default = true
	}

	if cfg.Endpoint == nil {
		cfg.Endpoint = &server.Endpoint
	}

	if cfg.Token != nil && cfg.TokenExpiresOn != nil && cfg.TokenExpiresOn.After(time.Now()) {
		log.Info("Already logged in")
		return nil
	}

	if cfg.Username != nil && cfg.Password != nil {
		log.Info("Non interactive login")
		return nil
	}

	loginURL := fmt.Sprintf(
		"%s/login?redirect_uri=%s&tenant=%s",
		*cfg.Endpoint,
		url.QueryEscape("http://localhost:17777/callback"),
		url.QueryEscape(server.Tenant),
	)

	browser.Stderr = io.Discard
	browser.Stdout = io.Discard
	err = browser.OpenURL(loginURL)
	if err != nil {
		return err
	}
	log.Info("Check your browser to complete the login flow")
	log.Info("Or paste this in your browser")
	log.Info("\t", loginURL)

	codeCh := make(chan string, 1)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Headers", "content-type")
		w.Header().Add("Access-Control-Allow-Credentials", "true")

		queryCode := r.URL.Query().Get("code")
		if r.Method == "GET" && queryCode != "" {
			codeCh <- queryCode
			close(codeCh)
			http.Redirect(w, r, fmt.Sprintf("%s/cli-success", *cfg.Endpoint), http.StatusTemporaryRedirect)
		}

		_, err := w.Write([]byte{})
		if err != nil {
			log.Error(err)
		}
	})
	go func() {
		err = http.ListenAndServe("localhost:17777", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	code := <-codeCh
	log.Info("Received auth code")

	resp, err := http.PostForm(
		fmt.Sprintf("%s/oauth2/token", *cfg.Endpoint),
		url.Values{
			"code":         {code},
			"redirect_uri": {"http://localhost:17777/callback"},
			"grant_type":   {"authorization_code"},
			"client_id":    {"devx-server"},
		},
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	tokenResp := TokenResponse{}
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return err
	}

	log.Info("Logged in successfully")

	expiresOn := time.Now().Add(time.Second * time.Duration(tokenResp.ExpiresIn))

	cfg.Token = &tokenResp.AccessToken
	cfg.TokenExpiresOn = &expiresOn

	cfgFile[server.Tenant] = cfg
	err = saveConfig(cfgFile)
	if err != nil {
		return err
	}

	return nil
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".devx"), nil
}

func loadDefaultConfig() (string, Config, error) {
	cfgFile, err := loadConfig()
	if err != nil {
		return "", Config{}, err
	}
	for tenant, cfg := range cfgFile {
		if cfg.Default {
			return tenant, cfg, nil
		}
	}
	return "", Config{}, fmt.Errorf("no credentials found\nTry logging in using: devx login --tenant <your tenant>")
}

func loadConfig() (ConfigFile, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return ConfigFile{}, nil
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg := ConfigFile{}
	err = yaml.Unmarshal(configData, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, err
}

func saveConfig(cfg ConfigFile) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0700)
		if err != nil {
			return err
		}
	}

	configPath := filepath.Join(configDir, "config")
	configData, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, configData, 0700)
}

func deleteConfig() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0700)
		if err != nil {
			return err
		}
	}

	configPath := filepath.Join(configDir, "config")

	if err := os.Remove(configPath); err != nil {
		return err
	}

	return nil
}
