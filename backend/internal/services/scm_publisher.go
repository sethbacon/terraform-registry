package services

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/terraform-registry/terraform-registry/internal/crypto"
	"github.com/terraform-registry/terraform-registry/internal/db/models"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
	"github.com/terraform-registry/terraform-registry/internal/scm"
	"github.com/terraform-registry/terraform-registry/internal/storage"
)

// SCMPublisher handles automated publishing from SCM repositories
type SCMPublisher struct {
	scmRepo        *repositories.SCMRepository
	moduleRepo     *repositories.ModuleRepository
	storageBackend storage.Storage
	tokenCipher    *crypto.TokenCipher
	tempDir        string
}

// NewSCMPublisher creates a new SCM publisher
func NewSCMPublisher(scmRepo *repositories.SCMRepository, moduleRepo *repositories.ModuleRepository, storageBackend storage.Storage, tokenCipher *crypto.TokenCipher) *SCMPublisher {
	return &SCMPublisher{
		scmRepo:        scmRepo,
		moduleRepo:     moduleRepo,
		storageBackend: storageBackend,
		tokenCipher:    tokenCipher,
		tempDir:        os.TempDir(),
	}
}

// ProcessTagPush processes a tag push webhook and publishes a new version
func (p *SCMPublisher) ProcessTagPush(ctx context.Context, logID uuid.UUID, moduleSourceRepo *scm.ModuleSourceRepoRecord, hook *scm.IncomingHook, connector scm.Connector) {
	// Update webhook log to processing
	if err := p.scmRepo.UpdateWebhookLogState(ctx, logID, "processing", nil, nil); err != nil {
		return
	}

	// Extract version from tag name
	version := p.extractVersionFromTag(hook.TagName, moduleSourceRepo.TagPattern)
	if version == "" {
		errMsg := "could not extract version from tag"
		p.scmRepo.UpdateWebhookLogState(ctx, logID, "failed", &errMsg, nil)
		return
	}

	// TODO: Implement version conflict checking
	// Check if version already exists with a different commit
	// existingVersion, err := p.moduleRepo.GetVersion(ctx, moduleSourceRepo.ModuleID.String(), version)
	// For now, skip this check

	// Get OAuth token for downloading
	// TODO: Need user context to get token - for now we'll skip token-based download

	// Download source archive at the specific commit
	archivePath, checksum, err := p.downloadAndPackage(ctx, connector, nil, moduleSourceRepo.RepositoryOwner,
		moduleSourceRepo.RepositoryName, hook.CommitSHA, moduleSourceRepo.ModulePath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to download source: %v", err)
		p.scmRepo.UpdateWebhookLogState(ctx, logID, "failed", &errMsg, nil)
		return
	}
	defer os.Remove(archivePath)

	// Upload to storage
	file, err := os.Open(archivePath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to open archive: %v", err)
		p.scmRepo.UpdateWebhookLogState(ctx, logID, "failed", &errMsg, nil)
		return
	}
	defer file.Close()

	// TODO: Add proper module lookup by ID
	// For now use placeholder - GetByID method doesn't exist yet
	// module, err := p.moduleRepo.GetByID(ctx, moduleSourceRepo.ModuleID)
	module := &models.Module{
		ID:        moduleSourceRepo.ModuleID.String(),
		Namespace: "placeholder",
		Name:      "placeholder",
		System:    "placeholder",
	}

	storagePath := fmt.Sprintf("modules/%s/%s/%s/%s-%s.tar.gz",
		module.Namespace, module.Name, module.System, module.Name, version)

	// Get file size for upload
	fileInfo, err := os.Stat(archivePath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to stat temp file: %v", err)
		p.scmRepo.UpdateWebhookLogState(ctx, logID, "failed", &errMsg, nil)
		return
	}

	if _, err := p.storageBackend.Upload(ctx, storagePath, file, fileInfo.Size()); err != nil {
		errMsg := fmt.Sprintf("failed to upload to storage: %v", err)
		p.scmRepo.UpdateWebhookLogState(ctx, logID, "failed", &errMsg, nil)
		return
	}

	// Create module version record
	// TODO: Store sourceTag and sourceCommit in extended metadata
	// sourceTag := hook.TagName
	// sourceCommit := hook.CommitSHA
	versionID := uuid.New().String()

	moduleVersion := &models.ModuleVersion{
		ID:             versionID,
		ModuleID:       moduleSourceRepo.ModuleID.String(),
		Version:        version,
		StoragePath:    storagePath,
		StorageBackend: "default",
		Checksum:       checksum,
		CreatedAt:      time.Now(),
	}

	if err := p.moduleRepo.CreateVersion(ctx, moduleVersion); err != nil {
		errMsg := fmt.Sprintf("failed to create version: %v", err)
		p.scmRepo.UpdateWebhookLogState(ctx, logID, "failed", &errMsg, nil)
		return
	}

	// Update webhook log to success
	versionUUID, _ := uuid.Parse(versionID)
	p.scmRepo.UpdateWebhookLogState(ctx, logID, "completed", nil, &versionUUID)
}

// downloadAndPackage downloads the repository and creates a tarball
func (p *SCMPublisher) downloadAndPackage(ctx context.Context, connector scm.Connector, token *scm.OAuthToken,
	owner, repo, commitSHA, subpath string) (string, string, error) {

	// Download source archive
	archive, err := connector.DownloadSourceArchive(ctx, token, owner, repo, commitSHA, scm.ArchiveTarball)
	if err != nil {
		return "", "", fmt.Errorf("download failed: %w", err)
	}
	defer archive.Close()

	// Create temp directory for extraction
	tempDir := filepath.Join(p.tempDir, fmt.Sprintf("scm-publish-%s", uuid.New().String()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", "", err
	}
	defer os.RemoveAll(tempDir)

	// Extract archive
	if err := p.extractTarGz(archive, tempDir); err != nil {
		return "", "", fmt.Errorf("extraction failed: %w", err)
	}

	// Find the module subpath
	modulePath := filepath.Join(tempDir, subpath)
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		// Try to find it in subdirectories (GitHub/GitLab wrap in repo name directory)
		entries, _ := os.ReadDir(tempDir)
		if len(entries) == 1 && entries[0].IsDir() {
			modulePath = filepath.Join(tempDir, entries[0].Name(), subpath)
		}
	}

	// Validate module structure
	if err := p.validateModuleStructure(modulePath); err != nil {
		return "", "", fmt.Errorf("invalid module structure: %w", err)
	}

	// Create new tarball with commit SHA manifest
	outputPath := filepath.Join(p.tempDir, fmt.Sprintf("module-%s.tar.gz", uuid.New().String()))
	checksum, err := p.createImmutableTarball(modulePath, outputPath, commitSHA)
	if err != nil {
		return "", "", fmt.Errorf("packaging failed: %w", err)
	}

	return outputPath, checksum, nil
}

// extractTarGz extracts a tar.gz archive
func (p *SCMPublisher) extractTarGz(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Prevent path traversal
		target := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

// validateModuleStructure validates that the directory contains a valid Terraform module
func (p *SCMPublisher) validateModuleStructure(path string) error {
	// Check for at least one .tf file
	files, err := filepath.Glob(filepath.Join(path, "*.tf"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .tf files found in module directory")
	}
	return nil
}

// createImmutableTarball creates a tarball with a commit manifest
func (p *SCMPublisher) createImmutableTarball(srcPath, destPath, commitSHA string) (string, error) {
	outFile, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	// Calculate checksum while writing
	hasher := sha256.New()
	mw := io.MultiWriter(outFile, hasher)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add commit manifest file
	manifestContent := fmt.Sprintf("commit: %s\npublished: %s\n", commitSHA, time.Now().Format(time.RFC3339))
	manifestHeader := &tar.Header{
		Name:    ".terraform-registry-commit",
		Size:    int64(len(manifestContent)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(manifestHeader); err != nil {
		return "", err
	}
	if _, err := tw.Write([]byte(manifestContent)); err != nil {
		return "", err
	}

	// Add all module files
	err = filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})

	if err != nil {
		return "", err
	}

	// Close writers to flush
	tw.Close()
	gzw.Close()

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// extractVersionFromTag extracts a semantic version from a tag name
func (p *SCMPublisher) extractVersionFromTag(tag, glob string) string {
	// Convert glob pattern to regex
	pattern := strings.ReplaceAll(glob, "*", "(.*)")
	pattern = fmt.Sprintf("^%s$", pattern)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}

	matches := re.FindStringSubmatch(tag)
	if len(matches) < 2 {
		return ""
	}

	version := matches[1]

	// Remove leading 'v' if present
	version = strings.TrimPrefix(version, "v")

	// Validate semantic version format
	semverPattern := `^(\d+)\.(\d+)\.(\d+)(-[0-9A-Za-z-]+)?(\+[0-9A-Za-z-]+)?$`
	if matched, _ := regexp.MatchString(semverPattern, version); !matched {
		return ""
	}

	return version
}
