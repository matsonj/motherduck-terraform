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

		// Disable colors in Terraform commands
		NoColor: true,
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
	cmd := fmt.Sprintf(`duckdb -c "
		ATTACH 'md:?motherduck_token=%s';
		SELECT name FROM md:information_schema.databases WHERE name = '%s';"`,
		token, dbName)
	
	result := runCommand(t, cmd)
	assert.Contains(t, result, dbName, "Database should exist")
}

func verifySchemaExists(t *testing.T, token, dbName, schemaName string) {
	cmd := fmt.Sprintf(`duckdb -c "
		ATTACH 'md:?motherduck_token=%s';
		SELECT schema_name FROM md:%s.information_schema.schemata WHERE schema_name = '%s';"`,
		token, dbName, schemaName)
	
	result := runCommand(t, cmd)
	assert.Contains(t, result, schemaName, "Schema should exist")
}

func verifyUserExists(t *testing.T, token, email string) {
	cmd := fmt.Sprintf(`curl -s -X GET "https://api.motherduck.com/api/v0/organizations/self/users" \
		-H "Authorization: Bearer %s" | jq -r '.users[] | select(.email=="%s") | .email'`,
		token, email)
	
	result := runCommand(t, cmd)
	assert.Contains(t, result, email, "User should exist")
}
