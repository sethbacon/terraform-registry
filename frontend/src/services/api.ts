import axios, { AxiosInstance, AxiosError } from 'axios';

// In dev mode, use empty baseURL to use relative paths (goes through Vite proxy)
// In production, use the configured URL or default to current origin
const IS_DEV_MODE = import.meta.env.DEV;
const API_BASE_URL = IS_DEV_MODE ? '' : (import.meta.env.VITE_API_URL || '');

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
      // Only validate successful status codes (2xx and 3xx)
      // This ensures errors are properly caught by the error interceptor
      validateStatus: (status) => status >= 200 && status < 400,
    });

    // Request interceptor to add auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('auth_token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => {
        return response;
      },
      (error: AxiosError) => {
        // In dev mode, return mock data for all errors
        if (IS_DEV_MODE) {
          return this.getMockResponse(error.config?.url || '');
        }
        
        if (error.response?.status === 401) {
          // Token expired or invalid
          localStorage.removeItem('auth_token');
          localStorage.removeItem('user');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  private getMockResponse(url: string): any {
    // Mock responses for development when backend is not available
    let mockData: any = { data: [] };

    if (url.includes('/modules') && !url.includes('/versions')) {
      mockData.data = { modules: [], meta: { total: 0, limit: 10, offset: 0 } };
    } else if (url.includes('/providers') && !url.includes('/versions') && !url.includes('/scm-providers')) {
      mockData.data = { providers: [], meta: { total: 0, limit: 10, offset: 0 } };
    } else if (url.includes('/users')) {
      mockData.data = { users: [], meta: { total: 0, limit: 10, offset: 0 } };
    } else if (url.includes('/organizations')) {
      mockData.data = [];
    } else if (url.includes('/apikeys')) {
      mockData.data = [];
    } else if (url.includes('/scm-providers')) {
      mockData.data = [];
    } else if (url.includes('/versions')) {
      mockData.data = { versions: [] };
    }

    return { data: mockData.data, status: 200 };
  }

  // Authentication
  async login(provider: 'oidc' | 'azuread') {
    window.location.href = `${API_BASE_URL}/api/v1/auth/login?provider=${provider}`;
  }

  async refreshToken() {
    const response = await this.client.post('/api/v1/auth/refresh');
    return response.data;
  }

  async getCurrentUser() {
    const response = await this.client.get('/api/v1/auth/me');
    return response.data.user;
  }

  // Modules
  async searchModules(options?: { query?: string; limit?: number; offset?: number; page?: number; per_page?: number }) {
    const params: any = {};
    
    if (options?.query) params.q = options.query;
    if (options?.limit) params.limit = options.limit;
    if (options?.offset) params.offset = options.offset;
    if (options?.page) params.page = options.page;
    if (options?.per_page) params.per_page = options.per_page;
    
    const response = await this.client.get('/api/v1/modules/search', { params });
    return response.data;
  }

  async getModuleVersions(namespace: string, name: string, system: string) {
    const response = await this.client.get(
      `/v1/modules/${namespace}/${name}/${system}/versions`
    );
    return response.data;
  }

  async uploadModule(formData: FormData) {
    const response = await this.client.post('/api/v1/modules', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  async getModule(namespace: string, name: string, system: string) {
    const response = await this.client.get(`/api/v1/modules/${namespace}/${name}/${system}`);
    return response.data;
  }

  async deleteModule(namespace: string, name: string, system: string) {
    const response = await this.client.delete(`/api/v1/modules/${namespace}/${name}/${system}`);
    return response.data;
  }

  async deleteModuleVersion(namespace: string, name: string, system: string, version: string) {
    const response = await this.client.delete(`/api/v1/modules/${namespace}/${name}/${system}/versions/${version}`);
    return response.data;
  }

  async deprecateModuleVersion(namespace: string, name: string, system: string, version: string, message?: string) {
    const response = await this.client.post(
      `/api/v1/modules/${namespace}/${name}/${system}/versions/${version}/deprecate`,
      message ? { message } : {}
    );
    return response.data;
  }

  async undeprecateModuleVersion(namespace: string, name: string, system: string, version: string) {
    const response = await this.client.delete(`/api/v1/modules/${namespace}/${name}/${system}/versions/${version}/deprecate`);
    return response.data;
  }

  // Providers
  async searchProviders(options?: { query?: string; limit?: number; offset?: number; page?: number; per_page?: number }) {
    const params: any = {};
    
    if (options?.query) params.q = options.query;
    if (options?.limit) params.limit = options.limit;
    if (options?.offset) params.offset = options.offset;
    if (options?.page) params.page = options.page;
    if (options?.per_page) params.per_page = options.per_page;
    
    const response = await this.client.get('/api/v1/providers/search', { params });
    return response.data;
  }

  async getProviderVersions(namespace: string, type: string) {
    const response = await this.client.get(
      `/v1/providers/${namespace}/${type}/versions`
    );
    return response.data;
  }

  async uploadProvider(formData: FormData) {
    const response = await this.client.post('/api/v1/providers', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  async getProvider(namespace: string, type: string) {
    const response = await this.client.get(`/api/v1/providers/${namespace}/${type}`);
    return response.data;
  }

  async deleteProvider(namespace: string, type: string) {
    const response = await this.client.delete(`/api/v1/providers/${namespace}/${type}`);
    return response.data;
  }

  async deleteProviderVersion(namespace: string, type: string, version: string) {
    const response = await this.client.delete(`/api/v1/providers/${namespace}/${type}/versions/${version}`);
    return response.data;
  }

  async deprecateProviderVersion(namespace: string, type: string, version: string, message?: string) {
    const response = await this.client.post(
      `/api/v1/providers/${namespace}/${type}/versions/${version}/deprecate`,
      message ? { message } : {}
    );
    return response.data;
  }

  async undeprecateProviderVersion(namespace: string, type: string, version: string) {
    const response = await this.client.delete(`/api/v1/providers/${namespace}/${type}/versions/${version}/deprecate`);
    return response.data;
  }

  // Helper to transform user from API format to frontend format
  private transformUser(user: any) {
    return {
      id: user.ID,
      email: user.Email,
      name: user.Name,
      oidc_sub: user.OidcSub,
      created_at: user.CreatedAt,
      updated_at: user.UpdatedAt,
    };
  }

  // Users
  async listUsers(page = 1, perPage = 20) {
    const response = await this.client.get('/api/v1/users', {
      params: { page, per_page: perPage },
    });
    const users = response.data.users || [];
    return {
      users: users.map((user: any) => this.transformUser(user)),
      pagination: response.data.pagination,
    };
  }

  async searchUsers(query: string, page = 1, perPage = 20) {
    const response = await this.client.get('/api/v1/users/search', {
      params: { q: query, page, per_page: perPage },
    });
    const users = response.data.users || [];
    return {
      users: users.map((user: any) => this.transformUser(user)),
      pagination: response.data.pagination,
    };
  }

  async getUser(id: string) {
    const response = await this.client.get(`/api/v1/users/${id}`);
    return this.transformUser(response.data.user);
  }

  async createUser(data: { email: string; name: string }) {
    const response = await this.client.post('/api/v1/users', data);
    return this.transformUser(response.data.user);
  }

  async updateUser(id: string, data: { name: string }) {
    const response = await this.client.put(`/api/v1/users/${id}`, data);
    return this.transformUser(response.data.user);
  }

  async deleteUser(id: string) {
    const response = await this.client.delete(`/api/v1/users/${id}`);
    return response.data;
  }

  // Helper to transform organization from API format to frontend format
  private transformOrganization(org: any) {
    if (!org) {
      throw new Error('Cannot transform undefined organization');
    }
    return {
      id: org.ID,
      name: org.Name,
      display_name: org.DisplayName,
      created_at: org.CreatedAt,
      updated_at: org.UpdatedAt,
    };
  }

  // Organizations
  async listOrganizations(page = 1, perPage = 20) {
    const response = await this.client.get('/api/v1/organizations', {
      params: { page, per_page: perPage },
    });
    const orgs = response.data.organizations || [];
    return orgs.map((org: any) => this.transformOrganization(org));
  }

  async searchOrganizations(query: string, page = 1, perPage = 20) {
    const response = await this.client.get('/api/v1/organizations/search', {
      params: { q: query, page, per_page: perPage },
    });
    const orgs = response.data.organizations || [];
    return orgs.map((org: any) => this.transformOrganization(org));
  }

  async getOrganization(id: string) {
    const response = await this.client.get(`/api/v1/organizations/${id}`);
    return this.transformOrganization(response.data.organization);
  }

  async createOrganization(data: { name: string; display_name: string }) {
    const response = await this.client.post('/api/v1/organizations', data);
    // Check if the response contains an error
    if (response.status !== 200 && response.status !== 201) {
      throw new Error(response.data?.error || 'Failed to create organization');
    }
    if (!response.data.organization) {
      throw new Error('Invalid response from server: missing organization data');
    }
    return this.transformOrganization(response.data.organization);
  }

  async updateOrganization(id: string, data: { display_name: string }) {
    const response = await this.client.put(`/api/v1/organizations/${id}`, data);
    return this.transformOrganization(response.data.organization);
  }

  async deleteOrganization(id: string) {
    const response = await this.client.delete(`/api/v1/organizations/${id}`);
    return response.data;
  }

  async addOrganizationMember(orgId: string, data: { user_id: string; role: string }) {
    const response = await this.client.post(
      `/api/v1/organizations/${orgId}/members`,
      data
    );
    return response.data;
  }

  async updateOrganizationMember(
    orgId: string,
    userId: string,
    data: { role: string }
  ) {
    const response = await this.client.put(
      `/api/v1/organizations/${orgId}/members/${userId}`,
      data
    );
    return response.data;
  }

  async removeOrganizationMember(orgId: string, userId: string) {
    const response = await this.client.delete(
      `/api/v1/organizations/${orgId}/members/${userId}`
    );
    return response.data;
  }

  // API Keys
  async listAPIKeys(organizationId?: string) {
    const response = await this.client.get('/api/v1/apikeys', {
      params: organizationId ? { organization_id: organizationId } : {},
    });
    return response.data;
  }

  async createAPIKey(data: {
    name: string;
    organization_id: string;
    scopes: string[];
    expires_at?: string;
  }) {
    const response = await this.client.post('/api/v1/apikeys', data);
    return response.data;
  }

  async getAPIKey(id: string) {
    const response = await this.client.get(`/api/v1/apikeys/${id}`);
    return response.data;
  }

  async updateAPIKey(
    id: string,
    data: { name?: string; scopes?: string[]; expires_at?: string }
  ) {
    const response = await this.client.put(`/api/v1/apikeys/${id}`, data);
    return response.data;
  }

  async deleteAPIKey(id: string) {
    const response = await this.client.delete(`/api/v1/apikeys/${id}`);
    return response.data;
  }

  // SCM Provider Management
  async listSCMProviders(organizationId?: string) {
    const params = organizationId ? { organization_id: organizationId } : {};
    const response = await this.client.get('/api/v1/scm-providers', { params });
    return response.data;
  }

  async createSCMProvider(data: {
    organization_id: string;
    provider_type: string;
    name: string;
    base_url?: string | null;
    client_id: string;
    client_secret: string;
    webhook_secret?: string;
  }) {
    const response = await this.client.post('/api/v1/scm-providers', data);
    return response.data;
  }

  async getSCMProvider(id: string) {
    const response = await this.client.get(`/api/v1/scm-providers/${id}`);
    return response.data;
  }

  async updateSCMProvider(
    id: string,
    data: {
      name?: string;
      base_url?: string | null;
      client_id?: string;
      client_secret?: string;
      webhook_secret?: string;
      is_active?: boolean;
    }
  ) {
    const response = await this.client.put(`/api/v1/scm-providers/${id}`, data);
    return response.data;
  }

  async deleteSCMProvider(id: string) {
    const response = await this.client.delete(`/api/v1/scm-providers/${id}`);
    return response.data;
  }

  // SCM OAuth
  async initiateSCMOAuth(providerId: string) {
    const response = await this.client.get(`/api/v1/scm-providers/${providerId}/oauth/authorize`);
    return response.data;
  }

  async refreshSCMToken(providerId: string) {
    const response = await this.client.post(`/api/v1/scm-providers/${providerId}/oauth/refresh`);
    return response.data;
  }

  async revokeSCMToken(providerId: string) {
    const response = await this.client.delete(`/api/v1/scm-providers/${providerId}/oauth/token`);
    return response.data;
  }

  // Module SCM Linking
  async linkModuleToSCM(
    moduleId: string,
    data: {
      provider_id: string;
      repository_owner: string;
      repository_name: string;
      repository_path?: string;
      default_branch?: string;
      auto_publish_enabled?: boolean;
      tag_pattern?: string;
    }
  ) {
    const response = await this.client.post(`/api/v1/admin/modules/${moduleId}/scm`, data);
    return response.data;
  }

  async getModuleSCMInfo(moduleId: string) {
    const response = await this.client.get(`/api/v1/admin/modules/${moduleId}/scm`);
    return response.data;
  }

  async updateModuleSCMLink(
    moduleId: string,
    data: {
      repository_path?: string;
      default_branch?: string;
      auto_publish_enabled?: boolean;
      tag_pattern?: string;
    }
  ) {
    const response = await this.client.put(`/api/v1/admin/modules/${moduleId}/scm`, data);
    return response.data;
  }

  async unlinkModuleFromSCM(moduleId: string) {
    const response = await this.client.delete(`/api/v1/admin/modules/${moduleId}/scm`);
    return response.data;
  }

  async triggerManualSync(moduleId: string, data?: { tag_name?: string; commit_sha?: string }) {
    const response = await this.client.post(`/api/v1/admin/modules/${moduleId}/scm/sync`, data || {});
    return response.data;
  }

  async getWebhookEvents(moduleId: string) {
    const response = await this.client.get(`/api/v1/admin/modules/${moduleId}/scm/events`);
    return response.data;
  }

  // Dashboard Stats
  async getDashboardStats() {
    const response = await this.client.get('/api/v1/admin/stats/dashboard');
    return response.data;
  }

  // Mirror Management
  async listMirrors(enabledOnly = false) {
    const params = enabledOnly ? { enabled: 'true' } : {};
    const response = await this.client.get('/api/v1/admin/mirrors', { params });
    return response.data.mirrors || [];
  }

  async getMirror(id: string) {
    const response = await this.client.get(`/api/v1/admin/mirrors/${id}`);
    return response.data;
  }

  async createMirror(data: {
    name: string;
    description?: string;
    upstream_registry_url: string;
    organization_id?: string;
    namespace_filter?: string[];
    provider_filter?: string[];
    enabled?: boolean;
    sync_interval_hours?: number;
  }) {
    const response = await this.client.post('/api/v1/admin/mirrors', data);
    return response.data;
  }

  async updateMirror(
    id: string,
    data: {
      name?: string;
      description?: string;
      upstream_registry_url?: string;
      organization_id?: string;
      namespace_filter?: string[];
      provider_filter?: string[];
      enabled?: boolean;
      sync_interval_hours?: number;
    }
  ) {
    const response = await this.client.put(`/api/v1/admin/mirrors/${id}`, data);
    return response.data;
  }

  async deleteMirror(id: string) {
    const response = await this.client.delete(`/api/v1/admin/mirrors/${id}`);
    return response.data;
  }

  async triggerMirrorSync(id: string, data?: { namespace?: string; provider_name?: string }) {
    const response = await this.client.post(`/api/v1/admin/mirrors/${id}/sync`, data || {});
    return response.data;
  }

  async getMirrorStatus(id: string) {
    const response = await this.client.get(`/api/v1/admin/mirrors/${id}/status`);
    return response.data;
  }
}

export const apiClient = new ApiClient();
export default apiClient;

