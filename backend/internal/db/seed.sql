-- backend/internal/db/seed.sql

-- Categorias do sistema (user_id IS NULL)
INSERT INTO categories (name, color, icon) VALUES
('Alimentação', '#FF6B6B', 'restaurant'),
('Transporte', '#4D96FF', 'directions_car'),
('Saúde', '#6BCB77', 'medical_services'),
('Lazer', '#FFD93D', 'celebration'),
('Assinaturas', '#7C6FFF', 'subscriptions'),
('Contas', '#64748B', 'receipt_long'),
('Compras', '#F97316', 'shopping_bag'),
('Financeiro', '#14B8A6', 'account_balance'),
('Moradia', '#FF9F45', 'home'),
('Educação', '#A084E8', 'school'),
('Investimentos', '#4ECDC4', 'trending_up'),
('Renda', '#19A7CE', 'payments'),
('Pet', '#FF85B3', 'pets'),
('Transferência', '#94A3B8', 'swap_horiz'),
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

INSERT INTO categories (user_id, name, color, icon, parent_id)
SELECT NULL, child.name, parent.color, child.icon, parent.id
FROM (VALUES
    ('Moradia', 'home', 'Contas'), ('Energia', 'bolt', 'Contas'), ('Água', 'water_drop', 'Contas'),
    ('Internet e telefonia', 'wifi', 'Contas'), ('Seguros', 'shield', 'Contas'), ('Impostos e taxas', 'receipt', 'Contas'),
    ('Investimentos', 'trending_up', 'Financeiro'), ('Tarifas e juros', 'percent', 'Financeiro'), ('Impostos', 'account_balance_wallet', 'Financeiro'),
    ('Vestuário', 'checkroom', 'Compras'), ('Casa', 'chair', 'Compras'), ('Eletrônicos', 'devices', 'Compras'), ('Marketplace', 'storefront', 'Compras'),
    ('Farmácia', 'medication', 'Saúde'), ('Consultas', 'medical_information', 'Saúde'), ('Exames', 'science', 'Saúde'), ('Plano de saúde', 'health_and_safety', 'Saúde'),
    ('Cinema e eventos', 'local_activity', 'Lazer'), ('Viagens', 'flight', 'Lazer'), ('Hobbies', 'sports_esports', 'Lazer'), ('Jogos', 'sports_esports', 'Lazer'),
    ('Streaming', 'play_circle', 'Assinaturas'), ('Software', 'apps', 'Assinaturas'), ('Nuvem', 'cloud', 'Assinaturas'),
    ('Cursos', 'school', 'Educação'), ('Escola e faculdade', 'school', 'Educação'), ('Livros', 'menu_book', 'Educação'),
    ('Salário', 'payments', 'Renda'), ('Freelance', 'work', 'Renda'), ('Vendas', 'sell', 'Renda'), ('Rendimentos', 'savings', 'Renda'),
    ('Ração', 'pets', 'Pet'), ('Veterinário', 'medical_services', 'Pet'), ('Higiene', 'cleaning_services', 'Pet'),
    ('Multas', 'gavel', 'Emergências'), ('Reparos', 'build', 'Emergências'), ('Imprevistos', 'warning', 'Emergências'),
    ('Entre contas', 'swap_horiz', 'Transferência'), ('Pix pessoal', 'payments', 'Transferência'), ('Pagamento de fatura', 'credit_card', 'Transferência')
) AS child(name, icon, parent_name)
JOIN categories parent ON parent.user_id IS NULL AND parent.name = child.parent_name
WHERE NOT EXISTS (SELECT 1 FROM categories existing WHERE existing.user_id IS NULL AND existing.name = child.name AND existing.parent_id = parent.id);

UPDATE categories child SET parent_id = parent.id
FROM categories parent
WHERE child.user_id IS NULL AND parent.user_id IS NULL
  AND ((child.name = 'Moradia' AND parent.name = 'Contas')
    OR (child.name = 'Investimentos' AND parent.name = 'Financeiro'))
  AND child.id <> parent.id;

-- Usuário de teste
-- Password: admin123
INSERT INTO users (id, name, email, password_hash)
VALUES ('00000000-0000-4000-a000-000000000001', 'Admin Teste', 'admin@example.com', '$2a$12$IQKiXSr1ncPI9ZDiJ6jMHuNbzISlfxo29kiyj3s9OsB5DTzsbEZGC')
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;
