package main
import (
	"fmt"
	"log"
	"github.com/joho/godotenv"
	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/pluggy"
)
func main() {
	godotenv.Load("../.env")
	cfg, _ := config.Load()
	fmt.Printf("ID: %s\n", cfg.PluggyClientID)
	client := pluggy.NewClient(cfg.PluggyClientID, cfg.PluggyClientSecret)
	
	// test connection
	item, err := client.GetItem("7e023d56-c989-46db-a00f-552efe27a583") // This is Manoel's item
	if err != nil {
		log.Fatalf("Erro auth: %v", err)
	}
	fmt.Printf("Item status: %s\n", item.Status)
}
