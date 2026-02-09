variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "name" {
  description = "Name prefix for all resources"
  type        = string
  default     = "terraform-registry"
}

variable "environment" {
  description = "Environment name (e.g., production, staging)"
  type        = string
  default     = "production"
}

variable "domain" {
  description = "Custom domain name. Leave empty to use Cloud Run default URL."
  type        = string
  default     = ""
}

variable "image_tag" {
  description = "Container image tag"
  type        = string
  default     = "latest"
}

variable "db_tier" {
  description = "Cloud SQL instance tier"
  type        = string
  default     = "db-f1-micro"
}

variable "backend_min_instances" {
  description = "Backend minimum instances"
  type        = number
  default     = 1
}

variable "backend_max_instances" {
  description = "Backend maximum instances"
  type        = number
  default     = 10
}

variable "frontend_min_instances" {
  description = "Frontend minimum instances"
  type        = number
  default     = 1
}

variable "frontend_max_instances" {
  description = "Frontend maximum instances"
  type        = number
  default     = 5
}

variable "database_password" {
  description = "PostgreSQL password"
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "JWT signing secret (min 32 chars)"
  type        = string
  sensitive   = true
}

variable "encryption_key" {
  description = "AES-256 encryption key (32 bytes)"
  type        = string
  sensitive   = true
}
