package version

import (
	"context"
	"dbsage/internal/models"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Service handles version checking in the background
type Service struct {
	checker   *Checker
	program   *tea.Program
	stopCh    chan struct{}
	checkDone chan struct{}
}

// NewService creates a new version checking service
func NewService() *Service {
	return &Service{
		checker:   GetDefaultChecker(),
		stopCh:    make(chan struct{}),
		checkDone: make(chan struct{}),
	}
}

// SetProgram sets the Bubble Tea program for sending messages
func (s *Service) SetProgram(program *tea.Program) {
	s.program = program
}

// Start begins the version checking service
func (s *Service) Start() {
	go s.backgroundCheck()
}

// Stop stops the version checking service
func (s *Service) Stop() {
	close(s.stopCh)
	<-s.checkDone
}

// CheckNow performs an immediate version check
func (s *Service) CheckNow(ctx context.Context) (*UpdateInfo, error) {
	return s.checker.CheckForUpdates(ctx)
}

// backgroundCheck runs a single version check at startup
func (s *Service) backgroundCheck() {
	defer close(s.checkDone)

	// Single check after a short delay to allow UI to initialize
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		s.performCheck()
		// Only check once at startup, no recurring checks

	case <-s.stopCh:
		return
	}
}

// performCheck performs a single version check and sends result to UI
func (s *Service) performCheck() {
	if s.program == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	updateInfo, err := s.checker.CheckForUpdates(ctx)
	if err != nil {
		// Log error but don't disturb the user
		log.Printf("Version check failed: %v", err)
		return
	}

	if updateInfo == nil || !updateInfo.HasUpdate {
		// No update available, nothing to report
		return
	}

	// Convert to UI model
	versionUpdateInfo := &models.VersionUpdateInfo{
		HasUpdate:      updateInfo.HasUpdate,
		CurrentVersion: updateInfo.CurrentVersion,
		LatestVersion:  updateInfo.LatestVersion,
		ReleaseURL:     updateInfo.ReleaseURL,
		ReleaseNotes:   updateInfo.ReleaseNotes,
	}

	// Send message to UI
	s.program.Send(models.VersionUpdateMsg{
		UpdateInfo: versionUpdateInfo,
	})
}

// Global service instance
var globalService *Service

// InitVersionService initializes the global version checking service
func InitVersionService() {
	globalService = NewService()
}

// StartVersionService starts the global version checking service
func StartVersionService(program *tea.Program) {
	if globalService != nil {
		globalService.SetProgram(program)
		globalService.Start()
	}
}

// StopVersionService stops the global version checking service
func StopVersionService() {
	if globalService != nil {
		globalService.Stop()
		globalService = nil
	}
}

// GetVersionService returns the global version service
func GetVersionService() *Service {
	return globalService
}
