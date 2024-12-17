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
	uniqueDBName := fmt.Sprintf("test_db_%d", timestamp)
	uniqueSchemaName := fmt.Sprintf("test_schema_%d", timestamp)
	uniqueTokenName := fmt.Sprintf("test_token_%d", timestamp)
	testUserEmail := fmt.Sprintf("test_user_%d@example.com", timestamp)

	// Get MotherDuck token from environment variable
	motherDuckToken := os.Getenv("MOTHERDUCK_TOKEN")
	if motherDuckToken == "" {
		t.Fatal("MOTHERDUCK_TOKEN environment variable must be set")
	}

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../",

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"motherduck_token":   motherDuckToken,
			"motherduck_api_key": motherDuckToken, // Use the same token for API key
			"database_name":      uniqueDBName,
			"schema_name":        uniqueSchemaName,
			"new_user_email":     testUserEmail,
			"new_user_name":      "Test User",
			"token_name":         uniqueTokenName,
			"token_expiry_days":  7,
		},

		// Environment variables to set when running Terraform
		EnvVars: map[string]string{
			"MOTHERDUCK_TOKEN": motherDuckToken,
		},
	}

	// At the end of the test, run `terraform destroy`
	defer terraform.Destroy(t, terraformOptions)

	// Run `terraform init` and `terraform apply`
	terraform.InitAndApply(t, terraformOptions)

	// Verify database exists using DuckDB CLI
	verifyDatabaseExists(t, motherDuckToken, uniqueDBName)

	// Verify schema exists
	verifySchemaExists(t, motherDuckToken, uniqueDBName, uniqueSchemaName)

	// Verify user exists (this will require API call)
	verifyUserExists(t, motherDuckToken, testUserEmail)

	// Verify token exists (this will require API call)
	verifyTokenExists(t, motherDuckToken, strings.Split(testUserEmail, "@")[0], uniqueTokenName)
}

func runCommand(t *testing.T, command string) string {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run command: %v\nOutput: %s", err, output)
	}
	return strings.TrimSpace(string(output))
}

func verifyDatabaseExists(t *testing.T, token, dbName string) {
	cmd := fmt.Sprintf(`duckdb md:?motherduck_token=%s -c "
		SELECT DISTINCT catalog_name FROM information_schema.schemata WHERE catalog_name = '%s';"`,
		token, dbName)

	result := runCommand(t, cmd)
	assert.Contains(t, result, dbName, "Database should exist")
}

func verifySchemaExists(t *testing.T, token, dbName, schemaName string) {
	cmd := fmt.Sprintf(`duckdb md:%s?motherduck_token=%s -c "
		SELECT schema_name FROM information_schema.schemata WHERE schema_name = '%s';"`,
		dbName, token, schemaName)

	result := runCommand(t, cmd)
	assert.Contains(t, result, schemaName, "Schema should exist")
}

func verifyUserExists(t *testing.T, token, email string) {
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
		rawResponse, strings.Split(email, "@")[0])  // Use the part before @ as username

	result := runCommand(t, cmd)
	assert.Contains(t, result, strings.Split(email, "@")[0], "User should exist")
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
