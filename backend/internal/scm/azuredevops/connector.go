package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/terraform-registry/terraform-registry/internal/scm"
)

const (
	defaultAzureDevOpsURL = "https://dev.azure.com"
	azureAuthURL          = "https://app.vssps.visualstudio.com/oauth2/authorize"
	azureTokenURL         = "https://app.vssps.visualstudio.com/oauth2/token"
)

// AzureDevOpsConnector implements scm.Connector for Azure DevOps
type AzureDevOpsConnector struct {
	clientID     string
	clientSecret string
	callbackURL  string
	baseURL      string
	organization string
}

// NewAzureDevOpsConnector creates an Azure DevOps connector
func NewAzureDevOpsConnector(settings *scm.ConnectorSettings) (*AzureDevOpsConnector, error) {
	baseURL := defaultAzureDevOpsURL
	if settings.InstanceBaseURL != "" {
		baseURL = settings.InstanceBaseURL
	}

	return &AzureDevOpsConnector{
		clientID:     settings.ClientID,
		clientSecret: settings.ClientSecret,
		callbackURL:  settings.CallbackURL,
		baseURL:      baseURL,
	}, nil
}

func (c *AzureDevOpsConnector) Platform() scm.ProviderKind {
	return scm.KindAzureDevOps
}

func (c *AzureDevOpsConnector) AuthorizationEndpoint(stateParam string, requestedScopes []string) string {
	scopes := "vso.code vso.project"
	if len(requestedScopes) > 0 {
		scopes = strings.Join(requestedScopes, " ")
	}

	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("response_type", "Assertion")
	params.Set("state", stateParam)
	params.Set("scope", scopes)
	params.Set("redirect_uri", c.callbackURL)

	return fmt.Sprintf("%s?%s", azureAuthURL, params.Encode())
}

func (c *AzureDevOpsConnector) CompleteAuthorization(ctx context.Context, authCode string) (*scm.AccessToken, error) {
	data := url.Values{}
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", c.clientSecret)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("assertion", authCode)
	data.Set("redirect_uri", c.callbackURL)

	req, err := http.NewRequestWithContext(ctx, "POST", azureTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, scm.WrapRemoteError(0, "failed to exchange code", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, scm.WrapRemoteError(resp.StatusCode, "oauth code exchange failed", fmt.Errorf("%s", body))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	scopes := []string{}
	if result.Scope != "" {
		scopes = strings.Split(result.Scope, " ")
	}

	return &scm.AccessToken{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresAt:    &expiresAt,
		Scopes:       scopes,
	}, nil
}

func (c *AzureDevOpsConnector) RenewToken(ctx context.Context, refreshToken string) (*scm.AccessToken, error) {
	data := url.Values{}
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", c.clientSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("assertion", refreshToken)
	data.Set("redirect_uri", c.callbackURL)

	req, err := http.NewRequestWithContext(ctx, "POST", azureTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, scm.WrapRemoteError(0, "failed to refresh token", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, scm.ErrTokenRefreshFailed
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return &scm.AccessToken{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresAt:    &expiresAt,
	}, nil
}

func (c *AzureDevOpsConnector) FetchRepositories(ctx context.Context, creds *scm.AccessToken, pagination scm.Pagination) (*scm.RepoListResult, error) {
	// First, get projects
	projects, err := c.fetchProjects(ctx, creds)
	if err != nil {
		return nil, err
	}

	allRepos := []*scm.SourceRepo{}

	// Fetch repos for each project
	for _, project := range projects {
		endpoint := fmt.Sprintf("%s/%s/%s/_apis/git/repositories?api-version=7.0", c.baseURL, c.organization, project.Name)

		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			continue
		}
		c.setAuthHeaders(req, creds)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}

		var result struct {
			Value []adoRepo `json:"value"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		for _, adoRepo := range result.Value {
			allRepos = append(allRepos, c.convertRepo(&adoRepo, project.Name))
		}
	}

	return &scm.RepoListResult{
		Repos:     allRepos,
		MorePages: false,
	}, nil
}

func (c *AzureDevOpsConnector) FetchRepository(ctx context.Context, creds *scm.AccessToken, ownerName, repoName string) (*scm.SourceRepo, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s?api-version=7.0", c.baseURL, c.organization, ownerName, repoName)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeaders(req, creds)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, scm.WrapRemoteError(0, "failed to fetch repository", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, scm.ErrRepoNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, scm.WrapRemoteError(resp.StatusCode, "failed to fetch repository", nil)
	}

	var adoRepo adoRepo
	if err := json.NewDecoder(resp.Body).Decode(&adoRepo); err != nil {
		return nil, err
	}

	return c.convertRepo(&adoRepo, ownerName), nil
}

func (c *AzureDevOpsConnector) SearchRepositories(ctx context.Context, creds *scm.AccessToken, searchTerm string, pagination scm.Pagination) (*scm.RepoListResult, error) {
	// Azure DevOps doesn't have direct repo search, so fetch all and filter
	allRepos, err := c.FetchRepositories(ctx, creds, pagination)
	if err != nil {
		return nil, err
	}

	filtered := []*scm.SourceRepo{}
	searchLower := strings.ToLower(searchTerm)
	for _, repo := range allRepos.Repos {
		if strings.Contains(strings.ToLower(repo.RepoName), searchLower) ||
			strings.Contains(strings.ToLower(repo.Description), searchLower) {
			filtered = append(filtered, repo)
		}
	}

	return &scm.RepoListResult{
		Repos:     filtered,
		MorePages: false,
	}, nil
}

func (c *AzureDevOpsConnector) FetchBranches(ctx context.Context, creds *scm.AccessToken, ownerName, repoName string, pagination scm.Pagination) ([]*scm.GitBranch, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/refs?filter=heads/&api-version=7.0", c.baseURL, c.organization, ownerName, repoName)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeaders(req, creds)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, scm.WrapRemoteError(0, "failed to fetch branches", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, scm.WrapRemoteError(resp.StatusCode, "failed to fetch branches", nil)
	}

	var result struct {
		Value []struct {
			Name     string `json:"name"`
			ObjectID string `json:"objectId"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	branches := make([]*scm.GitBranch, len(result.Value))
	for i, ref := range result.Value {
		branchName := strings.TrimPrefix(ref.Name, "refs/heads/")
		branches[i] = &scm.GitBranch{
			BranchName: branchName,
			HeadCommit: ref.ObjectID,
		}
	}

	return branches, nil
}

func (c *AzureDevOpsConnector) FetchTags(ctx context.Context, creds *scm.AccessToken, ownerName, repoName string, pagination scm.Pagination) ([]*scm.GitTag, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/refs?filter=tags/&api-version=7.0", c.baseURL, c.organization, ownerName, repoName)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeaders(req, creds)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, scm.WrapRemoteError(0, "failed to fetch tags", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, scm.WrapRemoteError(resp.StatusCode, "failed to fetch tags", nil)
	}

	var result struct {
		Value []struct {
			Name     string `json:"name"`
			ObjectID string `json:"objectId"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	tags := make([]*scm.GitTag, len(result.Value))
	for i, ref := range result.Value {
		tagName := strings.TrimPrefix(ref.Name, "refs/tags/")
		tags[i] = &scm.GitTag{
			TagName:      tagName,
			TargetCommit: ref.ObjectID,
		}
	}

	return tags, nil
}

func (c *AzureDevOpsConnector) FetchTagByName(ctx context.Context, creds *scm.AccessToken, ownerName, repoName, tagName string) (*scm.GitTag, error) {
	tags, err := c.FetchTags(ctx, creds, ownerName, repoName, scm.DefaultPagination())
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		if tag.TagName == tagName {
			return tag, nil
		}
	}

	return nil, scm.ErrTagNotFound
}

func (c *AzureDevOpsConnector) FetchCommit(ctx context.Context, creds *scm.AccessToken, ownerName, repoName, commitHash string) (*scm.GitCommit, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/commits/%s?api-version=7.0", c.baseURL, c.organization, ownerName, repoName, commitHash)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeaders(req, creds)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, scm.WrapRemoteError(0, "failed to fetch commit", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, scm.ErrCommitNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, scm.WrapRemoteError(resp.StatusCode, "failed to fetch commit", nil)
	}

	var adoCommit struct {
		CommitID string `json:"commitId"`
		Comment  string `json:"comment"`
		Author   struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
		RemoteURL string `json:"remoteUrl"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&adoCommit); err != nil {
		return nil, err
	}

	return &scm.GitCommit{
		CommitHash:  adoCommit.CommitID,
		Subject:     adoCommit.Comment,
		AuthorName:  adoCommit.Author.Name,
		AuthorEmail: adoCommit.Author.Email,
		CommittedAt: adoCommit.Author.Date,
		CommitURL:   adoCommit.RemoteURL,
	}, nil
}

func (c *AzureDevOpsConnector) DownloadSourceArchive(ctx context.Context, creds *scm.AccessToken, ownerName, repoName, gitRef string, format scm.ArchiveKind) (io.ReadCloser, error) {
	// Azure DevOps archive download
	endpoint := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/items?path=/&versionDescriptor.version=%s&$format=zip&api-version=7.0",
		c.baseURL, c.organization, ownerName, repoName, gitRef)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeaders(req, creds)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, scm.WrapRemoteError(0, "failed to download archive", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, scm.WrapRemoteError(resp.StatusCode, "failed to download archive", nil)
	}

	return resp.Body, nil
}

// Stub methods for webhooks
func (c *AzureDevOpsConnector) RegisterWebhook(ctx context.Context, creds *scm.AccessToken, ownerName, repoName string, hookConfig scm.WebhookSetup) (*scm.WebhookInfo, error) {
	return nil, scm.ErrWebhookSetupFailed
}

func (c *AzureDevOpsConnector) RemoveWebhook(ctx context.Context, creds *scm.AccessToken, ownerName, repoName, hookID string) error {
	return scm.ErrWebhookNotFound
}

func (c *AzureDevOpsConnector) ParseDelivery(payloadBytes []byte, httpHeaders map[string]string) (*scm.IncomingHook, error) {
	return nil, scm.ErrWebhookPayloadMalformed
}

func (c *AzureDevOpsConnector) VerifyDeliverySignature(payloadBytes []byte, signatureHeader, sharedSecret string) bool {
	return false
}

// Helper methods

func (c *AzureDevOpsConnector) setAuthHeaders(req *http.Request, creds *scm.AccessToken) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", creds.AccessToken))
	req.Header.Set("Content-Type", "application/json")
}

func (c *AzureDevOpsConnector) fetchProjects(ctx context.Context, creds *scm.AccessToken) ([]adoProject, error) {
	endpoint := fmt.Sprintf("%s/%s/_apis/projects?api-version=7.0", c.baseURL, c.organization)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeaders(req, creds)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Value []adoProject `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Value, nil
}

func (c *AzureDevOpsConnector) convertRepo(adoRepo *adoRepo, projectName string) *scm.SourceRepo {
	return &scm.SourceRepo{
		Owner:         projectName,
		OwnerName:     projectName,
		Name:          adoRepo.Name,
		RepoName:      adoRepo.Name,
		FullName:      fmt.Sprintf("%s/%s", projectName, adoRepo.Name),
		FullPath:      fmt.Sprintf("%s/%s", projectName, adoRepo.Name),
		HTMLURL:       adoRepo.WebURL,
		WebURL:        adoRepo.WebURL,
		CloneURL:      adoRepo.RemoteURL,
		GitCloneURL:   adoRepo.RemoteURL,
		DefaultBranch: adoRepo.DefaultBranch,
		MainBranch:    adoRepo.DefaultBranch,
	}
}

type adoProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type adoRepo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	WebURL        string `json:"webUrl"`
	RemoteURL     string `json:"remoteUrl"`
	DefaultBranch string `json:"defaultBranch"`
}

// Register the Azure DevOps connector
func init() {
	scm.RegisterConnector(scm.KindAzureDevOps, func(settings *scm.ConnectorSettings) (scm.Connector, error) {
		return NewAzureDevOpsConnector(settings)
	})
}
