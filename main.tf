# Main Terraform configuration file for MotherDuck resources

terraform {
  required_version = ">= 1.0.0"
}

# TODO: Add MotherDuck provider configuration once officially available
# For now, you can use local-exec provisioners or other providers as needed

# Resource to create a database
resource "null_resource" "database" {
  triggers = {
    database_name    = var.database_name
    motherduck_token = var.motherduck_token
  }

  provisioner "local-exec" {
    command = <<-EOT
      duckdb -c "
        ATTACH 'md:?motherduck_token=${var.motherduck_token}';
        CREATE DATABASE IF NOT EXISTS ${var.database_name};
      "
    EOT
  }

  # Destroy-time provisioner to clean up the database
  provisioner "local-exec" {
    when    = destroy
    command = <<-EOT
      duckdb -c "
        ATTACH 'md:?motherduck_token=${self.triggers.motherduck_token}';
        DROP DATABASE IF EXISTS ${self.triggers.database_name} CASCADE;
      "
    EOT
  }
}

# Resource to create a schema
resource "null_resource" "schema" {
  triggers = {
    database_name    = var.database_name
    schema_name      = var.schema_name
    motherduck_token = var.motherduck_token
  }

  depends_on = [null_resource.database]

  provisioner "local-exec" {
    command = <<-EOT
      duckdb -c "
        ATTACH 'md:?motherduck_token=${var.motherduck_token}';
        CREATE SCHEMA IF NOT EXISTS ${var.database_name}.${var.schema_name};
      "
    EOT
  }

  # Destroy-time provisioner to clean up the schema
  provisioner "local-exec" {
    when    = destroy
    command = <<-EOT
      duckdb -c "
        ATTACH 'md:?motherduck_token=${self.triggers.motherduck_token}';
        DROP SCHEMA IF EXISTS ${self.triggers.database_name}.${self.triggers.schema_name} CASCADE;
      "
    EOT
  }
}

# Resource to create a user
resource "null_resource" "user" {
  triggers = {
    email            = var.new_user_email
    name             = var.new_user_name
    motherduck_token = var.motherduck_token
  }

  provisioner "local-exec" {
    command = <<-EOT
      curl -X POST "https://api.motherduck.com/api/v0/organizations/self/users" \
        -H "Authorization: Bearer ${var.motherduck_token}" \
        -H "Content-Type: application/json" \
        -d '{
          "email": "${var.new_user_email}",
          "name": "${var.new_user_name}"
        }' > user_response.json

      # Store the user ID for token creation
      echo $(cat user_response.json | jq -r '.id') > user_id.txt
    EOT
  }

  # Destroy-time provisioner to remove the user
  provisioner "local-exec" {
    when    = destroy
    command = <<-EOT
      USER_ID=$(cat user_id.txt)
      curl -X DELETE "https://api.motherduck.com/api/v0/organizations/self/users/$USER_ID" \
        -H "Authorization: Bearer ${self.triggers.motherduck_token}"
    EOT
  }
}

# Resource to create a token
resource "null_resource" "token" {
  triggers = {
    motherduck_token = var.motherduck_token
    token_name       = var.token_name
    expiry_days      = var.token_expiry_days
  }

  depends_on = [null_resource.user]

  provisioner "local-exec" {
    command = <<-EOT
      USER_ID=$(cat user_id.txt)
      EXPIRY_DATE=$(date -v +${var.token_expiry_days}d -u +"%Y-%m-%dT%H:%M:%SZ")
      
      curl -X POST "https://api.motherduck.com/api/v0/organizations/self/users/$USER_ID/tokens" \
        -H "Authorization: Bearer ${var.motherduck_token}" \
        -H "Content-Type: application/json" \
        -d '{
          "name": "${var.token_name}",
          "expiresAt": "'$EXPIRY_DATE'"
        }' > token_response.json

      # Store the token ID for cleanup
      echo $(cat token_response.json | jq -r '.id') > token_id.txt
    EOT
  }

  # Destroy-time provisioner to revoke the token
  provisioner "local-exec" {
    when    = destroy
    command = <<-EOT
      USER_ID=$(cat user_id.txt)
      TOKEN_ID=$(cat token_id.txt)
      curl -X DELETE "https://api.motherduck.com/api/v0/organizations/self/users/$USER_ID/tokens/$TOKEN_ID" \
        -H "Authorization: Bearer ${self.triggers.motherduck_token}"
    EOT
  }
}
