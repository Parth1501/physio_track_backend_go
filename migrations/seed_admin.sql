-- Default user for Dency (password: Dency@1121). Change after first login if needed.
INSERT INTO users (username, password_hash)
VALUES (
  'dency',
  '$2a$10$OwISyimywnKwtDYz8qG39eygIqPJTANXuiRmvZQGs5ci8YUM2KEDG'
)
ON CONFLICT (username) DO NOTHING;
