INSERT INTO users (email, password_hash, name)
VALUES ('owner-{{suffix}}@example.com', 'hash', 'Owner {{suffix}}');
SET @report_owner_id := LAST_INSERT_ID();

INSERT INTO users (email, password_hash, name)
VALUES ('admin-{{suffix}}@example.com', 'hash', 'Admin {{suffix}}');
SET @report_admin_id := LAST_INSERT_ID();

INSERT INTO users (email, password_hash, name)
VALUES ('member-{{suffix}}@example.com', 'hash', 'Member {{suffix}}');
SET @report_member_id := LAST_INSERT_ID();

INSERT INTO users (email, password_hash, name)
VALUES ('outsider-{{suffix}}@example.com', 'hash', 'Outsider {{suffix}}');
SET @report_outsider_id := LAST_INSERT_ID();

INSERT INTO teams (name, created_by)
VALUES ('backend {{suffix}}', @report_owner_id);
SET @report_team_id := LAST_INSERT_ID();

INSERT INTO team_members (team_id, user_id, role)
VALUES
    (@report_team_id, @report_owner_id, 'owner'),
    (@report_team_id, @report_admin_id, 'admin'),
    (@report_team_id, @report_member_id, 'member');

INSERT INTO tasks (title, description, status, assignee_id, team_id, created_by)
VALUES
    ('done {{suffix}}', 'done report task {{suffix}}', 'done', @report_member_id, @report_team_id, @report_owner_id),
    ('invalid {{suffix}}', 'invalid assignee report task {{suffix}}', 'todo', @report_outsider_id, @report_team_id, @report_owner_id),
    ('member-created {{suffix}}', 'member-created report task {{suffix}}', 'todo', @report_member_id, @report_team_id, @report_member_id);
