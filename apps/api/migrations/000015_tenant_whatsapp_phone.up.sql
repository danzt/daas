-- S9-T3: WhatsApp notification phone number for tenant owners
ALTER TABLE tenants
    ADD COLUMN IF NOT EXISTS notifications_whatsapp_phone VARCHAR(30);
