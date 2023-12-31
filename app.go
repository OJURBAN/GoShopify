package main

import (
	"alin/packages/shopify"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func main() {
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
	}
	respDump, err := httputil.DumpResponse(resp, true)
	fmt.Printf("[GET] Resp: %s", respDump)

	shopify.Test()
}

// startup is called when the app starts. The context is saved
// , so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
