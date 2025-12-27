// Package printer provides interfaces and implementations for direct label printing
// via IPP (Internet Printing Protocol) and CUPS.
package printer

import (
	"context"
)

// PrinterType represents the type of printer connection
type PrinterType string

const (
	PrinterTypeIPP           PrinterType = "ipp"
	PrinterTypeCUPS          PrinterType = "cups"
	PrinterTypeBrotherRaster PrinterType = "brother_raster"
)

// PrinterStatus represents the current status of a printer
type PrinterStatus string

const (
	StatusOnline  PrinterStatus = "online"
	StatusOffline PrinterStatus = "offline"
	StatusUnknown PrinterStatus = "unknown"
)

// PrinterInfo contains information about a printer
type PrinterInfo struct {
	Name           string
	Make           string
	Model          string
	State          PrinterStatus
	StateMessage   string
	MediaSupported []string
	MediaReady     []string
	SupportsColor  bool
	SupportsDuplex bool
	MaxDPI         int
}

// PrintJob represents a print job to be sent to a printer
type PrintJob struct {
	DocumentName string
	ContentType  string // "image/png", "application/pdf", etc.
	Data         []byte
	Copies       int
	MediaSize    string // Media size (e.g., "oe_62mm-x-29mm")
	Orientation  int    // 3=portrait, 4=landscape
}

// PrintResult contains the result of a print operation
type PrintResult struct {
	JobID   int
	Success bool
	Message string
}

// JobStatus represents the status of a print job
type JobStatus struct {
	JobID        int
	State        string
	StateMessage string
	Completed    bool
}

// PrinterClient is the interface for printer communication
type PrinterClient interface {
	// GetPrinterInfo retrieves printer capabilities and status
	GetPrinterInfo(ctx context.Context) (*PrinterInfo, error)

	// Print sends a document to the printer
	Print(ctx context.Context, job *PrintJob) (*PrintResult, error)

	// GetJobStatus checks the status of a print job
	GetJobStatus(ctx context.Context, jobID int) (*JobStatus, error)

	// CancelJob cancels a pending print job
	CancelJob(ctx context.Context, jobID int) error
}

// NewPrinterClient creates a new printer client based on the printer type and address
func NewPrinterClient(printerType PrinterType, address string) (PrinterClient, error) {
	switch printerType {
	case PrinterTypeIPP:
		return NewIPPClient(address)
	case PrinterTypeCUPS:
		return NewCUPSClient(address)
	case PrinterTypeBrotherRaster:
		return NewBrotherRasterClient(address)
	default:
		return NewIPPClient(address)
	}
}
