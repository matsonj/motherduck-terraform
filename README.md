# MotherDuck Terraform Configuration

This repository contains Terraform configurations for managing MotherDuck resources.

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
