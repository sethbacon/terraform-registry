-- Add bitbucket_dc to allowed SCM provider types
ALTER TABLE scm_providers DROP CONSTRAINT IF EXISTS scm_providers_provider_type_check;
ALTER TABLE scm_providers ADD CONSTRAINT scm_providers_provider_type_check
  CHECK (provider_type IN ('github', 'azuredevops', 'gitlab', 'bitbucket_dc'));
