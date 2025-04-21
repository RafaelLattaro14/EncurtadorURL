package api

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewHandler cria e retorna um handler HTTP configurado com as rotas e middlewares necessários.
// Ele utiliza o pacote `chi` para gerenciar as rotas e middlewares.
func NewHandler(db map[string]string) http.Handler {
	r := chi.NewMux()
	// Middlewares para recuperação de erros, geração de IDs de requisição e logging.
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// Rota para encurtar URLs (POST /api/shorten).
	r.Post("/api/shorten", handlePost(db))
	// Rota para redirecionar URLs curtas para as originais (GET /{code}).
	r.Get("/{code}", handleGet(db))
	return r
}

// PostBody representa o corpo da requisição JSON para encurtar uma URL.
type PostBody struct {
	URL string `json:"url"` // URL a ser encurtada.
}

// Response representa a estrutura de resposta JSON enviada ao cliente.
type Response struct {
	Error string `json:"error,omitempty"` // Mensagem de erro, se houver.
	Data  any    `json:"data,omitempty"`  // Dados retornados, como o código gerado.
}

// sendJSON é uma função utilitária para enviar respostas JSON ao cliente.
// Ela define o cabeçalho "Content-Type", serializa a resposta e escreve no ResponseWriter.
func sendJSON(w http.ResponseWriter, resp Response, status int) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(resp)
	if err != nil {
		// Loga o erro e envia uma resposta genérica caso a serialização falhe.
		slog.Error("failed to marshal json data", "error", err)
		sendJSON(
			w,
			Response{Error: "something went wrong"},
			http.StatusInternalServerError,
		)
		return
	}
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		// Loga o erro caso a escrita da resposta falhe.
		slog.Error("failed to write response to client", "error", err)
		return
	}
}

// handlePost processa requisições para encurtar URLs (POST /api/shorten).
// Ele valida o corpo da requisição, gera um código único e armazena o mapeamento no banco de dados.
func handlePost(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body PostBody
		// Decodifica o corpo da requisição JSON.
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			// Retorna erro caso o corpo seja inválido.
			sendJSON(w, Response{Error: "invalid body"},
				http.StatusUnprocessableEntity,
			)
			return
		}
		// Valida a URL fornecida.
		if _, err := url.Parse(body.URL); err != nil {
			sendJSON(w, Response{Error: "invalid url passed"},
				http.StatusBadRequest,
			)
		}
		// Gera um código único e armazena no banco de dados.
		code := genCode()
		db[code] = body.URL
		// Retorna o código gerado ao cliente.
		sendJSON(w, Response{Data: code}, http.StatusCreated)
	}
}

// Conjunto de caracteres usados para gerar códigos curtos.
const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// genCode gera um código aleatório de 8 caracteres para representar uma URL curta.
func genCode() string {
	const n = 8
	byts := make([]byte, n)
	for i := range byts {
		byts[i] = characters[rand.Intn(len(characters))]
	}
	return string(byts)
}

// handleGet processa requisições para redirecionar URLs curtas (GET /{code}).
// Ele busca o código no banco de dados e redireciona o cliente para a URL original.
func handleGet(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtém o código da URL a partir dos parâmetros da rota.
		code := chi.URLParam(r, "code")
		// Busca a URL original no banco de dados.
		data, ok := db[code]
		if !ok {
			// Retorna erro 404 caso o código não seja encontrado.
			http.Error(w, "url not found", http.StatusNotFound)
			return
		}
		// Redireciona o cliente para a URL original.
		http.Redirect(w, r, data, http.StatusPermanentRedirect)
	}
}
