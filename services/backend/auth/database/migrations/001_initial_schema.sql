-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE,
    full_name VARCHAR(255),
    avatar_url TEXT,
    provider VARCHAR(50) DEFAULT 'local',
    provider_id VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    preferences JSONB DEFAULT '{
        "default_currency": "USD",
        "language": "es",
        "time_zone": "UTC",
        "date_format": "YYYY-MM-DD",
        "currency_format": "$0,0.00",
        "theme": "light",
        "notification_settings": {
            "email_frequency": "weekly",
            "push_enabled": true,
            "desktop_alerts": true,
            "sound_enabled": true
        }
    }',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_provider ON users(provider, provider_id);
CREATE INDEX idx_users_is_active ON users(is_active);

-- User authentication table
CREATE TABLE user_auth (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255),
    salt VARCHAR(255),
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255),
    last_password_change TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    failed_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Indexes for user_auth
CREATE INDEX idx_user_auth_user_id ON user_auth(user_id);

-- User sessions table
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    device_info TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for user_sessions
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(token);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);

-- Circles table
CREATE TABLE circles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    currency VARCHAR(3) DEFAULT 'USD',
    created_by UUID NOT NULL REFERENCES users(id),
    is_active BOOLEAN DEFAULT TRUE,
    is_public BOOLEAN DEFAULT FALSE,
    join_code VARCHAR(20) UNIQUE,
    settings JSONB DEFAULT '{
        "allow_member_add_transactions": true,
        "allow_member_edit_transactions": true,
        "allow_member_delete_transactions": false,
        "require_approval_for_add": false,
        "require_approval_for_edit": false,
        "default_categories": [
            "Food & Dining",
            "Transportation",
            "Shopping",
            "Entertainment",
            "Utilities",
            "Housing",
            "Healthcare",
            "Education",
            "Travel",
            "Other"
        ],
        "budget_alerts_enabled": true,
        "monthly_budget": 0,
        "notification_preferences": {
            "email_notifications": true,
            "push_notifications": true,
            "weekly_summary": true,
            "budget_alerts": true,
            "large_transaction_alerts": true,
            "new_member_alerts": true,
            "transaction_approval_alerts": true
        }
    }',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB
);

-- Indexes for circles
CREATE INDEX idx_circles_created_by ON circles(created_by);
CREATE INDEX idx_circles_is_active ON circles(is_active);
CREATE INDEX idx_circles_is_public ON circles(is_public);
CREATE INDEX idx_circles_join_code ON circles(join_code);

-- Circle members table
CREATE TABLE circle_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(circle_id, user_id)
);

-- Indexes for circle_members
CREATE INDEX idx_circle_members_circle_id ON circle_members(circle_id);
CREATE INDEX idx_circle_members_user_id ON circle_members(user_id);
CREATE INDEX idx_circle_members_role ON circle_members(role);
CREATE INDEX idx_circle_members_is_active ON circle_members(is_active);

-- Transactions table
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    circle_id UUID REFERENCES circles(id) ON DELETE SET NULL,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    category VARCHAR(100) NOT NULL,
    description TEXT,
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense', 'transfer')),
    payment_method VARCHAR(50),
    location TEXT,
    tags JSONB DEFAULT '[]',
    receipt_url TEXT,
    is_recurring BOOLEAN DEFAULT FALSE,
    recurring_id UUID,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for transactions
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_circle_id ON transactions(circle_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_category ON transactions(category);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_recurring_id ON transactions(recurring_id);

-- Recurring transactions table
CREATE TABLE recurring_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    circle_id UUID REFERENCES circles(id) ON DELETE SET NULL,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    category VARCHAR(100) NOT NULL,
    description TEXT,
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
    day_of_month INTEGER CHECK (day_of_month >= 1 AND day_of_month <= 31),
    day_of_week VARCHAR(10),
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    last_run TIMESTAMP WITH TIME ZONE,
    next_run TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for recurring_transactions
CREATE INDEX idx_recurring_transactions_user_id ON recurring_transactions(user_id);
CREATE INDEX idx_recurring_transactions_circle_id ON recurring_transactions(circle_id);
CREATE INDEX idx_recurring_transactions_next_run ON recurring_transactions(next_run);
CREATE INDEX idx_recurring_transactions_is_active ON recurring_transactions(is_active);

-- Attachments table
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    file_type VARCHAR(100),
    file_size BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for attachments
CREATE INDEX idx_attachments_transaction_id ON attachments(transaction_id);
CREATE INDEX idx_attachments_user_id ON attachments(user_id);

-- Comments table
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_edited BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for comments
CREATE INDEX idx_comments_transaction_id ON comments(transaction_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);

-- Circle invites table
CREATE TABLE circle_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',
    token VARCHAR(100) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for circle_invites
CREATE INDEX idx_circle_invites_circle_id ON circle_invites(circle_id);
CREATE INDEX idx_circle_invites_email ON circle_invites(email);
CREATE INDEX idx_circle_invites_token ON circle_invites(token);
CREATE INDEX idx_circle_invites_expires_at ON circle_invites(expires_at);

-- Circle activities table
CREATE TABLE circle_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity VARCHAR(100) NOT NULL,
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for circle_activities
CREATE INDEX idx_circle_activities_circle_id ON circle_activities(circle_id);
CREATE INDEX idx_circle_activities_user_id ON circle_activities(user_id);
CREATE INDEX idx_circle_activities_created_at ON circle_activities(created_at);

-- Split transactions table
CREATE TABLE split_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(15, 2) NOT NULL,
    percentage DECIMAL(5, 2) NOT NULL,
    is_paid BOOLEAN DEFAULT FALSE,
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for split_transactions
CREATE INDEX idx_split_transactions_transaction_id ON split_transactions(transaction_id);
CREATE INDEX idx_split_transactions_user_id ON split_transactions(user_id);
CREATE INDEX idx_split_transactions_is_paid ON split_transactions(is_paid);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_auth_updated_at BEFORE UPDATE ON user_auth
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_circles_updated_at BEFORE UPDATE ON circles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_recurring_transactions_updated_at BEFORE UPDATE ON recurring_transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_circle_invites_updated_at BEFORE UPDATE ON circle_invites
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_split_transactions_updated_at BEFORE UPDATE ON split_transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to generate join code
CREATE OR REPLACE FUNCTION generate_join_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.join_code IS NULL THEN
        NEW.join_code = upper(substring(md5(random()::text) from 1 for 8));
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for join code generation
CREATE TRIGGER generate_circle_join_code BEFORE INSERT ON circles
    FOR EACH ROW EXECUTE FUNCTION generate_join_code();

-- Create function to log circle activities
CREATE OR REPLACE FUNCTION log_circle_activity()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND TG_TABLE_NAME = 'circle_members' THEN
        INSERT INTO circle_activities (circle_id, user_id, activity, details)
        VALUES (NEW.circle_id, NEW.user_id, 'member_joined', 'User joined the circle');
    ELSIF TG_OP = 'UPDATE' AND TG_TABLE_NAME = 'circle_members' AND NEW.is_active = FALSE AND OLD.is_active = TRUE THEN
        INSERT INTO circle_activities (circle_id, user_id, activity, details)
        VALUES (NEW.circle_id, NEW.user_id, 'member_left', 'User left the circle');
    ELSIF TG_OP = 'INSERT' AND TG_TABLE_NAME = 'transactions' AND NEW.circle_id IS NOT NULL THEN
        INSERT INTO circle_activities (circle_id, user_id, activity, details)
        VALUES (NEW.circle_id, NEW.user_id, 'transaction_added', 
                CONCAT('Added ', NEW.type, ' of ', NEW.amount, ' ', NEW.currency, ' - ', NEW.description));
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for activity logging
CREATE TRIGGER log_circle_member_activity AFTER INSERT OR UPDATE ON circle_members
    FOR EACH ROW EXECUTE FUNCTION log_circle_activity();

CREATE TRIGGER log_circle_transaction_activity AFTER INSERT ON transactions
    FOR EACH ROW EXECUTE FUNCTION log_circle_activity();

-- Create view for circle summaries
CREATE VIEW circle_summaries AS
SELECT 
    c.id as circle_id,
    COUNT(DISTINCT cm.id) as total_members,
    COUNT(DISTINCT CASE WHEN cm.is_active THEN cm.id END) as active_members,
    COUNT(DISTINCT t.id) as total_transactions,
    COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0) as total_amount,
    COALESCE(c.settings->>'monthly_budget', '0')::DECIMAL as monthly_budget,
    COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0) as budget_used,
    (COALESCE(c.settings->>'monthly_budget', '0')::DECIMAL - COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0)) as budget_remaining,
    MAX(t.date) as last_transaction_at,
    c.created_at
FROM circles c
LEFT JOIN circle_members cm ON c.id = cm.circle_id
LEFT JOIN transactions t ON c.id = t.circle_id AND t.deleted_at IS NULL
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.created_at;

-- Create view for user transaction summaries
CREATE VIEW user_transaction_summaries AS
SELECT 
    u.id as user_id,
    COUNT(DISTINCT t.id) as total_transactions,
    COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END), 0) as total_income,
    COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0) as total_expenses,
    COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END), 0) - 
    COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE t.amount END), 0) as net_balance,
    MAX(t.date) as last_transaction_date
FROM users u
LEFT JOIN transactions t ON u.id = t.user_id AND t.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id;