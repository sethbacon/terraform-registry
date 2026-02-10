# Test Terraform configuration using the local registry
# Make sure the backend is running on port 443 (HTTPS)

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "registry.local/hashicorp/aws"
      version = "6.31.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"

  # Skip credential validation for testing module download
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

# Use the VPC module from our local registry
module "vpc" {
  source  = "registry.local/bconline/vpc/aws"
  version = "1.1.0"

  # Module variables
  cidr_block  = "10.0.0.0/16"
  name        = "test-vpc"
  enable_ipv6 = false

  tags = {
    Environment = "test"
    ManagedBy   = "terraform"
  }
}

output "vpc_id" {
  description = "The ID of the VPC"
  value       = module.vpc.vpc_id
}
