package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// KamiwazaDeployment represents a model deployment in Kamiwaza
type KamiwazaDeployment struct {
	ID           string `json:"id"`
	ModelName    string `json:"m_name"`
	ConfigName   string `json:"m_config_name"`
	Status       string `json:"status"`
	LBPort       int    `json:"lb_port"`
	ServePath    string `json:"serve_path"`
	Engine       string `json:"engine"`
	DeployedAt   string `json:"deployed_at"`
}

// KamiwazaAuthResponse represents the token response from Kamiwaza
type KamiwazaAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// KamiwazaService handles interactions with Kamiwaza API
type KamiwazaService struct {
	baseURL  string
	client   *http.Client
	username string
	password string
	token    string
}

// NewKamiwazaService creates a new Kamiwaza service instance with authentication
// Default credentials are admin/kamiwaza
func NewKamiwazaService(baseURL string) *KamiwazaService {
	if baseURL == "" {
		baseURL = "https://localhost"
	}

	// Create HTTP client with TLS verification disabled for self-signed certs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &KamiwazaService{
		baseURL:  baseURL,
		username: "admin",
		password: "kamiwaza",
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
	}
}

// authenticate obtains an access token from Kamiwaza
func (k *KamiwazaService) authenticate() error {
	authURL := fmt.Sprintf("%s/api/auth/token", k.baseURL)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", k.username)
	data.Set("password", k.password)
	data.Set("scope", "")
	data.Set("client_id", "string")
	data.Set("client_secret", "********")

	req, err := http.NewRequest("POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := k.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp KamiwazaAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	k.token = authResp.AccessToken
	return nil
}

// ensureAuthenticated checks if we have a token and authenticates if needed
func (k *KamiwazaService) ensureAuthenticated() error {
	if k.token == "" {
		return k.authenticate()
	}
	return nil
}

// ListDeployments retrieves all deployments from Kamiwaza
func (k *KamiwazaService) ListDeployments() ([]KamiwazaDeployment, error) {
	// Ensure we have a valid token
	if err := k.ensureAuthenticated(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	url := fmt.Sprintf("%s/api/serving/deployments", k.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Bearer token authentication
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", k.token))
	req.Header.Set("Accept", "application/json")

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var deployments []KamiwazaDeployment
	if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return deployments, nil
}

// GetActiveDeployments returns only deployments with status "DEPLOYED"
func (k *KamiwazaService) GetActiveDeployments() ([]KamiwazaDeployment, error) {
	deployments, err := k.ListDeployments()
	if err != nil {
		return nil, err
	}

	var active []KamiwazaDeployment
	for _, d := range deployments {
		if d.Status == "DEPLOYED" {
			active = append(active, d)
		}
	}

	return active, nil
}

// GetDeploymentByModelName finds a deployment by model name and returns its endpoint info
func (k *KamiwazaService) GetDeploymentByModelName(modelName string) (*KamiwazaDeployment, error) {
	deployments, err := k.GetActiveDeployments()
	if err != nil {
		return nil, err
	}

	for _, d := range deployments {
		if d.ModelName == modelName {
			return &d, nil
		}
	}

	return nil, fmt.Errorf("no active deployment found for model: %s", modelName)
}

// GetModelEndpoint returns the base URL for a specific model deployment
// Format: https://localhost:{lb_port}
func (k *KamiwazaService) GetModelEndpoint(modelName string) (string, error) {
	deployment, err := k.GetDeploymentByModelName(modelName)
	if err != nil {
		return "", err
	}

	// Extract host from baseURL (remove https:// or http://)
	host := k.baseURL
	if len(host) > 8 && host[:8] == "https://" {
		host = host[8:]
	} else if len(host) > 7 && host[:7] == "http://" {
		host = host[7:]
	}

	return fmt.Sprintf("https://%s:%d", host, deployment.LBPort), nil
}

// GetModelIdentifier returns the model identifier to use in API requests
// For Kamiwaza, this is simply "model"
func (k *KamiwazaService) GetModelIdentifier() string {
	return "model"
}
