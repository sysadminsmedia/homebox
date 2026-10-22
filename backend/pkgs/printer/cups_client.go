package printer

import (
	"context"
	"fmt"
)

// CUPSClient implements PrinterClient using the CUPS system
// This uses IPP under the hood since CUPS is IPP-based
type CUPSClient struct {
	printerName string
	ippClient   *IPPClient
}

// NewCUPSClient creates a new CUPS client for the given printer name
// The printer name should be the CUPS printer name (e.g., "Brother_QL820NWB")
func NewCUPSClient(printerName string) (*CUPSClient, error) {
	// CUPS runs on localhost:631 by default
	// The IPP address for a CUPS printer is: ipp://localhost:631/printers/{printer_name}
	ippAddress := fmt.Sprintf("ipp://localhost:631/printers/%s", printerName)

	ippClient, err := NewIPPClient(ippAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create CUPS client: %w", err)
	}

	return &CUPSClient{
		printerName: printerName,
		ippClient:   ippClient,
	}, nil
}

// GetPrinterInfo retrieves printer capabilities and status
func (c *CUPSClient) GetPrinterInfo(ctx context.Context) (*PrinterInfo, error) {
	return c.ippClient.GetPrinterInfo(ctx)
}

// Print sends a document to the printer
func (c *CUPSClient) Print(ctx context.Context, job *PrintJob) (*PrintResult, error) {
	return c.ippClient.Print(ctx, job)
}

// GetJobStatus checks the status of a print job
func (c *CUPSClient) GetJobStatus(ctx context.Context, jobID int) (*JobStatus, error) {
	return c.ippClient.GetJobStatus(ctx, jobID)
}

// CancelJob cancels a pending print job
func (c *CUPSClient) CancelJob(ctx context.Context, jobID int) error {
	return c.ippClient.CancelJob(ctx, jobID)
}
