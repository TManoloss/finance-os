-- backend/internal/db/seed.sql

-- Categorias do sistema (user_id IS NULL)
INSERT INTO categories (name, color, icon) VALUES
('Alimentação', '#FF6B6B', 'restaurant'),
('Transporte', '#4D96FF', 'directions_car'),
('Saúde', '#6BCB77', 'medical_services'),
('Lazer', '#FFD93D', 'celebration'),
('Assinaturas', '#7C6FFF', 'subscriptions'),
('Moradia', '#FF9F45', 'home'),
('Educação', '#A084E8', 'school'),
('Investimentos', '#4ECDC4', 'trending_up'),
('Renda', '#19A7CE', 'payments'),
('Pet', '#FF85B3', 'pets'),
('Emergências', '#FF4949', 'report_problem'),
('Outros', '#8888A0', 'more_horiz')
ON CONFLICT DO NOTHING;

-- Usuário de teste
-- Password: admin123
INSERT INTO users (id, name, email, password_hash)
VALUES ('00000000-0000-4000-a000-000000000001', 'Admin Teste', 'admin@example.com', '$2a$12$IQKiXSr1ncPI9ZDiJ6jMHuNbzISlfxo29kiyj3s9OsB5DTzsbEZGC')
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;
