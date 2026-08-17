package main
import (
	"context"
	"fmt"
	"log"
	"github.com/joho/godotenv"
	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/pluggy"
	"github.com/jackc/pgx/v5/pgxpool"
)
func main() {
	godotenv.Load("../.env")
	cfg, _ := config.Load()

	dbPool, _ := pgxpool.New(context.Background(), cfg.DatabaseURL)
	defer dbPool.Close()

	var itemID, accID string
	dbPool.QueryRow(context.Background(), "SELECT pluggy_item_id, pluggy_account_id FROM connected_accounts WHERE user_id = '44f160ec-549c-4a0c-9ebf-bbb11eea18bf'").Scan(&itemID, &accID)

	client := pluggy.NewClient(cfg.PluggyClientID, cfg.PluggyClientSecret)
	
	fmt.Printf("Fetching Item %s...\n", itemID)
	item, err := client.GetItem(itemID)
	if err != nil {
		log.Fatalf("Erro auth: %v", err)
	}
	fmt.Printf("Item status: %s\n", item.Status)

	fmt.Printf("Fetching Accs for %s...\n", itemID)
	accs, err := client.GetAccounts(itemID)
	if err != nil {
		log.Fatalf("Erro acc: %v", err)
	}
	for _, a := range accs {
		fmt.Printf("Acc %s Balance: %v\n", a.ID, a.Balance)
	}
}
