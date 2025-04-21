package main

import (
	"EncurtadorUrl/api"
	"log/slog"
	"net/http"
	"time"
)

// A função principal do programa. Ela chama a função `run` para iniciar o servidor HTTP.
// Caso ocorra um erro durante a execução, ele será registrado no log.
func main() {
	if err := run(); err != nil {
		// Registra o erro caso o servidor falhe ao iniciar ou durante a execução.
		slog.Error("failed to execute code", "error", err)
		return
	}
	// Mensagem de log indicando que o sistema foi encerrado.
	slog.Info("all systems offline")
}

// A função `run` inicializa e inicia um servidor HTTP com timeouts predefinidos e um handler de URLs.
// Ela cria um banco de dados em memória (mapa) para armazenar os mapeamentos de URLs e configura o servidor
// para processar as requisições recebidas. O servidor escuta na porta 8080.
// Retorna um erro caso o servidor falhe ao iniciar ou encontre problemas durante a execução.
func run() error {
	// Banco de dados em memória para armazenar os mapeamentos de URLs curtas e originais.
	db := make(map[string]string)

	// Cria o handler que gerencia as rotas e a lógica do encurtador de URLs.
	handler := api.NewHandler(db)

	// Configura o servidor HTTP com timeouts e o handler definido.
	s := http.Server{
		ReadTimeout:  10 * time.Second, // Tempo máximo para leitura de requisições.
		IdleTimeout:  time.Minute,      // Tempo máximo de inatividade de conexões.
		WriteTimeout: 10 * time.Second, // Tempo máximo para escrita de respostas.
		Addr:         ":8080",          // Porta onde o servidor irá escutar.
		Handler:      handler,          // Handler responsável por processar as requisições.
	}

	// Inicia o servidor e retorna um erro caso algo dê errado.
	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
