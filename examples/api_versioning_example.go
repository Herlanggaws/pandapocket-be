package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Example client demonstrating API versioning usage
func main() {
	baseURL := "http://localhost:8080"

	fmt.Println("=== PandaPocket API Versioning Example ===")

	// Example 1: Using latest version (v120)
	fmt.Println("\n1. Using latest version (v120):")
	useLatestVersion(baseURL)

	// Example 2: Using deprecated version (v100)
	fmt.Println("\n2. Using deprecated version (v100):")
	useDeprecatedVersion(baseURL)

	// Example 3: Version information
	fmt.Println("\n3. Getting version information:")
	getVersionInfo(baseURL)

	// Example 4: Migration recommendations
	fmt.Println("\n4. Getting migration recommendations:")
	getMigrationRecommendations(baseURL)
}

func useLatestVersion(baseURL string) {
	// Example: Get transactions using v120 (latest)
	url := baseURL + "/api/v120/transactions"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Add auth token (in real usage)
	req.Header.Set("Authorization", "Bearer your-token-here")
	req.Header.Set("X-API-Version", "v120")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check version headers
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("X-API-Version: %s\n", resp.Header.Get("X-API-Version"))
	fmt.Printf("X-API-Deprecated: %s\n", resp.Header.Get("X-API-Deprecated"))

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err == nil {
		fmt.Printf("Response: %+v\n", response)
	}
}

func useDeprecatedVersion(baseURL string) {
	// Example: Get transactions using v100 (deprecated)
	url := baseURL + "/api/v100/transactions"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Add auth token (in real usage)
	req.Header.Set("Authorization", "Bearer your-token-here")
	req.Header.Set("X-API-Version", "v100")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check deprecation headers
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("X-API-Version: %s\n", resp.Header.Get("X-API-Version"))
	fmt.Printf("X-API-Deprecated: %s\n", resp.Header.Get("X-API-Deprecated"))
	fmt.Printf("X-API-Sunset-Date: %s\n", resp.Header.Get("X-API-Sunset-Date"))
	fmt.Printf("X-API-Upgrade-URL: %s\n", resp.Header.Get("X-API-Upgrade-URL"))

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err == nil {
		fmt.Printf("Response: %+v\n", response)
	}
}

func getVersionInfo(baseURL string) {
	// Example: Get version information
	url := baseURL + "/api/version/info/v100"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err == nil {
		fmt.Printf("Version Info: %+v\n", response)
	}
}

func getMigrationRecommendations(baseURL string) {
	// Example: Get migration recommendations
	url := baseURL + "/api/version/recommendations/v100"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err == nil {
		fmt.Printf("Migration Recommendations: %+v\n", response)
	}
}

// Example of a client library that handles versioning
type PandaPocketClient struct {
	BaseURL string
	Version string
	Token   string
	Client  *http.Client
}

func NewPandaPocketClient(baseURL, version, token string) *PandaPocketClient {
	return &PandaPocketClient{
		BaseURL: baseURL,
		Version: version,
		Token:   token,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *PandaPocketClient) GetTransactions() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/%s/transactions", c.BaseURL, c.Version)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Version", c.Version)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for deprecation warnings
	if resp.Header.Get("X-API-Deprecated") == "true" {
		fmt.Printf("⚠️  Warning: API version %s is deprecated!\n", c.Version)
		fmt.Printf("   Sunset date: %s\n", resp.Header.Get("X-API-Sunset-Date"))
		fmt.Printf("   Upgrade URL: %s\n", resp.Header.Get("X-API-Upgrade-URL"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	return response, err
}

func (c *PandaPocketClient) CreateTransaction(transaction map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/%s/transactions", c.BaseURL, c.Version)

	jsonData, err := json.Marshal(transaction)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Version", c.Version)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	return response, err
}

// Example usage of the client
func exampleClientUsage() {
	fmt.Println("\n=== Client Library Example ===")

	// Create client for latest version
	client := NewPandaPocketClient("http://localhost:8080", "v120", "your-token-here")

	// Get transactions
	transactions, err := client.GetTransactions()
	if err != nil {
		fmt.Printf("Error getting transactions: %v\n", err)
		return
	}

	fmt.Printf("Transactions: %+v\n", transactions)

	// Create a transaction
	newTransaction := map[string]interface{}{
		"category_id": 1,
		"amount":      100.0,
		"description": "Test transaction",
		"date":        "2024-01-01",
		"type":        "expense",
	}

	response, err := client.CreateTransaction(newTransaction)
	if err != nil {
		fmt.Printf("Error creating transaction: %v\n", err)
		return
	}

	fmt.Printf("Created transaction: %+v\n", response)
}
