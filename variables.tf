# Variables for MotherDuck Terraform configuration

variable "motherduck_api_key" {
  description = "MotherDuck API Key"
  type        = string
  sensitive   = true
}

variable "motherduck_token" {
  description = "MotherDuck API token"
  type        = string
  sensitive   = true
}

variable "database_name" {
  description = "Name of the database to create"
  type        = string
}

variable "schema_name" {
  description = "Name of the schema to create"
  type        = string
  default     = "main"
}

variable "new_user_name" {
  description = "Name for the new user"
  type        = string
}

variable "token_name" {
  description = "Name for the new token"
  type        = string
}

variable "token_expiry_days" {
  description = "Number of days until token expires"
  type        = number
  default     = 30
}
