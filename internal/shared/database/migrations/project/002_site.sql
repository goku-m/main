CREATE EXTENSION IF NOT EXISTS pgcrypto;



CREATE TABLE project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    user_id TEXT  NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    priority TEXT NOT NULL DEFAULT 'medium',
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
       sort_order SERIAL
);

CREATE INDEX idx_project_user_id ON project(user_id);
CREATE INDEX idx_project_status ON project(status);
CREATE INDEX idx_project_priority ON project(priority);
CREATE INDEX idx_project_due_date ON project(due_date);

CREATE TRIGGER set_updated_at_project
    BEFORE UPDATE ON project
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();





-- Composite index for user project with status and priority
CREATE INDEX idx_project_user_status_priority ON project(user_id, status, priority);
