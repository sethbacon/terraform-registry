-- Add tenant_id column to scm_providers for Microsoft Entra ID OAuth support
ALTER TABLE scm_providers ADD COLUMN tenant_id VARCHAR(255);
