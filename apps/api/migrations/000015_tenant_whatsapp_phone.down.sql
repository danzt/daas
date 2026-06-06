-- S9-T3: rollback — remove WhatsApp notification phone column
ALTER TABLE tenants DROP COLUMN IF EXISTS notifications_whatsapp_phone;
