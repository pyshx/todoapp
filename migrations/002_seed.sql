-- 002_seed.sql
-- Seed data for testing

-- Companies
INSERT INTO companies (id, name) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Acme Corp'),
    ('22222222-2222-2222-2222-222222222222', 'Beta Inc');

-- Users
INSERT INTO users (id, company_id, email, role) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 'alice@acme.com', 'editor'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', 'bob@acme.com', 'viewer'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222222', 'charlie@beta.com', 'editor');

-- Sample tasks
INSERT INTO tasks (id, company_id, creator_id, assignee_id, title, description, visibility, status) VALUES
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Complete project setup', 'Set up the initial project structure', 'company_wide', 'done'),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Review documentation', 'Review the API documentation', 'company_wide', 'todo'),
    ('ffffffff-ffff-ffff-ffff-ffffffffffff', '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL, 'Private task', 'This is a private task', 'only_me', 'in_progress'),
    ('00000000-0000-0000-0000-000000000001', '22222222-2222-2222-2222-222222222222', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'Beta task', 'A task for Beta Inc', 'company_wide', 'todo');
