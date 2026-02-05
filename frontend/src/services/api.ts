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
        // In dev mode with no backend, return mock data for failed requests
        if (IS_DEV_MODE && (response.status === 404 || response.status >= 500)) {
          return this.getMockResponse(response.config.url || '');
        }
        return response;
      },
      (error: AxiosError) => {
        // In dev mode, return mock data instead of failing
        if (IS_DEV_MODE && error.code === 'ERR_NETWORK') {
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
    } else if (url.includes('/providers') && !url.includes('/versions')) {
      mockData.data = { providers: [], meta: { total: 0, limit: 10, offset: 0 } };
    } else if (url.includes('/users')) {
      mockData.data = { users: [], meta: { total: 0, limit: 10, offset: 0 } };
    } else if (url.includes('/organizations')) {
      mockData.data = [];
    } else if (url.includes('/apikeys')) {
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
}

export const apiClient = new ApiClient();
export default apiClient;
