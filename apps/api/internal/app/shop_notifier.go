package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/danzt/daas/api/internal/domain/notification"
	"github.com/danzt/daas/api/internal/domain/shop"
)

// ShopOrderNotifier dispatches transactional emails to the customer when
// the order changes state. It deliberately runs async (best-effort) — a
// notification failure must NEVER block the admin response or roll back
// the state transition.
type ShopOrderNotifier struct {
	svc           notification.NotificationService
	pool          *pgxpool.Pool
	storefrontURL string // e.g. https://daas.app or http://localhost:3000
}

// NewShopOrderNotifier wires a notifier. If svc is nil, all sends are no-ops.
// storefrontURL is the public base of the storefront (used to build customer
// tracking links). Falls back to http://localhost:3000 when empty.
func NewShopOrderNotifier(svc notification.NotificationService, pool *pgxpool.Pool, storefrontURL string) *ShopOrderNotifier {
	if storefrontURL == "" {
		storefrontURL = "http://localhost:3000"
	}
	return &ShopOrderNotifier{svc: svc, pool: pool, storefrontURL: strings.TrimRight(storefrontURL, "/")}
}

// Notify enqueues a transactional email for the given lifecycle event.
// It returns immediately; the send happens in a background goroutine.
func (n *ShopOrderNotifier) Notify(parent context.Context, tenantID uuid.UUID, order *shop.ShopOrder) {
	if n == nil || n.svc == nil || order == nil || order.CustomerEmail == "" {
		return
	}
	// Snapshot what we need so the goroutine doesn't capture the request ctx.
	go n.send(tenantID, *order)
}

func (n *ShopOrderNotifier) send(tenantID uuid.UUID, order shop.ShopOrder) {
	// Use a fresh context with a hard deadline so a slow SMTP doesn't hang
	// the worker forever.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tenantName, slug, err := n.lookupTenant(ctx, tenantID)
	if err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID.String()).Msg("notifier: tenant lookup failed; skipping email")
		return
	}

	subject, body := n.renderTemplate(order, tenantName, slug)
	if subject == "" {
		// No template for this status (e.g. pending) — skip silently.
		return
	}

	msg := notification.Message{
		Channel: "email",
		To:      order.CustomerEmail,
		Subject: subject,
		Body:    body,
		Data: map[string]any{
			"order_id":  order.ID.String(),
			"tenant_id": tenantID.String(),
			"status":    string(order.Status),
		},
	}
	if err := n.svc.Send(ctx, msg); err != nil {
		log.Warn().
			Err(err).
			Str("order_id", order.ID.String()).
			Str("to", order.CustomerEmail).
			Msg("notifier: failed to send order email")
		return
	}
	log.Info().
		Str("order_id", order.ID.String()).
		Str("to", order.CustomerEmail).
		Str("status", string(order.Status)).
		Msg("notifier: order email sent")
}

// NotifyProofUploaded sends an email to the tenant owner when a customer
// uploads a payment proof. This is the inverse notification to Notify —
// it goes to the seller, not the buyer.
// Like Notify, it runs async and never blocks the HTTP response.
func (n *ShopOrderNotifier) NotifyProofUploaded(parent context.Context, tenantID uuid.UUID, order *shop.ShopOrder) {
	if n == nil || n.svc == nil || order == nil {
		return
	}
	go n.sendProofAlert(tenantID, *order)
}

func (n *ShopOrderNotifier) sendProofAlert(tenantID uuid.UUID, order shop.ShopOrder) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tenantName, slug, err := n.lookupTenant(ctx, tenantID)
	if err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID.String()).Msg("notifier: tenant lookup failed for proof alert")
		return
	}

	ownerEmail, err := n.lookupOwnerEmail(ctx, tenantID)
	if err != nil || ownerEmail == "" {
		log.Warn().Err(err).Str("tenant_id", tenantID.String()).Msg("notifier: owner email not found; skipping proof alert")
		return
	}

	shortID := strings.ToUpper(order.ID.String()[:8])
	adminURL := fmt.Sprintf("%s/shop-orders/%s", n.storefrontURL, order.ID.String())

	subject := fmt.Sprintf("💰 Nuevo comprobante de pago · Pedido #%s", shortID)
	body := renderProofAlertEmail(proofAlertParams{
		TenantName:   tenantName,
		CustomerName: order.CustomerName,
		OrderShort:   shortID,
		OrderTotal:   formatMoney(order.Total),
		AdminURL:     adminURL,
		Slug:         slug,
	})

	msg := notification.Message{
		Channel: "email",
		To:      ownerEmail,
		Subject: subject,
		Body:    body,
		Data: map[string]any{
			"order_id":  order.ID.String(),
			"tenant_id": tenantID.String(),
			"event":     "proof_uploaded",
		},
	}
	if err := n.svc.Send(ctx, msg); err != nil {
		log.Warn().Err(err).Str("order_id", order.ID.String()).Msg("notifier: failed to send proof alert")
		return
	}
	log.Info().Str("order_id", order.ID.String()).Str("to", ownerEmail).Msg("notifier: proof alert sent to tenant")
}

func (n *ShopOrderNotifier) lookupOwnerEmail(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var email string
	err := n.pool.QueryRow(ctx,
		`SELECT email FROM tenant_users WHERE tenant_id = $1 AND role = 'owner' LIMIT 1`,
		tenantID,
	).Scan(&email)
	if err != nil {
		return "", fmt.Errorf("owner email: %w", err)
	}
	return email, nil
}

type proofAlertParams struct {
	TenantName   string
	CustomerName string
	OrderShort   string
	OrderTotal   string
	AdminURL     string
	Slug         string
}

func renderProofAlertEmail(p proofAlertParams) string {
	return fmt.Sprintf(`<!doctype html>
<html><body style="margin:0;padding:0;background:#f8fafc;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
  <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%" style="background:#f8fafc;padding:32px 16px">
    <tr><td align="center">
      <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="560" style="max-width:560px;background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,0.04)">
        <tr><td style="padding:32px 32px 16px">
          <p style="margin:0;font-size:12px;color:#94a3b8;letter-spacing:0.5px;text-transform:uppercase">%s · DaaS</p>
          <h1 style="margin:8px 0 16px;font-size:24px;color:#0f172a">💰 Nuevo comprobante recibido</h1>
          <p style="margin:0;color:#334155;font-size:15px;line-height:1.5">
            El cliente <strong>%s</strong> subió un comprobante de pago para el pedido <strong>#%s</strong>.
          </p>
        </td></tr>
        <tr><td style="padding:16px 32px">
          <table role="presentation" width="100%%" style="background:#f8fafc;border-radius:8px;padding:16px">
            <tr>
              <td style="font-size:13px;color:#64748b">Pedido</td>
              <td align="right" style="font-size:13px;color:#0f172a;font-family:ui-monospace,Menlo,monospace;font-weight:600">#%s</td>
            </tr>
            <tr>
              <td style="font-size:13px;color:#64748b;padding-top:8px">Total</td>
              <td align="right" style="font-size:16px;color:#0f172a;font-family:ui-monospace,Menlo,monospace;font-weight:700;padding-top:8px">Bs.S %s</td>
            </tr>
          </table>
        </td></tr>
        <tr><td align="center" style="padding:8px 32px 32px">
          <a href="%s" style="display:inline-block;background:#7c3aed;color:#ffffff;text-decoration:none;padding:12px 24px;border-radius:8px;font-size:14px;font-weight:600">Ver pedido y confirmar pago</a>
          <p style="color:#475569;font-size:14px;line-height:1.5;margin:24px 0 0">
            Revisá el comprobante y hacé click en <strong>"Confirmar pago"</strong> para procesar el pedido.
          </p>
        </td></tr>
        <tr><td style="padding:0 32px 24px;border-top:1px solid #e2e8f0;padding-top:16px">
          <p style="margin:0;color:#94a3b8;font-size:12px;text-align:center">Este email fue enviado por DaaS. Respondé directamente si tenés preguntas.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body></html>`,
		htmlEscape(p.TenantName),
		htmlEscape(p.CustomerName),
		htmlEscape(p.OrderShort),
		htmlEscape(p.OrderShort),
		htmlEscape(p.OrderTotal),
		htmlEscape(p.AdminURL),
	)
}

func (n *ShopOrderNotifier) lookupTenant(ctx context.Context, tenantID uuid.UUID) (name, slug string, err error) {
	err = n.pool.QueryRow(ctx,
		`SELECT name, slug FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&name, &slug)
	return
}

// renderTemplate returns (subject, htmlBody) for the given order state.
// Returns ("", "") for states we don't notify about.
func (n *ShopOrderNotifier) renderTemplate(order shop.ShopOrder, tenantName, slug string) (string, string) {
	trackingURL := fmt.Sprintf("%s/t/%s/order/%s?access_token=%s",
		n.storefrontURL, slug, order.ID.String(), order.AccessToken)
	shortID := strings.ToUpper(order.ID.String()[:8])

	switch order.Status {
	case shop.OrderStatusPaid:
		return fmt.Sprintf("Tu pago fue confirmado · Pedido #%s", shortID),
			renderEmail(emailParams{
				TenantName:  tenantName,
				Headline:    "¡Pago confirmado!",
				Lead:        fmt.Sprintf("Hola %s, recibimos tu pago. Estamos preparando tu pedido para enviarlo.", order.CustomerName),
				StatusLabel: "Pagado",
				StatusColor: "#3b82f6", // blue
				OrderShort:  shortID,
				OrderTotal:  formatMoney(order.Total),
				TrackingURL: trackingURL,
				NextStep:    "Te avisamos en cuanto despachemos tu pedido.",
			})
	case shop.OrderStatusFulfilled:
		return fmt.Sprintf("Tu pedido está en camino · #%s", shortID),
			renderEmail(emailParams{
				TenantName:  tenantName,
				Headline:    "Tu pedido va en camino 🚚",
				Lead:        fmt.Sprintf("Hola %s, despachamos tu pedido. Pronto lo vas a recibir.", order.CustomerName),
				StatusLabel: "Despachado",
				StatusColor: "#a855f7", // purple
				OrderShort:  shortID,
				OrderTotal:  formatMoney(order.Total),
				TrackingURL: trackingURL,
				NextStep:    "Te avisamos en cuanto lo confirmen como entregado.",
			})
	case shop.OrderStatusDelivered:
		return fmt.Sprintf("¡Tu pedido fue entregado! · #%s", shortID),
			renderEmail(emailParams{
				TenantName:  tenantName,
				Headline:    "¡Tu pedido llegó! 🎉",
				Lead:        fmt.Sprintf("Hola %s, esperamos que disfrutes tu compra. Gracias por elegirnos.", order.CustomerName),
				StatusLabel: "Entregado",
				StatusColor: "#10b981", // emerald
				OrderShort:  shortID,
				OrderTotal:  formatMoney(order.Total),
				TrackingURL: trackingURL,
				NextStep:    "Si necesitás ayuda con tu compra, respondé este email.",
			})
	case shop.OrderStatusCancelled:
		return fmt.Sprintf("Tu pedido fue cancelado · #%s", shortID),
			renderEmail(emailParams{
				TenantName:  tenantName,
				Headline:    "Pedido cancelado",
				Lead:        fmt.Sprintf("Hola %s, tu pedido fue cancelado. Si esto fue un error, contactanos respondiendo este email.", order.CustomerName),
				StatusLabel: "Cancelado",
				StatusColor: "#f43f5e", // rose
				OrderShort:  shortID,
				OrderTotal:  formatMoney(order.Total),
				TrackingURL: trackingURL,
				NextStep:    "",
			})
	default:
		return "", ""
	}
}

type emailParams struct {
	TenantName  string
	Headline    string
	Lead        string
	StatusLabel string
	StatusColor string
	OrderShort  string
	OrderTotal  string
	TrackingURL string
	NextStep    string
}

// renderEmail returns a minimal, mobile-friendly HTML email.
// Kept inline to avoid a templating dep in this PR.
func renderEmail(p emailParams) string {
	nextStepBlock := ""
	if p.NextStep != "" {
		nextStepBlock = fmt.Sprintf(`<p style="color:#475569;font-size:14px;line-height:1.5;margin:24px 0 0">%s</p>`, htmlEscape(p.NextStep))
	}
	return fmt.Sprintf(`<!doctype html>
<html><body style="margin:0;padding:0;background:#f8fafc;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
  <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%" style="background:#f8fafc;padding:32px 16px">
    <tr><td align="center">
      <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="560" style="max-width:560px;background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,0.04)">
        <tr><td style="padding:32px 32px 16px">
          <p style="margin:0;font-size:12px;color:#94a3b8;letter-spacing:0.5px;text-transform:uppercase">%s</p>
          <h1 style="margin:8px 0 16px;font-size:24px;color:#0f172a">%s</h1>
          <p style="margin:0;color:#334155;font-size:15px;line-height:1.5">%s</p>
        </td></tr>
        <tr><td style="padding:0 32px 16px">
          <div style="display:inline-block;padding:4px 12px;border-radius:999px;background:%s20;color:%s;font-size:12px;font-weight:600">%s</div>
        </td></tr>
        <tr><td style="padding:16px 32px">
          <table role="presentation" width="100%%" style="background:#f8fafc;border-radius:8px;padding:16px">
            <tr>
              <td style="font-size:13px;color:#64748b">Pedido</td>
              <td align="right" style="font-size:13px;color:#0f172a;font-family:ui-monospace,Menlo,monospace;font-weight:600">#%s</td>
            </tr>
            <tr>
              <td style="font-size:13px;color:#64748b;padding-top:8px">Total</td>
              <td align="right" style="font-size:16px;color:#0f172a;font-family:ui-monospace,Menlo,monospace;font-weight:700;padding-top:8px">Bs.S %s</td>
            </tr>
          </table>
        </td></tr>
        <tr><td align="center" style="padding:8px 32px 32px">
          <a href="%s" style="display:inline-block;background:#7c3aed;color:#ffffff;text-decoration:none;padding:12px 24px;border-radius:8px;font-size:14px;font-weight:600">Ver mi pedido</a>
          %s
        </td></tr>
        <tr><td style="padding:0 32px 24px;border-top:1px solid #e2e8f0;padding-top:16px">
          <p style="margin:0;color:#94a3b8;font-size:12px;text-align:center">Este email fue enviado por DaaS. Si tenés dudas, respondé directamente.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body></html>`,
		htmlEscape(p.TenantName),
		htmlEscape(p.Headline),
		htmlEscape(p.Lead),
		p.StatusColor, p.StatusColor, htmlEscape(p.StatusLabel),
		htmlEscape(p.OrderShort),
		htmlEscape(p.OrderTotal),
		htmlEscape(p.TrackingURL),
		nextStepBlock,
	)
}

func formatMoney(n float64) string {
	return fmt.Sprintf("%.2f", n)
}

// htmlEscape is a tiny escaper for the small set of values we interpolate.
// Customer-controlled fields (name, address) go through this before landing
// in the HTML body.
func htmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}
