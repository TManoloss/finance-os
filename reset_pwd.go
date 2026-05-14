package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Substitua pela sua URL de conexão se for diferente
	connStr := "postgres://finance:finance123@localhost:5432/financedb"
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}
	defer conn.Close(ctx)

	password := "Manoel@123456"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatalf("Erro ao gerar hash: %v", err)
	}

	email := "manoel@example.com"
	tag, err := conn.Exec(ctx, "UPDATE users SET password_hash = $1 WHERE email = $2", string(hashedPassword), email)
	if err != nil {
		log.Fatalf("Erro ao atualizar banco: %v", err)
	}

	if tag.RowsAffected() == 0 {
		fmt.Println("Nenhum usuário encontrado com este email.")
	} else {
		fmt.Printf("Senha de %s atualizada com sucesso para '%s'\n", email, password)
	}
}
