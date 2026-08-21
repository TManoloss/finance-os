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

-- Subcategorias usadas para distinguir canais dentro de cada categoria principal.
INSERT INTO categories (user_id, name, color, icon, parent_id)
SELECT NULL, child.name, parent.color, child.icon, parent.id
FROM (VALUES
    ('Delivery', 'restaurant'),
    ('Restaurante', 'restaurant'),
    ('Mercado', 'shopping_cart'),
    ('Padaria', 'bakery_dining'),
    ('Loja de conveniência', 'local_convenience_store')
) AS child(name, icon)
JOIN categories parent ON parent.user_id IS NULL AND parent.name = 'Alimentação'
WHERE NOT EXISTS (
    SELECT 1 FROM categories existing
    WHERE existing.user_id IS NULL AND existing.name = child.name AND existing.parent_id = parent.id
);

INSERT INTO categories (user_id, name, color, icon, parent_id)
SELECT NULL, child.name, parent.color, child.icon, parent.id
FROM (VALUES
    ('Transporte por aplicativo', 'local_taxi'),
    ('Combustível', 'local_gas_station'),
    ('Transporte público', 'directions_bus')
) AS child(name, icon)
JOIN categories parent ON parent.user_id IS NULL AND parent.name = 'Transporte'
WHERE NOT EXISTS (
    SELECT 1 FROM categories existing
    WHERE existing.user_id IS NULL AND existing.name = child.name AND existing.parent_id = parent.id
);

-- Usuário de teste
-- Password: admin123
INSERT INTO users (id, name, email, password_hash)
VALUES ('00000000-0000-4000-a000-000000000001', 'Admin Teste', 'admin@example.com', '$2a$12$IQKiXSr1ncPI9ZDiJ6jMHuNbzISlfxo29kiyj3s9OsB5DTzsbEZGC')
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;
