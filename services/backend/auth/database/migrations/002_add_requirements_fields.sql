-- Migration: Add requirements fields for PulseExpends functional specifications
-- Date: 2026-05-28 03:19:58
-- Description: Adds fields for private expenses, anomaly detection, secure credit cards,
--              exchange rates, mobile notifications, and other requirements

-- ============================================================================
-- TRANSACTIONS TABLE UPDATES
-- ============================================================================

-- Add new columns to transactions table for requirements
ALTER TABLE transactions 
ADD COLUMN IF NOT EXISTS is_private BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS hidden_until TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS private_description TEXT,
ADD COLUMN IF NOT EXISTS source VARCHAR(50) DEFAULT 'manual',
ADD COLUMN IF NOT EXISTS source_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS is_anomaly BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS anomaly_type VARCHAR(50),
ADD COLUMN IF NOT EXISTS anomaly_severity VARCHAR(20),
ADD COLUMN IF NOT EXISTS matched_transaction_id UUID REFERENCES transactions(id),
ADD COLUMN IF NOT EXISTS notification_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS parsed_from_text TEXT,
ADD COLUMN IF NOT EXISTS confidence_score DECIMAL(3,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS base_currency VARCHAR(3),
ADD COLUMN IF NOT EXISTS converted_amount DECIMAL(15,2);

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_transactions_is_private ON transactions(is_private);
CREATE INDEX IF NOT EXISTS idx_transactions_hidden_until ON transactions(hidden_until);
CREATE INDEX IF NOT EXISTS idx_transactions_source ON transactions(source);
CREATE INDEX IF NOT EXISTS idx_transactions_is_anomaly ON transactions(is_anomaly);
CREATE INDEX IF NOT EXISTS idx_transactions_anomaly_severity ON transactions(anomaly_severity);
CREATE INDEX IF NOT EXISTS idx_transactions_matched_transaction_id ON transactions(matched_transaction_id);
CREATE INDEX IF NOT EXISTS idx_transactions_notification_id ON transactions(notification_id);
CREATE INDEX IF NOT EXISTS idx_transactions_base_currency ON transactions(base_currency);

-- ============================================================================
-- NEW TABLES FOR REQUIREMENTS
-- ============================================================================

-- Credit cards table (secure storage - NO PAN/CVV)
CREATE TABLE IF NOT EXISTS credit_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    circle_id UUID REFERENCES circles(id) ON DELETE CASCADE,
    
    -- Secure storage (NO PAN/CVV per requirements)
    bank_name VARCHAR(255) NOT NULL,
    card_type VARCHAR(50) NOT NULL,
    last_four VARCHAR(4) NOT NULL,
    cardholder_name VARCHAR(255),
    
    -- Billing and limits
    credit_limit DECIMAL(15,2),
    current_balance DECIMAL(15,2) DEFAULT 0,
    payment_due_date TIMESTAMP WITH TIME ZONE,
    closing_date TIMESTAMP WITH TIME ZONE,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Composite unique constraint
    UNIQUE(user_id, bank_name, last_four)
);

-- Indexes for credit_cards
CREATE INDEX IF NOT EXISTS idx_credit_cards_user_id ON credit_cards(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_cards_circle_id ON credit_cards(circle_id);
CREATE INDEX IF NOT EXISTS idx_credit_cards_is_active ON credit_cards(is_active);
CREATE INDEX IF NOT EXISTS idx_credit_cards_is_default ON credit_cards(is_default);

-- Exchange rates table for ARS/USD/EUR
CREATE TABLE IF NOT EXISTS exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency VARCHAR(3) NOT NULL,
    target_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(10,6) NOT NULL,
    date DATE NOT NULL,
    source VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint for daily rates
    UNIQUE(base_currency, target_currency, date)
);

-- Indexes for exchange_rates
CREATE INDEX IF NOT EXISTS idx_exchange_rates_base_target ON exchange_rates(base_currency, target_currency);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_date ON exchange_rates(date);

-- Mobile notifications table
CREATE TABLE IF NOT EXISTS mobile_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    
    -- Notification details
    app_package VARCHAR(255) NOT NULL,
    app_name VARCHAR(255),
    title VARCHAR(500),
    text TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Parsed transaction data
    parsed_amount DECIMAL(15,2),
    parsed_currency VARCHAR(3),
    parsed_merchant VARCHAR(255),
    parsed_category VARCHAR(100),
    confidence DECIMAL(3,2) DEFAULT 0,
    
    -- Processing status
    status VARCHAR(50) DEFAULT 'pending',
    processed_at TIMESTAMP WITH TIME ZONE,
    transaction_id UUID REFERENCES transactions(id),
    error TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes for performance
    UNIQUE(user_id, device_id, app_package, timestamp, text)
);

-- Indexes for mobile_notifications
CREATE INDEX IF NOT EXISTS idx_mobile_notifications_user_id ON mobile_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_mobile_notifications_device_id ON mobile_notifications(device_id);
CREATE INDEX IF NOT EXISTS idx_mobile_notifications_app_package ON mobile_notifications(app_package);
CREATE INDEX IF NOT EXISTS idx_mobile_notifications_status ON mobile_notifications(status);
CREATE INDEX IF NOT EXISTS idx_mobile_notifications_timestamp ON mobile_notifications(timestamp);

-- Notification app whitelist
CREATE TABLE IF NOT EXISTS notification_app_whitelist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_package VARCHAR(255) NOT NULL,
    app_name VARCHAR(255) NOT NULL,
    is_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Composite unique constraint
    UNIQUE(user_id, app_package)
);

-- Indexes for notification_app_whitelist
CREATE INDEX IF NOT EXISTS idx_notification_app_whitelist_user_id ON notification_app_whitelist(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_app_whitelist_app_package ON notification_app_whitelist(app_package);
CREATE INDEX IF NOT EXISTS idx_notification_app_whitelist_is_enabled ON notification_app_whitelist(is_enabled);

-- Anomaly detection rules
CREATE TABLE IF NOT EXISTS anomaly_detection_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID REFERENCES circles(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Rule configuration
    rule_type VARCHAR(50) NOT NULL,
    condition TEXT,
    severity VARCHAR(20) DEFAULT 'warning',
    
    -- Thresholds
    amount_threshold DECIMAL(15,2),
    days_threshold INTEGER,
    percentage_diff DECIMAL(5,2),
    
    -- Activation
    is_active BOOLEAN DEFAULT TRUE,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    trigger_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for anomaly_detection_rules
CREATE INDEX IF NOT EXISTS idx_anomaly_detection_rules_circle_id ON anomaly_detection_rules(circle_id);
CREATE INDEX IF NOT EXISTS idx_anomaly_detection_rules_rule_type ON anomaly_detection_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_anomaly_detection_rules_severity ON anomaly_detection_rules(severity);
CREATE INDEX IF NOT EXISTS idx_anomaly_detection_rules_is_active ON anomaly_detection_rules(is_active);

-- Detected anomalies
CREATE TABLE IF NOT EXISTS detected_anomalies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    rule_id UUID REFERENCES anomaly_detection_rules(id) ON DELETE SET NULL,
    circle_id UUID REFERENCES circles(id) ON DELETE CASCADE,
    
    -- Anomaly details
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    description TEXT,
    
    -- Matching details (for duplicates)
    matched_transaction_id UUID REFERENCES transactions(id),
    amount_difference DECIMAL(15,2),
    percentage_diff DECIMAL(5,2),
    
    -- Resolution
    status VARCHAR(50) DEFAULT 'pending',
    resolved_by UUID REFERENCES users(id),
    resolution_note TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for detected_anomalies
CREATE INDEX IF NOT EXISTS idx_detected_anomalies_transaction_id ON detected_anomalies(transaction_id);
CREATE INDEX IF NOT EXISTS idx_detected_anomalies_rule_id ON detected_anomalies(rule_id);
CREATE INDEX IF NOT EXISTS idx_detected_anomalies_circle_id ON detected_anomalies(circle_id);
CREATE INDEX IF NOT EXISTS idx_detected_anomalies_type ON detected_anomalies(type);
CREATE INDEX IF NOT EXISTS idx_detected_anomalies_severity ON detected_anomalies(severity);
CREATE INDEX IF NOT EXISTS idx_detected_anomalies_status ON detected_anomalies(status);
CREATE INDEX IF NOT EXISTS idx_detected_anomalies_created_at ON detected_anomalies(created_at);

-- ============================================================================
-- DATA MIGRATIONS
-- ============================================================================

-- Update existing transactions to have default values for new columns
UPDATE transactions 
SET 
    source = 'manual',
    is_private = false,
    is_anomaly = false,
    confidence_score = 0
WHERE source IS NULL OR is_private IS NULL OR is_anomaly IS NULL OR confidence_score IS NULL;

-- Set base_currency from family currency for existing transactions
UPDATE transactions t
SET base_currency = f.currency
FROM families f
WHERE t.family_id = f.id 
  AND t.base_currency IS NULL;

-- Set converted_amount for existing transactions (assuming USD as base)
UPDATE transactions 
SET converted_amount = amount
WHERE converted_amount IS NULL 
  AND currency = 'USD';

-- For non-USD transactions, we'll need to run exchange rate updates separately
-- This will be handled by the exchange rate service

-- ============================================================================
-- VIEWS FOR REPORTING
-- ============================================================================

-- View for private expenses that are now visible
CREATE OR REPLACE VIEW visible_private_expenses AS
SELECT 
    t.*,
    u.full_name as creator_name,
    CASE 
        WHEN t.is_private = false THEN t.description
        WHEN t.is_private = true AND t.hidden_until IS NULL THEN 'Gasto Privado'
        WHEN t.is_private = true AND t.hidden_until <= CURRENT_TIMESTAMP THEN t.description
        ELSE 'Gasto Privado'
    END as display_description
FROM transactions t
JOIN users u ON t.user_id = u.id
WHERE t.is_private = true;

-- View for anomalies by severity
CREATE OR REPLACE VIEW anomaly_summary AS
SELECT 
    circle_id,
    severity,
    type,
    COUNT(*) as count,
    MIN(created_at) as first_detected,
    MAX(created_at) as last_detected
FROM detected_anomalies
WHERE status = 'pending'
GROUP BY circle_id, severity, type;

-- View for mobile notification processing status
CREATE OR REPLACE VIEW mobile_notification_stats AS
SELECT 
    user_id,
    app_package,
    status,
    COUNT(*) as total,
    AVG(confidence) as avg_confidence,
    MIN(timestamp) as earliest,
    MAX(timestamp) as latest
FROM mobile_notifications
GROUP BY user_id, app_package, status;

-- ============================================================================
-- FUNCTIONS FOR BUSINESS LOGIC
-- ============================================================================

-- Function to check if private expense is visible
CREATE OR REPLACE FUNCTION is_private_expense_visible(
    p_is_private BOOLEAN,
    p_hidden_until TIMESTAMP WITH TIME ZONE,
    p_viewer_id UUID,
    p_creator_id UUID
) RETURNS BOOLEAN AS $$
BEGIN
    IF NOT p_is_private THEN
        RETURN TRUE;
    END IF;
    
    -- Creator can always see their own private expenses
    IF p_viewer_id = p_creator_id THEN
        RETURN TRUE;
    END IF;
    
    -- Others can only see after hidden_until date
    IF p_hidden_until IS NOT NULL AND CURRENT_TIMESTAMP > p_hidden_until THEN
        RETURN TRUE;
    END IF;
    
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- Function to get display description based on privacy
CREATE OR REPLACE FUNCTION get_display_description(
    p_description TEXT,
    p_private_description TEXT,
    p_is_private BOOLEAN,
    p_hidden_until TIMESTAMP WITH TIME ZONE,
    p_viewer_id UUID,
    p_creator_id UUID
) RETURNS TEXT AS $$
BEGIN
    IF is_private_expense_visible(p_is_private, p_hidden_until, p_viewer_id, p_creator_id) THEN
        RETURN p_description;
    ELSE
        RETURN COALESCE(p_private_description, 'Gasto Privado');
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to convert amount to base currency
CREATE OR REPLACE FUNCTION convert_to_base_currency(
    p_amount DECIMAL,
    p_from_currency VARCHAR(3),
    p_to_currency VARCHAR(3),
    p_date DATE DEFAULT CURRENT_DATE
) RETURNS DECIMAL AS $$
DECLARE
    v_rate DECIMAL;
BEGIN
    IF p_from_currency = p_to_currency THEN
        RETURN p_amount;
    END IF;
    
    -- Get the exchange rate for the date
    SELECT rate INTO v_rate
    FROM exchange_rates
    WHERE base_currency = p_from_currency
      AND target_currency = p_to_currency
      AND date = p_date
    ORDER BY created_at DESC
    LIMIT 1;
    
    -- If no rate found for exact date, try to find the most recent rate
    IF v_rate IS NULL THEN
        SELECT rate INTO v_rate
        FROM exchange_rates
        WHERE base_currency = p_from_currency
          AND target_currency = p_to_currency
          AND date <= p_date
        ORDER BY date DESC
        LIMIT 1;
    END IF;
    
    -- If still no rate, return NULL
    IF v_rate IS NULL THEN
        RETURN NULL;
    END IF;
    
    RETURN p_amount * v_rate;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================

-- Log migration completion
INSERT INTO migration_log (migration_name, applied_at, status) 
VALUES ('002_add_requirements_fields', CURRENT_TIMESTAMP, 'completed')
ON CONFLICT (migration_name) DO UPDATE 
SET applied_at = CURRENT_TIMESTAMP, status = 'completed';
