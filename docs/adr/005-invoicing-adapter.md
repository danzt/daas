# ADR-005: Invoicing Adapter Pattern — FiscalGateway Port

Status: Accepted

## Decision

Define a FiscalGateway interface (port) in the domain layer with EmitInvoice, VoidInvoice, and QueryStatus methods; implement a country-specific adapter per fiscal authority (e.g. SENIAT for Venezuela, DGII for Dominican Republic); use a Registry factory that maps country_code to the correct adapter, making adding a new country a matter of implementing the port and registering the adapter.
