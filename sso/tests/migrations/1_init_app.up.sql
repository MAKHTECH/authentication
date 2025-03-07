INSERT INTO apps (id, name, secret)
VALUES (1, 'tests', 'test-secret')
ON CONFLICT DO NOTHING;