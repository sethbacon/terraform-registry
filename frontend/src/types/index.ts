export interface User {
  id: string;
  email: string;
  name: string;
  username?: string; // Alias for name for UI purposes
  role?: string; // User role (admin, user, etc.)
  organization_name?: string; // Associated organization
  oidc_sub?: string;
  created_at: string;
  updated_at: string;
}

export interface Organization {
  id: string;
  name: string;
  display_name: string;
  created_at: string;
  updated_at: string;
}

export interface OrganizationMember {
  organization_id: string;
  user_id: string;
  role: string;
  created_at: string;
}

export interface APIKey {
  id: string;
  user_id?: string;
  organization_id: string;
  name: string;
  description?: string;
  key_prefix: string;
  scopes: string[];
  expires_at?: string;
  last_used_at?: string;
  created_at: string;
}

export interface Module {
  id: string;
  namespace: string;
  name: string;
  system: string;
  provider?: string; // Alias for system for backward compatibility
  description?: string;
  source?: string;
  organization_id: string;
  latest_version?: string; // Latest version string
  download_count?: number; // Total downloads
  created_at: string;
  updated_at: string;
}

export interface ModuleVersion {
  id: string;
  module_id: string;
  version: string;
  storage_path: string;
  storage_backend: string;
  size_bytes: number;
  checksum: string;
  readme?: string;
  download_count: number;
  created_at: string;
}

export interface Provider {
  id: string;
  namespace: string;
  type: string;
  description?: string;
  source?: string;
  organization_id: string;
  organization_name?: string;
  latest_version?: string; // Latest version string
  download_count?: number; // Total downloads
  created_at: string;
  updated_at: string;
}

export interface ProviderPlatform {
  id: string;
  provider_version_id: string;
  os: string;
  arch: string;
  filename: string;
  storage_path: string;
  storage_backend: string;
  size_bytes: number;
  shasum: string;
  download_count: number;
}

export interface ProviderVersion {
  id: string;
  provider_id: string;
  version: string;
  protocols: string[];
  gpg_public_key: string;
  shasums_url: string;
  shasums_signature_url: string;
  published_at: string;
  download_count?: number;
  platforms?: ProviderPlatform[];
  created_at: string;
}

export interface ProviderPlatform {
  id: string;
  provider_version_id: string;
  os: string;
  arch: string;
  filename: string;
  size_bytes: number;
  shasum: string;
  download_count: number;
}

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
}

export interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (userOrProvider: User | 'oidc' | 'azuread') => void;
  logout: () => void;
  refreshToken: () => Promise<void>;
}
