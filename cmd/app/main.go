package main

import (
	"github.com/instinctG/lru-cache/internal/config"
)

func main() {

	//TODO:init config : env
	cfg := config.MustLoad()

	//TODO: init logger : slog

	//TODO: init lru-cache : in-memory lru-cache

	//TODO: init router : chi, "chi render"

	//TODO: run server
}
