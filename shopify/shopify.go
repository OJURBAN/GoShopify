package shopify

import (
	"alin/packages/session"
	"alin/packages/shopify/data_handling"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

type Cart struct {
	Id                           int64         `json:"id"`
	Properties                   interface{}   `json:"properties"`
	Quantity                     int           `json:"quantity"`
	VariantId                    int64         `json:"variant_id"`
	Key                          string        `json:"key"`
	Title                        string        `json:"title"`
	Price                        int           `json:"price"`
	OriginalPrice                int           `json:"original_price"`
	DiscountedPrice              int           `json:"discounted_price"`
	LinePrice                    int           `json:"line_price"`
	OriginalLinePrice            int           `json:"original_line_price"`
	TotalDiscount                int           `json:"total_discount"`
	Discounts                    []interface{} `json:"discounts"`
	Sku                          string        `json:"sku"`
	Grams                        int           `json:"grams"`
	Vendor                       string        `json:"vendor"`
	Taxable                      bool          `json:"taxable"`
	ProductId                    int64         `json:"product_id"`
	ProductHasOnlyDefaultVariant bool          `json:"product_has_only_default_variant"`
	GiftCard                     bool          `json:"gift_card"`
	FinalPrice                   int           `json:"final_price"`
	FinalLinePrice               int           `json:"final_line_price"`
	Url                          string        `json:"url"`
	FeaturedImage                struct {
		AspectRatio float64 `json:"aspect_ratio"`
		Alt         string  `json:"alt"`
		Height      int     `json:"height"`
		Url         string  `json:"url"`
		Width       int     `json:"width"`
	} `json:"featured_image"`
	Image                    string   `json:"image"`
	Handle                   string   `json:"handle"`
	RequiresShipping         bool     `json:"requires_shipping"`
	ProductType              string   `json:"product_type"`
	ProductTitle             string   `json:"product_title"`
	UntranslatedProductTitle string   `json:"untranslated_product_title"`
	ProductDescription       string   `json:"product_description"`
	VariantTitle             string   `json:"variant_title"`
	UntranslatedVariantTitle string   `json:"untranslated_variant_title"`
	VariantOptions           []string `json:"variant_options"`
	OptionsWithValues        []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"options_with_values"`
	LineLevelDiscountAllocations []interface{} `json:"line_level_discount_allocations"`
	LineLevelTotalDiscount       int           `json:"line_level_total_discount"`
}

type Variant struct {
	ID    string `json:"id"`
	Image struct {
		Src string `json:"src"`
	} `json:"image"`
	Price struct {
		Amount       float64 `json:"amount"`
		CurrencyCode string  `json:"currencyCode"`
	} `json:"price"`
	Product struct {
		ID                string `json:"id"`
		Title             string `json:"title"`
		UntranslatedTitle string `json:"untranslatedTitle"`
		Vendor            string `json:"vendor"`
		Type              string `json:"type"`
	} `json:"product"`
	Sku               string `json:"sku"`
	Title             string `json:"title"`
	UntranslatedTitle string `json:"untranslatedTitle"`
}

type ShopifyStore struct {
	Domain         string
	StoreCode      int
	CheckoutDomain string
	DepositDomain  string
}

type Instance struct {
	TaskID     int
	URL        string
	Profile    data_handling.CheckoutProfile
	VariantID  string
	Variants   []Variant
	Store      ShopifyStore
	Domain     string
	ProductLoc string
	Session    *session.Session
	Logger     *zap.Logger
	Status     string
}

func NewShopifyInstance(taskId int, url string, variantId string, profile data_handling.CheckoutProfile) (*Instance, error) {
	inst := new(Instance)

	r, _ := regexp.Compile("(?:https|http)\\://([\\w.]+)/([\\w\\d\\/-]+)?")
	match := r.FindStringSubmatch(url)

	if len(match) < 2 {
		fmt.Println("Could not regex match URL")
		return nil, errors.New("could not match Regex")
	} else {
		inst.Domain = match[1]
		inst.ProductLoc = match[2]
	}

	var stores []ShopifyStore
	stores = append(stores, ShopifyStore{
		Domain:         "launches.routeone.co.uk",
		StoreCode:      50487623851,
		CheckoutDomain: "checkout.shopifycs.com",
		DepositDomain:  "deposit.us.shopifycs.com/sessions",
	})
	stores = append(stores, ShopifyStore{
		Domain:         "www.routeone.co.uk",
		StoreCode:      50487623851,
		CheckoutDomain: "checkout.shopifycs.com",
		DepositDomain:  "deposit.us.shopifycs.com/sessions",
	})

	inst.Session = session.NewSession()
	inst.Logger = session.NewLogger()
	inst.TaskID = taskId
	inst.URL = url
	inst.Profile = profile
	inst.VariantID = variantId
	return inst, nil
}

func (inst Instance) getVariants() (bool, error) {
	resp, err := inst.Session.Get(inst.URL, map[string][]string{})

	if err != nil {
	}

	if resp.StatusCode != 200 {
		return false, errors.New("Could not retrieve variants")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	fmt.Printf("STATUS %v", bodyString)

	r, _ := regexp.Compile(`"productVariants":(\[[{":,\\/.?=} '\-\w]+\])`)
	match := r.FindStringSubmatch(bodyString)

	var m []Variant
	if err := json.Unmarshal([]byte(match[1]), &m); err != nil {
		panic(err)
	}

	inst.Variants = m
	inst.Logger.Info("GET Variants", zap.String("Num. loaded", fmt.Sprintf("%d variants", len(m))))

	return true, nil
}

func (inst Instance) cartVariant() (bool, error) {
	headers := http.Header{
		"Host":            {inst.Domain},
		"Accept":          {"application/json, text/plain, */*"},
		"Accept-Language": {"en-GB,en;q=0.5"},
		"Origin":          {fmt.Sprintf("https://%s", inst.Domain)},
		"Connection":      {"keep-alive"},
		"Referer":         {fmt.Sprintf("https://%s/%s?variant=%s", inst.Domain, inst.ProductLoc, inst.VariantID)},
		"Sec-Fetch-Dest":  {"empty"},
		"Sec-Fetch-Mode":  {"cors"},
		"Sec-Fetch-Site":  {"same-origin"},
		"Content-Type":    {"application/json;charset=utf-8"},
	}

	resp, err := inst.Session.Get(
		fmt.Sprintf("https://%s/cart/add.js", inst.Domain),
		headers,
	)

	if err != nil {
		inst.Logger.Error("GET CART", zap.Error(err))
	}

	//data = {"quantity": 1, "id": self.variant_id}
	data := map[string]interface{}{
		"quantity": 1,
		"id":       inst.VariantID,
	}

	resp, err = inst.Session.PostJson(
		fmt.Sprintf("https://%s/cart/add.js", inst.Domain),
		headers,
		data,
	)

	respDump, err := io.ReadAll(resp.Body)

	var cart Cart
	if err := json.Unmarshal([]byte(respDump), &cart); err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		inst.Logger.Info("Potential error", zap.String("Cart variant request status code", strconv.Itoa(resp.StatusCode)))
	}

	if err != nil {
	}

	inst.Status = fmt.Sprintf("Added %s to cart @ £%f", cart.Title, float32(cart.Price)/100)

	return true, nil
}

func (inst Instance) initCheckout() (bool, error) {
	headers := http.Header{
		"Host":            {inst.Domain},
		"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
		"Accept-Language": {"en-GB,en;q=0.5"},
		"Origin":          {fmt.Sprintf("https://%s", inst.Domain)},
		"Connection":      {"keep-alive"},
		"Referer":         {fmt.Sprintf("https://%s/%s?variant=%s", inst.Domain, inst.ProductLoc, inst.VariantID)},
		"Sec-Fetch-Dest":  {"document"},
		"Sec-Fetch-Mode":  {"navigate"},
		"Sec-Fetch-Site":  {"same-origin"},
		"Sec-Fetch-User":  {"?1"},
		"Content-Type":    {"application/x-www-form-urlencoded"},
	}

	resp, err := inst.Session.PostJson(
		fmt.Sprintf("https://%s/checkout", inst.Domain),
		headers,
		nil,
	)
	//resp, err := inst.Session.Get(
	//	"https://httpbin.org/cookies",
	//	headers,
	//)
	if err != nil {
		inst.Logger.Error("POST CART", zap.Error(err))
	}

	respDump, err := io.ReadAll(resp.Body)
	respStr := string(respDump)
	inst.Logger.Info("Begin checkout", zap.String("Resp", respStr))

	defer resp.Body.Close()

	success, err := writeToFile("beginCheckout", respStr)
	if err != nil {
		return false, err
	}

	if !success {
		inst.Logger.Info("Could not write to file")
	}

	if resp.StatusCode != 302 {
		inst.Logger.Info("Potential error", zap.String("Begin checkout request status code", strconv.Itoa(resp.StatusCode)))
	}

	return true, nil
}

func (inst Instance) printStatus(text string) {
	fmt.Printf("[TASK %d] %s\n", inst.TaskID, text)
}

func writeToFile(fileName string, text string) (success bool, err error) {
	f, err := os.Create(fmt.Sprintf("%s.json", fileName))

	if err != nil {
		return false, err
	}

	defer f.Close()

	_, err2 := f.WriteString(text)

	if err2 != nil {
		return false, err
	}

	return true, nil
}

func (inst Instance) run() {
	inst.printStatus("PLS")
	inst.getVariants()
	inst.cartVariant()
	inst.initCheckout()
}

func Test() {
	prof := data_handling.NewProfile()
	inst, err := NewShopifyInstance(1, "https://www.routeone.co.uk/products/route-one-super-baggy-denim-shorts-gingerbread-001155634", "40072422293581", prof)

	if err != nil {
	}

	inst.run()
	fmt.Print("INST: %v", inst)
}
