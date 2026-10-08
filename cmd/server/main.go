package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"consulta-resultado-votacao-presidente-2026-oz-gemini/internal/data"
	"consulta-resultado-votacao-presidente-2026-oz-gemini/internal/handler"
)

func main() {
	csvPath := flag.String("csv", "/home/hermes/projetos/referencia-dados-presidente/osasco-presidente.csv", "Caminho para o CSV de votacao")
	dbPath := flag.String("db", filepath.Join(os.TempDir(), fmt.Sprintf("eleicao_oz_%d.db", time.Now().UnixNano())), "Caminho do banco SQLite temporario")
	port := flag.Int("port", 8089, "Porta HTTP do backend Go")
	flag.Parse()

	log.Printf("[INFO] Inicializando banco SQLite temporário em: %s", *dbPath)
	repo, err := data.NewRepository(*dbPath)
	if err != nil {
		log.Fatalf("[FATAL] Erro ao criar banco SQLite: %v", err)
	}
	defer func() {
		_ = repo.Close()
		_ = os.Remove(*dbPath)
		log.Printf("[INFO] Banco SQLite temporário removido.")
	}()

	f, err := os.Open(*csvPath)
	if err != nil {
		log.Fatalf("[FATAL] Erro ao abrir CSV em %s: %v", *csvPath, err)
	}
	defer f.Close()

	log.Printf("[INFO] Importando dados de %s...", *csvPath)
	start := time.Now()
	total, err := repo.ImportCSV(f)
	if err != nil {
		log.Fatalf("[FATAL] Erro ao importar CSV: %v", err)
	}
	log.Printf("[INFO] %d registros importados em %v", total, time.Since(start))

	h := handler.NewHandler(repo)
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	log.Printf("[INFO] Backend Go escutando em http://%s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      h.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] Erro no servidor HTTP: %v", err)
	}
}
