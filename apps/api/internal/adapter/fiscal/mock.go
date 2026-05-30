package fiscal

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// MockAdapter simulates a SENIAT fiscal machine. It always succeeds and
// generates sequential fiscal numbers. Used in development and integration tests.
type MockAdapter struct {
	counter atomic.Int64
}

// NewMockAdapter returns a MockAdapter ready for use.
func NewMockAdapter() *MockAdapter {
	return &MockAdapter{}
}

func (m *MockAdapter) Send(_ context.Context, req SendRequest) (*SendResponse, error) {
	seq := m.counter.Add(1)
	return &SendResponse{
		FiscalNumber:  fmt.Sprintf("00%07d", seq),
		MachineSerial: "MOCK-SN-001",
		ReportZNumber: int((seq-1)/100) + 1, // new Z report every 100 receipts
		IssuedAt:      time.Now().UTC(),
	}, nil
}
