package printer

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/phin1x/go-ipp"
)

// IPPClient implements PrinterClient using the IPP protocol
// This client works directly with IPP printers (not just CUPS)
type IPPClient struct {
	printerURI  string // The full printer URI (ipp://host:port/path)
	httpURL     string // The HTTP URL to send requests to
	httpClient  *http.Client
	workingPath string   // Cached working path after successful connection
	pathsToTry  []string // Paths to try if the configured path doesn't work
}

// Common IPP paths to try if the configured path doesn't work
var commonIPPPaths = []string{
	"/ipp",
	"/ipp/print",
	"/ipp/printer",
	"/",
}

// NewIPPClient creates a new IPP client for the given printer address
// Address should be in the format: ipp://hostname:port/path or ipps://hostname:port/path
func NewIPPClient(address string) (*IPPClient, error) {
	// Parse the address to extract components
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("invalid IPP address: %w", err)
	}

	// Determine if using TLS
	useTLS := strings.HasPrefix(address, "ipps://")

	// Build HTTP URL from IPP URI
	scheme := "http"
	if useTLS {
		scheme = "https"
	}

	// Ensure we have a port
	host := u.Host
	if !strings.Contains(host, ":") {
		host += ":631"
	}

	// Build list of paths to try - user's path first, then common paths
	userPath := u.Path
	if userPath == "" {
		userPath = "/ipp"
	}

	pathsToTry := []string{userPath}
	for _, p := range commonIPPPaths {
		if p != userPath {
			pathsToTry = append(pathsToTry, p)
		}
	}

	// Create HTTP client
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return &IPPClient{
		printerURI: address,
		httpURL:    fmt.Sprintf("%s://%s", scheme, host),
		httpClient: httpClient,
		pathsToTry: pathsToTry,
	}, nil
}

// sendIPPRequest sends an IPP request and returns the response
func (c *IPPClient) sendIPPRequest(path string, req *ipp.Request, fileData io.Reader, fileSize int) (*ipp.Response, error) {
	// Encode the request
	payload, err := req.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode IPP request: %w", err)
	}

	// Build the request body
	var body io.Reader
	size := len(payload)
	if fileData != nil && fileSize > 0 {
		size += fileSize
		body = io.MultiReader(bytes.NewReader(payload), fileData)
	} else {
		body = bytes.NewReader(payload)
	}

	// Build the HTTP URL
	httpURL := c.httpURL + path

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", httpURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", ipp.ContentTypeIPP)
	httpReq.Header.Set("Content-Length", fmt.Sprintf("%d", size))

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("got http code %d", resp.StatusCode)
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Decode IPP response
	ippResp, err := ipp.NewResponseDecoder(bytes.NewReader(respBody)).Decode(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IPP response: %w", err)
	}

	return ippResp, nil
}

// tryPaths attempts an IPP request on multiple paths until one succeeds
func (c *IPPClient) tryPaths(makeRequest func(path, printerURI string) (*ipp.Response, error)) (*ipp.Response, error) {
	// If we have a cached working path, try it first
	if c.workingPath != "" {
		printerURI := c.buildPrinterURI(c.workingPath)
		resp, err := makeRequest(c.workingPath, printerURI)
		if err == nil {
			return resp, nil
		}
		// Clear cached path if it failed
		c.workingPath = ""
	}

	// Try each path until one works
	var lastErr error
	for _, path := range c.pathsToTry {
		printerURI := c.buildPrinterURI(path)
		resp, err := makeRequest(path, printerURI)
		if err == nil {
			// Cache the working path
			c.workingPath = path
			return resp, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("all paths failed (tried %v): %w", c.pathsToTry, lastErr)
}

// buildPrinterURI constructs the printer URI for IPP attributes
func (c *IPPClient) buildPrinterURI(path string) string {
	u, _ := url.Parse(c.printerURI)
	scheme := u.Scheme
	if scheme == "" {
		scheme = "ipp"
	}
	host := u.Host
	if !strings.Contains(host, ":") {
		host += ":631"
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}

// getAttrValue is a helper to safely get the first value of an attribute
func getAttrValue(attrs ipp.Attributes, name string) interface{} {
	if attrSlice, ok := attrs[name]; ok && len(attrSlice) > 0 {
		return attrSlice[0].Value
	}
	return nil
}

// getAttrValues is a helper to get all values of an attribute
func getAttrValues(attrs ipp.Attributes, name string) []interface{} {
	if attrSlice, ok := attrs[name]; ok {
		values := make([]interface{}, len(attrSlice))
		for i, attr := range attrSlice {
			values[i] = attr.Value
		}
		return values
	}
	return nil
}

// GetPrinterInfo retrieves printer capabilities and status
func (c *IPPClient) GetPrinterInfo(ctx context.Context) (*PrinterInfo, error) {
	resp, err := c.tryPaths(func(path, printerURI string) (*ipp.Response, error) {
		req := ipp.NewRequest(ipp.OperationGetPrinterAttributes, 1)
		req.OperationAttributes[ipp.AttributePrinterURI] = printerURI
		req.OperationAttributes[ipp.AttributeRequestedAttributes] = []string{
			"printer-name",
			"printer-make-and-model",
			"printer-state",
			"printer-state-message",
			"media-supported",
			"media-ready",
			"color-supported",
			"sides-supported",
			"printer-resolution-supported",
		}

		return c.sendIPPRequest(path, req, nil, 0)
	})

	if err != nil {
		return nil, err
	}

	if len(resp.PrinterAttributes) == 0 {
		return nil, fmt.Errorf("no printer attributes returned")
	}

	return c.parsePrinterInfo(resp.PrinterAttributes[0])
}

// parsePrinterInfo converts IPP attributes to PrinterInfo
func (c *IPPClient) parsePrinterInfo(attrs ipp.Attributes) (*PrinterInfo, error) {
	info := &PrinterInfo{
		State: StatusUnknown,
	}

	// Parse printer name
	if v := getAttrValue(attrs, "printer-name"); v != nil {
		if s, ok := v.(string); ok {
			info.Name = s
		}
	}

	// Parse make and model
	if v := getAttrValue(attrs, "printer-make-and-model"); v != nil {
		if s, ok := v.(string); ok {
			parts := strings.SplitN(s, " ", 2)
			if len(parts) > 0 {
				info.Make = parts[0]
			}
			if len(parts) > 1 {
				info.Model = parts[1]
			}
		}
	}

	// Parse printer state
	if v := getAttrValue(attrs, "printer-state"); v != nil {
		var state int
		switch val := v.(type) {
		case int:
			state = val
		case int8:
			state = int(val)
		case int16:
			state = int(val)
		case int32:
			state = int(val)
		case int64:
			state = int(val)
		}
		switch state {
		case 3: // idle
			info.State = StatusOnline
		case 4: // processing
			info.State = StatusOnline
		case 5: // stopped
			info.State = StatusOffline
		default:
			info.State = StatusUnknown
		}
	}

	// Parse state message
	if v := getAttrValue(attrs, "printer-state-message"); v != nil {
		if s, ok := v.(string); ok {
			info.StateMessage = s
		}
	}

	// Parse media supported
	if values := getAttrValues(attrs, "media-supported"); values != nil {
		for _, v := range values {
			if s, ok := v.(string); ok {
				info.MediaSupported = append(info.MediaSupported, s)
			}
		}
	}

	// Parse media ready
	if values := getAttrValues(attrs, "media-ready"); values != nil {
		for _, v := range values {
			if s, ok := v.(string); ok {
				info.MediaReady = append(info.MediaReady, s)
			}
		}
	}

	// Parse color support
	if v := getAttrValue(attrs, "color-supported"); v != nil {
		if b, ok := v.(bool); ok {
			info.SupportsColor = b
		}
	}

	// Parse duplex support
	if values := getAttrValues(attrs, "sides-supported"); values != nil {
		for _, v := range values {
			if s, ok := v.(string); ok {
				if strings.Contains(s, "duplex") {
					info.SupportsDuplex = true
					break
				}
			}
		}
	}

	return info, nil
}

// Print sends a document to the printer
func (c *IPPClient) Print(ctx context.Context, job *PrintJob) (*PrintResult, error) {
	// Set default values
	copies := job.Copies
	if copies <= 0 {
		copies = 1
	}

	contentType := job.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	documentName := job.DocumentName
	if documentName == "" {
		documentName = "label"
	}

	resp, err := c.tryPaths(func(path, printerURI string) (*ipp.Response, error) {
		req := ipp.NewRequest(ipp.OperationPrintJob, 1)
		req.OperationAttributes[ipp.AttributePrinterURI] = printerURI
		req.OperationAttributes[ipp.AttributeJobName] = documentName
		req.OperationAttributes[ipp.AttributeDocumentFormat] = contentType
		req.OperationAttributes[ipp.AttributeCopies] = copies

		if job.MediaSize != "" {
			req.JobAttributes["media"] = job.MediaSize
		}

		if job.Orientation > 0 {
			req.JobAttributes["orientation-requested"] = job.Orientation
		}

		return c.sendIPPRequest(path, req, bytes.NewReader(job.Data), len(job.Data))
	})

	if err != nil {
		return &PrintResult{
			Success: false,
			Message: fmt.Sprintf("Print failed: %v", err),
		}, err
	}

	// Extract job ID from response
	jobID := 0
	if len(resp.JobAttributes) > 0 {
		if v := getAttrValue(resp.JobAttributes[0], "job-id"); v != nil {
			switch val := v.(type) {
			case int:
				jobID = val
			case int32:
				jobID = int(val)
			}
		}
	}

	return &PrintResult{
		JobID:   jobID,
		Success: true,
		Message: fmt.Sprintf("Print job %d submitted successfully", jobID),
	}, nil
}

// GetJobStatus checks the status of a print job
func (c *IPPClient) GetJobStatus(ctx context.Context, jobID int) (*JobStatus, error) {
	path := c.workingPath
	if path == "" {
		path = "/ipp"
	}
	printerURI := c.buildPrinterURI(path)
	jobURI := fmt.Sprintf("%s/jobs/%d", strings.TrimSuffix(printerURI, "/"), jobID)

	req := ipp.NewRequest(ipp.OperationGetJobAttributes, 1)
	req.OperationAttributes["job-uri"] = jobURI
	req.OperationAttributes[ipp.AttributeRequestedAttributes] = []string{
		"job-state",
		"job-state-message",
	}

	resp, err := c.sendIPPRequest(path, req, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}

	if len(resp.JobAttributes) == 0 {
		return nil, fmt.Errorf("no job attributes returned")
	}

	attrs := resp.JobAttributes[0]
	status := &JobStatus{
		JobID: jobID,
	}

	// Parse job state
	if v := getAttrValue(attrs, "job-state"); v != nil {
		var state int
		switch val := v.(type) {
		case int:
			state = val
		case int8:
			state = int(val)
		case int16:
			state = int(val)
		case int32:
			state = int(val)
		case int64:
			state = int(val)
		}
		switch state {
		case 3: // pending
			status.State = "pending"
		case 4: // pending-held
			status.State = "held"
		case 5: // processing
			status.State = "processing"
		case 6: // processing-stopped
			status.State = "stopped"
		case 7: // canceled
			status.State = "canceled"
			status.Completed = true
		case 8: // aborted
			status.State = "aborted"
			status.Completed = true
		case 9: // completed
			status.State = "completed"
			status.Completed = true
		default:
			status.State = "unknown"
		}
	}

	// Parse state message
	if v := getAttrValue(attrs, "job-state-message"); v != nil {
		if s, ok := v.(string); ok {
			status.StateMessage = s
		}
	}

	return status, nil
}

// CancelJob cancels a pending print job
func (c *IPPClient) CancelJob(ctx context.Context, jobID int) error {
	path := c.workingPath
	if path == "" {
		path = "/ipp"
	}
	printerURI := c.buildPrinterURI(path)
	jobURI := fmt.Sprintf("%s/jobs/%d", strings.TrimSuffix(printerURI, "/"), jobID)

	req := ipp.NewRequest(ipp.OperationCancelJob, 1)
	req.OperationAttributes["job-uri"] = jobURI

	_, err := c.sendIPPRequest(path, req, nil, 0)
	return err
}
