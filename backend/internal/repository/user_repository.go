package repository

import (
	"context"
	"time"

	"github.com/finance-os/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository define a interface para persistência de usuários.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	UpdatePluggyCredentials(ctx context.Context, userID, clientID, encryptedSecret string) error
	SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error
	ValidateRefreshToken(ctx context.Context, token string) (string, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

// pgUserRepository implementa a interface UserRepository usando pgxpool.
type pgUserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository cria uma nova instância de UserRepository.
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &pgUserRepository{db: db}
}

// Create insere um novo usuário no banco de dados.
func (r *pgUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, name, email, password_hash, created_at) 
	          VALUES (gen_random_uuid(), $1, $2, $3, NOW()) 
	          RETURNING id, created_at`
	
	err := r.db.QueryRow(ctx, query, user.Name, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
	return err
}

// FindByEmail busca um usuário pelo email (case-insensitive).
func (r *pgUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, COALESCE(pluggy_client_id, ''), COALESCE(pluggy_client_secret_encrypted, ''), created_at 
	          FROM users WHERE LOWER(email) = LOWER($1)`
	
	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.PluggyClientID, &user.PluggyClientSecretEncrypted, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID busca um usuário pelo ID.
func (r *pgUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, COALESCE(pluggy_client_id, ''), COALESCE(pluggy_client_secret_encrypted, ''), created_at 
	          FROM users WHERE id = $1`
	
	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.PluggyClientID, &user.PluggyClientSecretEncrypted, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePluggyCredentials atualiza as credenciais da Pluggy para um usuário.
func (r *pgUserRepository) UpdatePluggyCredentials(ctx context.Context, userID, clientID, encryptedSecret string) error {
	query := `UPDATE users SET pluggy_client_id = $1, pluggy_client_secret_encrypted = $2 WHERE id = $3`
	
	_, err := r.db.Exec(ctx, query, clientID, encryptedSecret, userID)
	return err
}

// SaveRefreshToken salva um novo refresh token no banco de dados.
func (r *pgUserRepository) SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at) 
	          VALUES (gen_random_uuid(), $1, $2, $3, NOW())`
	
	_, err := r.db.Exec(ctx, query, userID, token, expiresAt)
	return err
}

// ValidateRefreshToken verifica se um token é válido e retorna o user_id.
func (r *pgUserRepository) ValidateRefreshToken(ctx context.Context, token string) (string, error) {
	query := `SELECT user_id FROM refresh_tokens WHERE token = $1 AND expires_at > NOW()`
	
	var userID string
	err := r.db.QueryRow(ctx, query, token).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}

// DeleteRefreshToken remove um refresh token do banco de dados (logout ou rotação).
func (r *pgUserRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	
	_, err := r.db.Exec(ctx, query, token)
	return err
}
