# ADR-006: Mixed Sale Splitting — Two Separate Documents

Status: Accepted

## Decision

When a sale contains both fiscal and non-fiscal products, emit two separate documents: a fiscal invoice (via FiscalGateway) for fiscal products and an internal invoice (via InternalInvoiceService) for non-fiscal products, both linked to the same sale_id as a foreign key. This keeps each document autonomous and auditably correct for its own regime, rather than producing one document with filtered lines.
