package ssh

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TransferContext represents a file transfer operation
type TransferContext struct {
	ID         string
	SessionID  string
	Type       string // "upload" or "download"
	Files      []string
	ctx        context.Context
	cancel     context.CancelFunc
	progress   TransferProgress
	mu         sync.Mutex
	startTime  time.Time
	lastUpdate time.Time
}

// TransferManager manages multiple file transfer operations
type TransferManager struct {
	mu        sync.RWMutex
	transfers map[string]*TransferContext
}

// NewTransferManager creates a new transfer manager
func NewTransferManager() *TransferManager {
	return &TransferManager{
		transfers: make(map[string]*TransferContext),
	}
}

// StartTransfer creates and registers a new transfer
func (tm *TransferManager) StartTransfer(sessionID, transferType string, files []string) (*TransferContext, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Generate unique transfer ID
	transferID := fmt.Sprintf("transfer_%d_%d", time.Now().UnixNano(), len(tm.transfers))

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	transfer := &TransferContext{
		ID:         transferID,
		SessionID:  sessionID,
		Type:       transferType,
		Files:      files,
		ctx:        ctx,
		cancel:     cancel,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
		progress: TransferProgress{
			TransferID: transferID,
			SessionID:  sessionID,
			Status:     "pending",
		},
	}

	tm.transfers[transferID] = transfer

	return transfer, nil
}

// GetTransfer retrieves a transfer by ID
func (tm *TransferManager) GetTransfer(transferID string) (*TransferContext, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	transfer, exists := tm.transfers[transferID]
	return transfer, exists
}

// UpdateProgress updates the progress of a transfer
func (tm *TransferManager) UpdateProgress(transferID string, progress TransferProgress) error {
	tm.mu.RLock()
	transfer, exists := tm.transfers[transferID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("transfer not found: %s", transferID)
	}

	transfer.mu.Lock()
	defer transfer.mu.Unlock()

	transfer.progress = progress
	transfer.lastUpdate = time.Now()

	return nil
}

// GetProgress retrieves the current progress of a transfer
func (tm *TransferManager) GetProgress(transferID string) (TransferProgress, error) {
	tm.mu.RLock()
	transfer, exists := tm.transfers[transferID]
	tm.mu.RUnlock()

	if !exists {
		return TransferProgress{}, fmt.Errorf("transfer not found: %s", transferID)
	}

	transfer.mu.Lock()
	defer transfer.mu.Unlock()

	return transfer.progress, nil
}

// CancelTransfer cancels an ongoing transfer
func (tm *TransferManager) CancelTransfer(transferID string) error {
	tm.mu.RLock()
	transfer, exists := tm.transfers[transferID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("transfer not found: %s", transferID)
	}

	// Cancel the context
	transfer.cancel()

	// Update status
	transfer.mu.Lock()
	transfer.progress.Status = "cancelled"
	transfer.progress.Error = "Transfer cancelled by user"
	transfer.mu.Unlock()

	return nil
}

// CleanupTransfer removes a transfer from the manager
func (tm *TransferManager) CleanupTransfer(transferID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if transfer, exists := tm.transfers[transferID]; exists {
		// Cancel context if still active
		transfer.cancel()
		delete(tm.transfers, transferID)
	}
}

// ListTransfers returns all active transfer IDs
func (tm *TransferManager) ListTransfers() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	ids := make([]string, 0, len(tm.transfers))
	for id := range tm.transfers {
		ids = append(ids, id)
	}
	return ids
}

// ListSessionTransfers returns transfer IDs for a specific session
func (tm *TransferManager) ListSessionTransfers(sessionID string) []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	ids := make([]string, 0)
	for _, transfer := range tm.transfers {
		if transfer.SessionID == sessionID {
			ids = append(ids, transfer.ID)
		}
	}
	return ids
}

// CleanupSessionTransfers cancels and removes all transfers for a session
func (tm *TransferManager) CleanupSessionTransfers(sessionID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for id, transfer := range tm.transfers {
		if transfer.SessionID == sessionID {
			transfer.cancel()
			delete(tm.transfers, id)
		}
	}
}

// Context returns the context for a transfer
func (tc *TransferContext) Context() context.Context {
	return tc.ctx
}

// IsCancelled checks if the transfer has been cancelled
func (tc *TransferContext) IsCancelled() bool {
	select {
	case <-tc.ctx.Done():
		return true
	default:
		return false
	}
}

// GetProgress returns the current progress
func (tc *TransferContext) GetProgress() TransferProgress {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	return tc.progress
}

// SetProgress updates the progress
func (tc *TransferContext) SetProgress(progress TransferProgress) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.progress = progress
	tc.lastUpdate = time.Now()
}
