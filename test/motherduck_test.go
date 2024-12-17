package test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestMotherDuckTerraform(t *testing.T) {
	t.Parallel()

	// Generate unique names for testing
	timestamp := time.Now().Unix()
	dbName := fmt.Sprintf("test_db_%d", timestamp)
	schemaName := fmt.Sprintf("test_schema_%d", timestamp)
	userName := "Test User"
	tokenName := fmt.Sprintf("test_token_%d", timestamp)
	shareUrl := "md:_share/sample_data/23b0d623-1361-421d-ae77-62d701d471e6"
	shareName := fmt.Sprintf("test_sample_data_%d", timestamp)

	// Get MotherDuck token from environment variable
	motherDuckToken := os.Getenv("MOTHERDUCK_TOKEN")
	if motherDuckToken == "" {
		t.Fatal("MOTHERDUCK_TOKEN environment variable must be set")
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../",
		Vars: map[string]interface{}{
			"motherduck_token":   motherDuckToken,
			"motherduck_api_key": motherDuckToken,
			"database_name":      dbName,
			"schema_name":        schemaName,
			"new_user_name":      userName,
			"token_name":         tokenName,
			"token_expiry_days":  7,
			"share_url":          shareUrl,
			"share_name":         shareName,
		},
	})

	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	token := motherDuckToken
	verifyDatabaseExists(t, token, dbName)
	verifySchemaExists(t, token, dbName, schemaName)
	verifyUserExists(t, token, userName)
	verifyTokenExists(t, token, userName, tokenName)
	verifyShareAttached(t, token, shareName)
}

func runCommand(t *testing.T, command string) string {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command output: %s", string(output))
		t.Logf("Command error: %v", err)
	}
	return string(output)
}

func verifyDatabaseExists(t *testing.T, token, dbName string) {
	cmd := fmt.Sprintf(`duckdb md:?motherduck_token=%s -c "
		SELECT catalog_name FROM information_schema.schemata WHERE catalog_name = '%s';"`, token, dbName)

	result := runCommand(t, cmd)
	assert.Contains(t, result, dbName, "Database should exist")
}

func verifySchemaExists(t *testing.T, token, dbName, schemaName string) {
	cmd := fmt.Sprintf(`duckdb md:?motherduck_token=%s -c "
		USE %s;
		SELECT schema_name FROM information_schema.schemata WHERE schema_name = '%s';"`, token, dbName, schemaName)

	result := runCommand(t, cmd)
	assert.Contains(t, result, schemaName, "Schema should exist")
}

func verifyUserExists(t *testing.T, token, username string) {
	// First get the raw response
	rawCmd := fmt.Sprintf(`curl -sL "https://api.motherduck.com/v1/users" \
		-H "Content-Type: application/json" \
		-H "Accept: application/json" \
		-H "Authorization: Bearer %s"`, token)
	rawResponse := runCommand(t, rawCmd)
	t.Logf("Raw API Response: %s", rawResponse)

	// Check for invalid token
	if strings.Contains(rawResponse, "Invalid MotherDuck token") {
		t.Fatal("Invalid MotherDuck token")
		return
	}

	// Check for permission error
	if strings.Contains(rawResponse, "UNAUTHORIZED") || strings.Contains(rawResponse, "Not Found") {
		t.Skip("Skipping user verification due to insufficient permissions")
		return
	}

	// Now try to parse it
	cmd := fmt.Sprintf(`echo '%s' | jq -r '.[] | select(.username=="%s") | .username'`,
		rawResponse, username)

	result := runCommand(t, cmd)
	assert.Contains(t, result, username, "User should exist")
}

func verifyTokenExists(t *testing.T, token, username, tokenName string) {
	// Get the user's tokens
	rawCmd := fmt.Sprintf(`curl -sL "https://api.motherduck.com/v1/users/%s/tokens" \
		-H "Content-Type: application/json" \
		-H "Accept: application/json" \
		-H "Authorization: Bearer %s"`, username, token)
	rawResponse := runCommand(t, rawCmd)
	t.Logf("Raw Token API Response: %s", rawResponse)

	// Check for invalid token
	if strings.Contains(rawResponse, "Invalid MotherDuck token") {
		t.Fatal("Invalid MotherDuck token")
		return
	}

	// Check for permission error
	if strings.Contains(rawResponse, "UNAUTHORIZED") || strings.Contains(rawResponse, "Not Found") {
		t.Skip("Skipping token verification due to insufficient permissions")
		return
	}

	// Check if our token exists
	cmd := fmt.Sprintf(`echo '%s' | jq -r '.[] | select(.name=="%s") | .name'`,
		rawResponse, tokenName)

	result := runCommand(t, cmd)
	assert.Contains(t, result, tokenName, "Token should exist")
}

func verifyShareAttached(t *testing.T, token, shareName string) {
	cmd := fmt.Sprintf(`duckdb md:?motherduck_token=%s -c "
		SELECT catalog_name FROM information_schema.schemata WHERE catalog_name = '%s';"`, token, shareName)

	result := runCommand(t, cmd)
	assert.Contains(t, result, shareName, "Share should be attached")
}
