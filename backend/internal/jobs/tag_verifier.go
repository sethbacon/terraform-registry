package jobs

import (
	"context"
	"log"
	"time"

	"github.com/terraform-registry/terraform-registry/internal/crypto"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
)

// TagVerifier periodically verifies that git tags haven't been moved
type TagVerifier struct {
	scmRepo     *repositories.SCMRepository
	moduleRepo  *repositories.ModuleRepository
	tokenCipher *crypto.TokenCipher
	interval    time.Duration
	stopChan    chan struct{}
}

// NewTagVerifier creates a new tag verification job
func NewTagVerifier(scmRepo *repositories.SCMRepository, moduleRepo *repositories.ModuleRepository, tokenCipher *crypto.TokenCipher, intervalHours int) *TagVerifier {
	if intervalHours <= 0 {
		intervalHours = 24 // Default to daily
	}

	return &TagVerifier{
		scmRepo:     scmRepo,
		moduleRepo:  moduleRepo,
		tokenCipher: tokenCipher,
		interval:    time.Duration(intervalHours) * time.Hour,
		stopChan:    make(chan struct{}),
	}
}

// Start begins the tag verification job
func (v *TagVerifier) Start(ctx context.Context) {
	ticker := time.NewTicker(v.interval)
	defer ticker.Stop()

	log.Printf("Tag verifier started with interval: %v", v.interval)

	// Run immediately on start
	v.runVerification(ctx)

	for {
		select {
		case <-ticker.C:
			v.runVerification(ctx)
		case <-v.stopChan:
			log.Println("Tag verifier stopped")
			return
		case <-ctx.Done():
			log.Println("Tag verifier context cancelled")
			return
		}
	}
}

// Stop stops the tag verification job
func (v *TagVerifier) Stop() {
	close(v.stopChan)
}

// runVerification performs a verification run
func (v *TagVerifier) runVerification(ctx context.Context) {
	log.Println("Starting tag verification run")

	// TODO: Implement GetAllWithSourceCommit method on ModuleRepository
	// This would query module_versions table for versions with source_commit populated
	// Once implemented, this job will:
	// 1. Fetch all module versions with SCM source info
	// 2. For each version, re-query the SCM tag to get current commit
	// 3. Compare current commit with stored commit
	// 4. Create immutability alerts for any mismatches

	log.Println("Tag verification skipped - GetAllWithSourceCommit not implemented")
	log.Println("Tag verification run completed: checked 0 tags, found 0 violations")
}
