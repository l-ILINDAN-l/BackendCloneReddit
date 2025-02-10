package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/gorilla/mux"
	"github.com/l-ILINDAN-l/BackendCloneReddit/internal/repository"
)

type MemoryService struct {
	UserRepo    repository.UserRepository
	SessionRepo repository.SessionRepository
	PostRepo    repository.PostRepository
}

type Server struct {
	MemServ *MemoryService
	Router  *mux.Router
	Addr    string
	KeyJWT  string
}

// Structure for reading JSON
type Config struct {
	KeyJWT string `json:"keyJWT"`
}

func NewServer(addr, pathConfig string) *Server {
	// Reading the secret key from a JSON file
	// You can do this via the environment (.env)
	config, err := readConfig(pathConfig)
	if err != nil {
		fmt.Println("Error reading config:", err)
		return nil
	}

	return &Server{
		MemServ: &MemoryService{
			UserRepo:    repository.NewMemoryUserRepository(),
			SessionRepo: repository.NewMemorySessionRepository(),
			PostRepo:    repository.NewMemoryPostRepository(),
		},
		Router: mux.NewRouter().StrictSlash(true),
		Addr:   addr,
		KeyJWT: config.KeyJWT,
	}
}

// Function returning config from json file by filePath
func readConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
