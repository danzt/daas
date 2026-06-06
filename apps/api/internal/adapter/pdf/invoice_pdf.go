// Package pdf provides PDF generation for DaaS invoices.
// Uses github.com/go-pdf/fpdf (pure Go, no system dependencies).
package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/danzt/daas/api/internal/domain/fiscal"
	"github.com/danzt/daas/api/internal/domain/invoice"
)

// TenantInfo holds the tenant fields needed for the invoice PDF header.
type TenantInfo struct {
	Name     string
	FiscalID string
	Country  string
}

// ─── Color palette (matches DaaS design system) ───────────────────────────────

const (
	purple1R, purple1G, purple1B = 124, 58, 237  // #7C3AED — primary
	gray1R, gray1G, gray1B       = 245, 245, 247 // table header bg
	gray2R, gray2G, gray2B       = 100, 100, 110 // muted text
	borderR, borderG, borderB    = 220, 220, 228 // table border
	black1R, black1G, black1B    = 25, 25, 35    // body text
)

// ─── Public API ───────────────────────────────────────────────────────────────

// BuildInternalInvoicePDF renders an InternalInvoice as a PDF and returns the bytes.
func BuildInternalInvoicePDF(tenant TenantInfo, inv *invoice.InternalInvoice) ([]byte, error) {
	p := newPDF()
	p.AddPage()

	drawHeader(p, tenant, inv.Correlative, formatDate(inv.CreatedAt), string(inv.Status))
	drawCustomer(p, inv.CustomerName, string(inv.CustomerIDType), inv.CustomerIDNumber, inv.Notes)

	// Line items — no tax column for internal invoices
	cols := []colDef{
		{header: "Descripción", width: 90, align: "L"},
		{header: "Cant.", width: 22, align: "R"},
		{header: "P. Unit.", width: 33, align: "R"},
		{header: "Subtotal", width: 35, align: "R"},
	}
	rows := make([][]string, len(inv.Lines))
	for i, l := range inv.Lines {
		rows[i] = []string{
			l.Description,
			fmt.Sprintf("%.2f", l.Quantity),
			fmt.Sprintf("%.2f $", l.UnitPrice),
			fmt.Sprintf("%.2f $", l.Subtotal),
		}
	}
	drawTable(p, cols, rows)

	// Totals
	drawTotalsInternal(p, inv.Subtotal, inv.Total)

	return output(p)
}

// BuildFiscalInvoicePDF renders a FiscalInvoice as a PDF and returns the bytes.
func BuildFiscalInvoicePDF(tenant TenantInfo, inv *fiscal.FiscalInvoice) ([]byte, error) {
	p := newPDF()
	p.AddPage()

	// Header — use fiscal number if issued, otherwise "BORRADOR"
	docNum := "BORRADOR"
	if inv.FiscalNumber != nil && *inv.FiscalNumber != "" {
		docNum = *inv.FiscalNumber
	}
	drawHeader(p, tenant, docNum, formatDate(inv.CreatedAt), string(inv.Status))
	drawFiscalMeta(p, inv)
	drawCustomer(p, inv.CustomerName, string(inv.CustomerIDType), inv.CustomerIDNumber, inv.Notes)

	// Line items — includes IVA columns
	cols := []colDef{
		{header: "Descripción", width: 72, align: "L"},
		{header: "Cant.", width: 20, align: "R"},
		{header: "P. Unit.", width: 28, align: "R"},
		{header: "IVA %", width: 20, align: "R"},
		{header: "IVA", width: 20, align: "R"},
		{header: "Subtotal", width: 20, align: "R"},
	}
	rows := make([][]string, len(inv.Lines))
	for i, l := range inv.Lines {
		rows[i] = []string{
			l.Description,
			fmt.Sprintf("%.2f", l.Quantity),
			fmt.Sprintf("%.2f", l.UnitPrice),
			fmt.Sprintf("%.0f%%", l.TaxRate*100),
			fmt.Sprintf("%.2f", l.TaxAmount),
			fmt.Sprintf("%.2f", l.Subtotal),
		}
	}
	drawTable(p, cols, rows)

	drawTotalsFiscal(p, inv.SubtotalBase, inv.TaxAmount, inv.Total)

	return output(p)
}

// ─── Layout helpers ───────────────────────────────────────────────────────────

func newPDF() *fpdf.Fpdf {
	p := fpdf.New("P", "mm", "A4", "")
	p.SetMargins(15, 15, 15)
	p.SetAutoPageBreak(true, 20)
	return p
}

func output(p *fpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return buf.Bytes(), nil
}

func formatDate(t time.Time) string {
	return t.Format("02/01/2006")
}

// drawHeader renders the top band: purple accent bar, company name, document number.
func drawHeader(p *fpdf.Fpdf, tenant TenantInfo, docNum, date, status string) {
	pageW, _ := p.GetPageSize()
	leftMargin, _, rightMargin, _ := p.GetMargins()
	contentW := pageW - leftMargin - rightMargin

	// Purple accent bar (left edge)
	p.SetFillColor(purple1R, purple1G, purple1B)
	p.Rect(leftMargin, 15, 5, 38, "F")

	// Company name (large, bold)
	p.SetXY(leftMargin+8, 17)
	p.SetFont("Helvetica", "B", 16)
	p.SetTextColor(black1R, black1G, black1B)
	p.CellFormat(contentW*0.55, 8, tenant.Name, "", 1, "L", false, 0, "")

	// RIF + country
	p.SetX(leftMargin + 8)
	p.SetFont("Helvetica", "", 9)
	p.SetTextColor(gray2R, gray2G, gray2B)
	rifLine := ""
	if tenant.FiscalID != "" {
		rifLine = "RIF: " + tenant.FiscalID
		if tenant.Country != "" {
			rifLine += "   •   " + tenant.Country
		}
	} else if tenant.Country != "" {
		rifLine = tenant.Country
	}
	p.CellFormat(contentW*0.55, 5, rifLine, "", 1, "L", false, 0, "")

	// Document number (right side, large)
	p.SetXY(leftMargin+contentW*0.55, 17)
	p.SetFont("Helvetica", "B", 13)
	p.SetTextColor(purple1R, purple1G, purple1B)
	p.CellFormat(contentW*0.45, 8, docNum, "", 1, "R", false, 0, "")

	// Date
	p.SetXY(leftMargin+contentW*0.55, 25)
	p.SetFont("Helvetica", "", 9)
	p.SetTextColor(gray2R, gray2G, gray2B)
	p.CellFormat(contentW*0.45, 5, "Fecha: "+date, "", 1, "R", false, 0, "")

	// Status badge
	p.SetXY(leftMargin+contentW*0.55, 30)
	statusLabel, sr, sg, sb := statusStyle(status)
	p.SetFont("Helvetica", "B", 8)
	p.SetTextColor(sr, sg, sb)
	p.CellFormat(contentW*0.45, 5, statusLabel, "", 1, "R", false, 0, "")

	p.SetTextColor(black1R, black1G, black1B)
	p.Ln(6)
	drawHRule(p)
	p.Ln(4)
}

// drawFiscalMeta renders the SENIAT machine/Z-number info for fiscal invoices.
func drawFiscalMeta(p *fpdf.Fpdf, inv *fiscal.FiscalInvoice) {
	if inv.MachineSerial == nil && inv.ReportZNumber == nil {
		return
	}
	p.SetFont("Helvetica", "", 8)
	p.SetTextColor(gray2R, gray2G, gray2B)
	meta := "FACTURA FISCAL SENIAT"
	if inv.MachineSerial != nil {
		meta += "   Máquina: " + *inv.MachineSerial
	}
	if inv.ReportZNumber != nil {
		meta += fmt.Sprintf("   Reporte Z: %d", *inv.ReportZNumber)
	}
	p.CellFormat(0, 5, meta, "", 1, "L", false, 0, "")
	p.SetTextColor(black1R, black1G, black1B)
	p.Ln(2)
}

// drawCustomer renders the "FACTURADO A" section.
func drawCustomer(p *fpdf.Fpdf, name, idType, idNumber, notes string) {
	p.SetFont("Helvetica", "B", 8)
	p.SetTextColor(gray2R, gray2G, gray2B)
	p.CellFormat(0, 5, "FACTURADO A", "", 1, "L", false, 0, "")

	p.SetFont("Helvetica", "B", 10)
	p.SetTextColor(black1R, black1G, black1B)
	customerDisplay := name
	if customerDisplay == "" {
		customerDisplay = "Consumidor Final"
	}
	p.CellFormat(0, 6, customerDisplay, "", 1, "L", false, 0, "")

	if idType != "" && idType != "anonymous" && idNumber != "" {
		p.SetFont("Helvetica", "", 9)
		p.SetTextColor(gray2R, gray2G, gray2B)
		p.CellFormat(0, 5, idTypeLabel(idType)+": "+idNumber, "", 1, "L", false, 0, "")
	}
	if notes != "" {
		p.SetFont("Helvetica", "I", 8)
		p.SetTextColor(gray2R, gray2G, gray2B)
		p.CellFormat(0, 5, notes, "", 1, "L", false, 0, "")
	}
	p.SetTextColor(black1R, black1G, black1B)
	p.Ln(4)
}

// colDef describes one table column.
type colDef struct {
	header string
	width  float64
	align  string
}

// drawTable renders a line-items table with a header row.
func drawTable(p *fpdf.Fpdf, cols []colDef, rows [][]string) {
	rowH := 7.0
	headerH := 7.0

	// Header row
	p.SetFillColor(gray1R, gray1G, gray1B)
	p.SetDrawColor(borderR, borderG, borderB)
	p.SetLineWidth(0.3)
	p.SetFont("Helvetica", "B", 8)
	p.SetTextColor(gray2R, gray2G, gray2B)
	for _, col := range cols {
		p.CellFormat(col.width, headerH, col.header, "TB", 0, col.align, true, 0, "")
	}
	p.Ln(headerH)

	// Data rows
	p.SetFont("Helvetica", "", 9)
	p.SetTextColor(black1R, black1G, black1B)
	p.SetFillColor(255, 255, 255)
	for i, row := range rows {
		// Alternating row background
		if i%2 == 0 {
			p.SetFillColor(250, 250, 252)
		} else {
			p.SetFillColor(255, 255, 255)
		}
		for j, cell := range row {
			p.CellFormat(cols[j].width, rowH, cell, "B", 0, cols[j].align, true, 0, "")
		}
		p.Ln(rowH)
	}
	p.Ln(2)
}

// drawTotalsInternal renders the totals block for internal invoices (no tax).
func drawTotalsInternal(p *fpdf.Fpdf, subtotal, total float64) {
	pageW, _ := p.GetPageSize()
	_, _, rightMargin, _ := p.GetMargins()
	labelW := 45.0
	valueW := 35.0
	x := pageW - rightMargin - labelW - valueW

	if subtotal != total {
		drawTotalRow(p, x, labelW, valueW, "Subtotal:", fmt.Sprintf("%.2f $", subtotal), false)
	}
	drawTotalRow(p, x, labelW, valueW, "TOTAL:", fmt.Sprintf("%.2f $", total), true)
}

// drawTotalsFiscal renders the totals block for fiscal invoices (with IVA breakdown).
func drawTotalsFiscal(p *fpdf.Fpdf, subtotalBase, taxAmount, total float64) {
	pageW, _ := p.GetPageSize()
	_, _, rightMargin, _ := p.GetMargins()
	labelW := 45.0
	valueW := 35.0
	x := pageW - rightMargin - labelW - valueW

	drawTotalRow(p, x, labelW, valueW, "Base imponible:", fmt.Sprintf("%.2f $", subtotalBase), false)
	drawTotalRow(p, x, labelW, valueW, "IVA:", fmt.Sprintf("%.2f $", taxAmount), false)
	drawTotalRow(p, x, labelW, valueW, "TOTAL:", fmt.Sprintf("%.2f $", total), true)
}

func drawTotalRow(p *fpdf.Fpdf, x, labelW, valueW float64, label, value string, highlight bool) {
	p.SetX(x)
	if highlight {
		p.SetFont("Helvetica", "B", 10)
		p.SetTextColor(purple1R, purple1G, purple1B)
	} else {
		p.SetFont("Helvetica", "", 9)
		p.SetTextColor(gray2R, gray2G, gray2B)
	}
	p.CellFormat(labelW, 7, label, "", 0, "R", false, 0, "")
	p.CellFormat(valueW, 7, value, "", 1, "R", false, 0, "")
	p.SetTextColor(black1R, black1G, black1B)
}

func drawHRule(p *fpdf.Fpdf) {
	pageW, _ := p.GetPageSize()
	leftMargin, _, rightMargin, _ := p.GetMargins()
	p.SetDrawColor(borderR, borderG, borderB)
	p.SetLineWidth(0.3)
	y := p.GetY()
	p.Line(leftMargin, y, pageW-rightMargin, y)
}

// ─── Label helpers ────────────────────────────────────────────────────────────

func statusStyle(status string) (label string, r, g, b int) {
	switch status {
	case "issued":
		return "EMITIDA", 22, 163, 74 // green
	case "cancelled":
		return "ANULADA", 220, 38, 38 // red
	case "failed":
		return "FALLIDA", 234, 88, 12 // orange
	case "pending_fiscal":
		return "PENDIENTE", 202, 138, 4 // yellow
	default:
		return "BORRADOR", gray2R, gray2G, gray2B
	}
}

func idTypeLabel(t string) string {
	switch t {
	case "cedula":
		return "Cédula"
	case "rif":
		return "RIF"
	case "passport":
		return "Pasaporte"
	default:
		return t
	}
}
