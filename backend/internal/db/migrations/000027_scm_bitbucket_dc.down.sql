-- Remove bitbucket_dc providers and their tokens
DELETE FROM scm_oauth_tokens WHERE scm_provider_id IN (SELECT id FROM scm_providers WHERE provider_type = 'bitbucket_dc');
DELETE FROM scm_providers WHERE provider_type = 'bitbucket_dc';

-- Restore original CHECK constraint
ALTER TABLE scm_providers DROP CONSTRAINT IF EXISTS scm_providers_provider_type_check;
ALTER TABLE scm_providers ADD CONSTRAINT scm_providers_provider_type_check
  CHECK (provider_type IN ('github', 'azuredevops', 'gitlab'));
