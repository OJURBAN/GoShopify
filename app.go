package main

import (
	"alin/packages/session"
	"alin/packages/shopify"
	"context"
	"fmt"
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
	newS := session.NewSession()
	resp, err := newS.Get("https://api.ipify.org?format=json", map[string][]string{})
	if err != nil {
	}
	respDump, err := httputil.DumpResponse(resp, true)
	fmt.Printf("[GET] Resp: %s", respDump)
	//
	//data := map[string]interface{}{
	//	"quantity": 1,
	//	"id":       123,
	//}
	//
	//headers := http.Header{
	//	"Host":            {"R1.co.uk"},
	//	"Accept":          {"application/json, text/plain, */*"},
	//	"Accept-Language": {"en-GB,en;q=0.5"},
	//	"Origin":          {fmt.Sprintf("https://%s", "inst.Domain")},
	//	"Connection":      {"keep-alive"},
	//	"Referer":         {fmt.Sprintf("https://%s/%s?variant=%s", "inst.Domain", "inst.ProductLoc", "inst.VariantID")},
	//	"Sec-Fetch-Dest":  {"empty"},
	//	//"Sec-Fetch-Mode":  {"cors"},
	//	"Sec-Fetch-Site": {"same-origin"},
	//	"Content-Type":   {"application/json;charset=utf-8"},
	//}
	//
	//postResp, err := newS.PostJson("https://httpbin.org/response-headers", headers, data)
	//if err != nil {
	//}
	//
	//respDump, err = io.ReadAll(postResp.Body)
	//respStr := string(respDump)
	//fmt.Printf("\n[POST] Resp: %s", respStr)
	//defer postResp.Body.Close()
	//
	//getHeaders, err := newS.Get("https://httpbin.org/headers", headers)
	//if err != nil {
	//}
	//
	//respDump, err = io.ReadAll(getHeaders.Body)
	//respStr = string(respDump)
	//fmt.Printf("\n[getHeaders] Headers: %s", respStr)
	//defer postResp.Body.Close()
	//
	//formData := url.Values{}
	//formData.Set("name", "foo")
	//formData.Set("surname", "bar")
	//
	//resp, err = newS.PostForm("https://httpbin.org/post", map[string][]string{}, formData)
	//if err != nil {
	//}
	//respDump, err = httputil.DumpResponse(resp, true)
	//fmt.Printf("[POST] Resp: %s", respDump)

	shopify.Test()
}

// startup is called when the app starts. The context is saved
// , so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) MakeRequest() string {
	// client := resty.New()
	// resp, err := client.R().
	// EnableTrace().
	// Get("https://httpbin.org/get")

	// var session Session = *NewSession()
	newS := session.NewSession()

	resp, err := newS.Get("https://httpbin.org/get", nil)
	if err != nil {
	}
	// resp, err = session.Session.Get("https://httpbin.org/get", nil)
	// session.Get("https://httpbin.org/get", nil)

	// Explore response object
	// fmt.Println("Response Info:")
	// fmt.Println("  Error      :", err)
	// fmt.Println("  Status Code:", resp.StatusCode())
	// fmt.Println("  Status     :", resp.Status())
	// fmt.Println("  Proto      :", resp.Proto())
	// fmt.Println("  Time       :", resp.Time())
	// fmt.Println("  Received At:", resp.ReceivedAt())
	// fmt.Println("  Body       :\n", resp)
	// fmt.Println()

	return fmt.Sprintf("User-Agent: %s\nGET Response Info:\n%v", newS.Useragent, resp)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
