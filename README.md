# MotherDuck Terraform Configuration

This repository contains Terraform configurations for managing MotherDuck resources.

> [!NOTE]  
> This repo serves as a demonstration of the capabilities of the MotherDuck REST API. Feel to use it as inspiration, but Production use is not recommended.

## Prerequisites

- Terraform >= 1.0.0
- MotherDuck API Key

## Setup

1. Set your MotherDuck API key as an environment variable:
   ```bash
   export MOTHERDUCK_API_KEY="your-api-key"
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Plan your changes:
   ```bash
   terraform plan
   ```

4. Apply the configuration:
   ```bash
   terraform apply
   ```

## Structure

- `main.tf` - Main Terraform configuration and provider settings
- `variables.tf` - Variable definitions
- `.gitignore` - Git ignore patterns for Terraform files

## Security

- Never commit sensitive information like API keys to version control
- Use environment variables for sensitive values
- The `sensitive = true` flag is set for sensitive variables

## Testing

The repository includes integration tests written in Go. The tests verify the creation and cleanup of MotherDuck resources including databases, schemas, users, and tokens.

### Running Tests

1. Ensure you have Go installed and your MotherDuck token is set:
   ```bash
   export MOTHERDUCK_TOKEN="your-token"
   ```

2. Run the tests:
   ```bash
   cd test
   go test -v ./...
   ```

### Test Behavior

- Tests will skip user and token verification if you don't have admin permissions
- Tests will fail if the MotherDuck token is invalid
- Each test run creates unique resources with timestamps to avoid conflicts
- All resources are automatically cleaned up after tests complete
