// Package fiscal provides the SENIAT adapter interface and implementations.
// The Adapter interface decouples the application service from the physical
// fiscal machine protocol, allowing a mock for development/testing and a real
// TCP/serial driver in production without changing service code.
package fiscal

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SendLine represents a single product line to be sent to the fiscal machine.
type SendLine struct {
	Description string
	Quantity    float64
	UnitPrice   float64
	TaxRate     float64 // e.g. 0.16
}

// SendRequest is the payload sent to the SENIAT machine.
type SendRequest struct {
	InvoiceID        uuid.UUID
	CustomerIDType   string
	CustomerIDNumber string
	CustomerName     string
	Lines            []SendLine
}

// SendResponse contains the fields the SENIAT machine returns on success.
type SendResponse struct {
	FiscalNumber  string
	MachineSerial string
	ReportZNumber int
	IssuedAt      time.Time
}

// Adapter abstracts communication with the physical SENIAT fiscal machine.
// Implementations: MockAdapter (dev/test), TCPAdapter (production).
type Adapter interface {
	Send(ctx context.Context, req SendRequest) (*SendResponse, error)
}
