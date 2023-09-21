package shopify

import (
	"alin/packages/session"
	"alin/packages/shopify/data_handling"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CloudProxyHeaders struct {
	XShopifyCheckoutAuthorizationToken string `json:"X-Shopify-Checkout-Authorization-Token"`
	Accept                             string `json:"Accept"`
}

type CloudProxy struct {
	Cmd        string            `json:"cmd"`
	URL        string            `json:"url"`
	UserAgent  string            `json:"userAgent"`
	MaxTimeout int               `json:"maxTimeout"`
	Headers    CloudProxyHeaders `json:"headers"`
}

type ShippingRate struct {
	ID       string `json:"id"`
	Price    string `json:"price"`
	Title    string `json:"title"`
	Checkout struct {
		TotalTax      string `json:"total_tax"`
		TotalPrice    string `json:"total_price"`
		SubtotalPrice string `json:"subtotal_price"`
	} `json:"checkout"`
	PhoneRequired          bool  `json:"phone_required"`
	DeliveryRange          []any `json:"delivery_range"`
	EstimatedTimeInTransit any   `json:"estimated_time_in_transit"`
}

type ShippingRates struct {
	ShippingRate []ShippingRate `json:"shipping_rates"`
}

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
	Code           string
	CheckoutDomain string
	DepositDomain  string
}

type Tokens struct {
	ShopifyCheckoutToken               string
	AuthenticityToken                  string
	DeliveryAuthenticityToken          string
	XShopifyCheckoutAuthorizationToken string
	CheckoutGateway                    string
	CheckoutToken                      string
}

type Instance struct {
	TaskID         int
	URL            string
	Profile        data_handling.CheckoutProfile
	VariantID      string
	Store          ShopifyStore
	Domain         string
	ProductLoc     string
	Session        *session.Session
	Logger         *zap.Logger
	Status         string
	Tokens         Tokens
	ShippingRates  ShippingRates
	ShippingRate   ShippingRate
	PaymentGateway string
	Cart           Cart
	TotalPrice     float64
	Options        data_handling.Options
}

func NewShopifyInstance(options data_handling.Options) (*Instance, error) {
	inst := new(Instance)

	r, _ := regexp.Compile("(?:https|http)\\://([\\w.]+)/([\\w\\d\\/-]+)?")
	match := r.FindStringSubmatch(options.URL)

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
		Code:           "50487623851",
		CheckoutDomain: "checkout.shopifycs.com",
		DepositDomain:  "deposit.us.shopifycs.com/sessions",
	})
	stores = append(stores, ShopifyStore{
		Domain:         "www.routeone.co.uk",
		Code:           "27442937933",
		CheckoutDomain: "checkout.shopifycs.com",
		DepositDomain:  "deposit.us.shopifycs.com/sessions",
	})
	stores = append(stores, ShopifyStore{
		Domain:         "releases.flatspot.com",
		Code:           "2744451133",
		CheckoutDomain: "checkout.shopifycs.com",
		DepositDomain:  "deposit.us.shopifycs.com/sessions",
	})

	matchedStore := false
	for i := range stores {
		if stores[i].Domain == inst.Domain {
			// Found!
			inst.Store = stores[i]
			matchedStore = true
		}
	}

	if !matchedStore {
		fmt.Println("Could not find store")
		return nil, errors.New("Could not find store")
	}

	inst.Session = session.NewSession(options)
	inst.Logger = session.NewLogger()
	inst.TaskID = options.TaskID
	inst.URL = options.URL
	inst.Profile = options.Profile
	//inst.VariantID = options.VariantID
	return inst, nil
}

func (inst *Instance) getVariants() (bool, error) {
	resp, err := inst.Session.Get(inst.URL, map[string][]string{})

	if err != nil {
	}

	if resp.StatusCode != 200 {
		return false, errors.New("Could not retrieve variants")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	r, _ := regexp.Compile(`"productVariants":(\[[\{\}\"\,\w\/\\\w-_.:=? ()]+\])`)
	match := r.FindStringSubmatch(bodyString)

	var m []Variant
	if err := json.Unmarshal([]byte(match[1]), &m); err != nil {
		panic(err)
	}

	for i := range m {
		if strings.Contains(m[i].Title, inst.Options.Size) {
			// Found!
			inst.VariantID = m[i].ID
		}
	}

	inst.Logger.Info("GET Variants", zap.String("Num. loaded", fmt.Sprintf("%d variants", len(m))))

	return true, nil
}

func (inst *Instance) cartVariant() (bool, error) {
	type Payload struct {
		Quantity int    `json:"quantity"`
		ID       string `json:"id"`
	}

	data := Payload{
		1,
		inst.VariantID,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		// handle err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/cart/add.js", inst.Domain), body)
	if err != nil {
		// handle err
	}
	req.Host = inst.Domain
	req.Header.Set("User-Agent", inst.Session.Useragent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Origin", fmt.Sprintf("https://%s", inst.Domain))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/%s?variant=%s", inst.Domain, inst.ProductLoc, inst.VariantID))
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	resp, err := inst.Session.Client.Do(req)
	if err != nil {
		// handle err
	}
	respDump, err := io.ReadAll(resp.Body)
	respStr := string(respDump)

	inst.Logger.Debug("Cart", zap.String("Resp", respStr))

	var cart Cart
	if err := json.Unmarshal([]byte(respDump), &cart); err != nil {
		panic(err)
	}

	inst.Cart = cart

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		inst.Logger.Info("Potential error", zap.String("Cart variant request status code", strconv.Itoa(resp.StatusCode)))
		return false, errors.New("Could not cart variant")
	}

	if err != nil {
	}

	inst.Status = fmt.Sprintf("Added %s to cart @ £%f", cart.Title, float32(cart.Price)/100)

	return true, nil
}

func (inst *Instance) initCheckout() (bool, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/checkout", inst.Domain), nil)
	if err != nil {
		// handle err
	}
	req.Host = inst.Domain
	req.Header.Set("User-Agent", inst.Session.Useragent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Origin", fmt.Sprintf("https://%s", inst.Domain))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/%s?variant=%s", inst.Domain, inst.ProductLoc, inst.VariantID))
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := inst.Session.Client.Do(req)
	if err != nil {
		// handle err
	}
	respDump, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	respStr := string(respDump)

	if resp.StatusCode != 302 {
		inst.Logger.Info("Potential error", zap.String("Begin checkout request status code", strconv.Itoa(resp.StatusCode)))
	}

	newLoc := resp.Header.Get("Location")

	//if newLoc doesn't contain checkout return false
	if !strings.Contains(newLoc, "checkout") {
		inst.Logger.Info("Redirected back to cart")
		return false, errors.New("Redirected back to cart")
	}

	// Extract token
	req, err = http.NewRequest("GET", newLoc, nil)
	if err != nil {
		// handle err
	}

	req.Host = inst.Domain
	req.Header.Set("User-Agent", inst.Session.Useragent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/%s?variant=%s", inst.Domain, inst.ProductLoc, inst.VariantID))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err = inst.Session.Client.Do(req)
	if err != nil {
		inst.Logger.Error("Error sending request", zap.Error(err))
	}

	respDump, err = io.ReadAll(resp.Body)
	respStr = string(respDump)

	defer resp.Body.Close()

	// Extract token with regex `Shopify.Checkout.token = "([\w]+)"` from respStr
	r, _ := regexp.Compile(`Shopify.Checkout.token = "(\w+)"`)
	match := r.FindStringSubmatch(respStr)

	if len(match) < 2 {
		inst.Logger.Info("Could not regex match Shopify.Checkout.token")
		return false, errors.New("Could not regex match Shopify.Checkout.token")
	}

	inst.Tokens.ShopifyCheckoutToken = match[1]

	return true, nil
}

func (inst *Instance) authToken() (bool, error) {
	// Extract token

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/%s/checkouts/%s", inst.Domain, inst.Store.Code, inst.Tokens.ShopifyCheckoutToken), nil)
	if err != nil {
		// handle err
	}

	req.Host = inst.Domain
	req.Header.Set("User-Agent", inst.Session.Useragent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/%s?variant=%s", inst.Domain, inst.ProductLoc, inst.VariantID))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err := inst.Session.Client.Do(req)
	if err != nil {
		inst.Logger.Error("Error sending request", zap.Error(err))
	}

	respDump, err := io.ReadAll(resp.Body)
	respStr := string(respDump)

	defer resp.Body.Close()

	// Extract token with regex `Shopify.Checkout.token = "([\w]+)"` from respStr
	r, _ := regexp.Compile(`name=\"authenticity_token" value="([a-zA-Z0-9_-]+)\"`)
	match := r.FindStringSubmatch(respStr)

	if len(match) < 2 {
		inst.Logger.Info("Could not regex match authenticity_token")
		return false, errors.New("Could not regex match authenticity_token")
	}

	inst.Tokens.AuthenticityToken = match[1]

	return true, nil
}

func (inst *Instance) submitAddress() (bool, error) {
	params := url.Values{}
	params.Add("_method", `patch`)
	params.Add("authenticity_token", inst.Tokens.AuthenticityToken)
	params.Add("previous_step", `contact_information`)
	params.Add("step", `shipping_method`)
	params.Add("checkout[email_or_phone]", inst.Profile.Email)
	params.Add("checkout[buyer_accepts_marketing]", `0`)
	params.Add("checkout[shipping_address][country]", inst.Profile.Country)
	params.Add("checkout[shipping_address][first_name]", inst.Profile.Fname)
	params.Add("checkout[shipping_address][last_name]", inst.Profile.Lname)
	params.Add("checkout[shipping_address][address1]", inst.Profile.Address1)
	params.Add("checkout[shipping_address][address2]", ``)
	params.Add("checkout[shipping_address][city]", inst.Profile.City)
	params.Add("checkout[shipping_address][zip]", inst.Profile.Zipcode)
	params.Add("checkout[shipping_address][phone]", fmt.Sprintf("0%s", inst.Profile.Phone))
	params.Add("checkout[buyer_accepts_sms]", `0`)
	params.Add("checkout[sms_marketing_phone]", ``)
	params.Add("checkout[client_details][browser_width]", strconv.Itoa(rand.Intn(2000-1000)+1000))
	params.Add("checkout[client_details][browser_height]", strconv.Itoa(rand.Intn(2000-1000)+1000))
	params.Add("checkout[client_details][javascript_enabled]", `1`)
	params.Add("checkout[client_details][color_depth]", `30`)
	params.Add("checkout[client_details][java_enabled]", `false`)
	params.Add("checkout[client_details][browser_tz]", `-60`)
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/%s/checkouts/%s", inst.Domain, inst.Store.Code, inst.Tokens.ShopifyCheckoutToken), body)
	if err != nil {
		// handle err
	}
	req.Host = inst.Domain
	req.Header.Set("User-Agent", inst.Session.Useragent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Origin", fmt.Sprintf("https://%s", inst.Domain))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/%s?variant=%s", inst.Domain, inst.ProductLoc, inst.VariantID))
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := inst.Session.Client.Do(req)
	if err != nil {
		inst.Logger.Error("Error sending request", zap.Error(err))
		return false, err
	}

	respDump, err := io.ReadAll(resp.Body)
	respStr := string(respDump)

	defer resp.Body.Close()
	//respStr := string(respDump)

	if resp.StatusCode != 302 {
		inst.Logger.Info("Potential error", zap.String("Begin checkout request status code", strconv.Itoa(resp.StatusCode)), zap.String("Resp message", respStr))
	}

	return true, nil
}

func (inst *Instance) deliveryToken() (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/%s/checkouts/%s", inst.Domain, inst.Store.Code, inst.Tokens.ShopifyCheckoutToken), nil)

	q := req.URL.Query()
	q.Add("previous_step", "contact_information")
	q.Add("step", "shipping_method")
	req.URL.RawQuery = q.Encode()

	if err != nil {
		inst.Logger.Error("Error creating request", zap.Error(err))
		return false, err
	}

	req.Host = inst.Domain
	req.Header.Set("User-Agent", inst.Session.Useragent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	//req.Header.Set("Referer", fmt.Sprintf("https://%s/%s?variant=%s", inst.Domain, inst.ProductLoc, inst.VariantID))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err := inst.Session.Client.Do(req)
	if err != nil {
		inst.Logger.Error("Error sending request", zap.Error(err))
	}

	respDump, err := io.ReadAll(resp.Body)
	respStr := string(respDump)

	defer resp.Body.Close()

	// Extract token with regex `Shopify.Checkout.token = "([\w]+)"` from respStr
	r, _ := regexp.Compile(`name="authenticity_token" value="([a-zA-Z0-9_-]+)"`)
	match := r.FindStringSubmatch(respStr)

	if len(match) < 2 {
		inst.Logger.Info("Could not regex match delivery authenticity_token")
		return false, errors.New("Could not regex match delivery authenticity_token")
	}

	inst.Tokens.DeliveryAuthenticityToken = match[1]

	// Extract token with regex `Shopify.Checkout.token = "([\w]+)"` from respStr
	r, _ = regexp.Compile(`name="shopify-checkout-authorization-token" content="(.+)"`)
	match = r.FindStringSubmatch(respStr)

	if len(match) < 2 {
		inst.Logger.Info("Could not regex match shopify-checkout-authorization-token")
		return false, errors.New("Could not regex match shopify-checkout-authorization-token")
	}
	//XShopifyCheckoutAuthorizationToken
	inst.Tokens.XShopifyCheckoutAuthorizationToken = match[1]

	return true, nil
}

func (inst *Instance) getShippingRates() (bool, error) {
	//cloudProxyHeaders := CloudProxyHeaders{
	//	XShopifyCheckoutAuthorizationToken: inst.Tokens.XShopifyCheckoutAuthorizationToken,
	//	Accept:                             "application/json",
	//}
	//cloudProxyPostData, err := json.Marshal(CloudProxy{
	//	Cmd:        "request.get",
	//	URL:        fmt.Sprintf("https://%s/api/checkouts/%s/shipping_rates", inst.Domain, inst.Tokens.ShopifyCheckoutToken),
	//	UserAgent:  inst.Session.Useragent,
	//	MaxTimeout: 10000,
	//	Headers:    cloudProxyHeaders,
	//})
	//
	//var jsonData = []byte(fmt.Sprintf(`{"cmd": "request.get", "url": "https://%s/api/checkouts/%s/shipping_rates", "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.41 Safari/537.36", "maxTimeout": 60000, "headers": {"X-Shopify-Checkout-Authorization-Token": %s, "Accept": "application/json"}}`, inst.Domain, inst.Tokens.ShopifyCheckoutToken, inst.Tokens.XShopifyCheckoutAuthorizationToken))
	//
	//if err != nil {
	//}
	//
	//fmt.Println("Create cloudProxyPostData", string(cloudProxyPostData))
	//
	//// Extract token
	////shippingRatesUrl := fmt.Sprintf("https://%s/api/checkouts/%s/shipping_rates", inst.Domain, inst.Tokens.ShopifyCheckoutToken)
	////req, err := http.NewRequest("GET", shippingRatesUrl, nil)
	//req, err := http.NewRequest("POST", "http://localhost:8191/v1", bytes.NewBuffer(jsonData))
	////req, err := http.NewRequest("GET", "https://httpbin.org/headers", nil)
	//if err != nil {
	//	// handle err
	//}
	//
	//// Setup transport
	//t := http.DefaultTransport.(*http.Transport).Clone()
	//t.MaxIdleConns = 100
	//t.MaxConnsPerHost = 100
	//t.MaxIdleConnsPerHost = 100
	//
	////proxy := http.ProxyURL(&url.URL{
	////	Scheme: "http",
	////	Host:   "127.0.0.1:9090",
	////})
	////t.Proxy = proxy
	//
	//tempClient := &http.Client{
	//	Transport: t,
	//	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	//		return http.ErrUseLastResponse
	//	},
	//}
	//
	////req.Header = http.Header{}
	//req.Host = inst.Domain
	//req.Header.Set("X-Shopify-Checkout-Authorization-Token", inst.Tokens.XShopifyCheckoutAuthorizationToken)
	//req.Header.Set("User-Agent", inst.Session.Useragent)
	//req.Header.Set("Upgrade-Insecure-Requests", "1")
	//req.Header.Set("Accept", "gzip")
	//req.Header.Set("Connection", "keep-alive")
	//
	//var resp *http.Response
	//
	//for {
	//	inst.Logger.Info("Retrieving shipping rates")
	//	resp, err = tempClient.Do(req)
	//	if err != nil {
	//		inst.Logger.Error("Error sending request", zap.Error(err))
	//	}
	//
	//	respDump, err := ioutil.ReadAll(resp.Body)
	//	if err != nil {
	//	}
	//	respStr := string(respDump)
	//	inst.Logger.Info("Shipping rates", zap.String("Rates resp", respStr))
	//	time.Sleep(1 * time.Second)
	//	defer resp.Body.Close()
	//}

	//{
	// "shipping_rates": [
	//   {
	//     "id": "Store Pickup + Delivery-standard-3-5-working-days-3.99",
	//     "price": "3.99",
	//     "title": "Standard 3 - 5 Working Days",
	//     "checkout": {
	//       "total_tax": "9.00",
	//       "total_price": "53.94",
	//       "subtotal_price": "49.95"
	//     },
	//     "phone_required": false,
	//     "delivery_range": [],
	//     "estimated_time_in_transit": null
	//   },
	//   {
	//     "id": "Store Pickup + Delivery-next-working-day-5.99",
	//     "price": "5.99",
	//     "title": "Next Working Day.",
	//     "checkout": {
	//       "total_tax": "9.33",
	//       "total_price": "55.94",
	//       "subtotal_price": "49.95"
	//     },
	//     "phone_required": false,
	//     "delivery_range": [],
	//     "estimated_time_in_transit": null
	//   },
	//   {
	//     "id": "Store Pickup + Delivery-saturday-delivery-9.99",
	//     "price": "9.99",
	//     "title": "Saturday Delivery",
	//     "checkout": {
	//       "total_tax": "10.00",
	//       "total_price": "59.94",
	//       "subtotal_price": "49.95"
	//     },
	//     "phone_required": false,
	//     "delivery_range": [],
	//     "estimated_time_in_transit": null
	//   }
	// ]
	//}

	rawJson := `{
	 "shipping_rates": [
	   {
	     "id": "Store Pickup + Delivery-standard-3-5-working-days-3.99",
	     "price": "3.99",
	     "title": "Standard 3 - 5 Working Days",
	     "checkout": {
	       "total_tax": "9.00",
	       "total_price": "53.94",
	       "subtotal_price": "49.95"
	     },
	     "phone_required": false,
	     "delivery_range": [],
	     "estimated_time_in_transit": null
	   },
	   {
	     "id": "Store Pickup + Delivery-next-working-day-5.99",
	     "price": "5.99",
	     "title": "Next Working Day.",
	     "checkout": {
	       "total_tax": "9.33",
	       "total_price": "55.94",
	       "subtotal_price": "49.95"
	     },
	     "phone_required": false,
	     "delivery_range": [],
	     "estimated_time_in_transit": null
	   },
	   {
	     "id": "Store Pickup + Delivery-saturday-delivery-9.99",
	     "price": "9.99",
	     "title": "Saturday Delivery",
	     "checkout": {
	       "total_tax": "10.00",
	       "total_price": "59.94",
	       "subtotal_price": "49.95"
	     },
	     "phone_required": false,
	     "delivery_range": [],
	     "estimated_time_in_transit": null
	   }
	 ]
	}`

	var shippingRates ShippingRates
	if err := json.Unmarshal([]byte(rawJson), &shippingRates); err != nil {
	}

	inst.ShippingRates = shippingRates
	inst.ShippingRate = shippingRates.ShippingRate[0]

	return true, nil
}

func (inst *Instance) submitDelivery() (bool, error) {
	params := url.Values{}
	params.Add("_method", `patch`)
	params.Add("authenticity_token", inst.Tokens.DeliveryAuthenticityToken)
	params.Add("previous_step", `shipping_method`)
	params.Add("step", `payment_method`)
	//params.Add("checkout[shipping_rate][id]", inst.ShippingRate.ID)
	params.Add("checkout[shipping_rate][id]", "shopify-DPD%20Local%20Standard%20Shipping%20(1-3%20Business%20Days)-6.00")
	params.Add("checkout[client_details][browser_width]", `1280`)
	params.Add("checkout[client_details][browser_height]", `643`)
	params.Add("checkout[client_details][javascript_enabled]", `1`)
	params.Add("checkout[client_details][color_depth]", `30`)
	params.Add("checkout[client_details][java_enabled]", `false`)
	params.Add("checkout[client_details][browser_tz]", `-60`)
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/%s/checkouts/%s", inst.Domain, inst.Store.Code, inst.Tokens.ShopifyCheckoutToken), body)
	if err != nil {
		// handle err
	}
	// TODO: Change host to be taken from inst.Store
	req.Host = inst.Domain
	req.Header.Set("User-Agent", inst.Session.Useragent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", inst.Domain))
	req.Header.Set("Origin", fmt.Sprintf("https://%s/", inst.Domain))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := inst.Session.Client.Do(req)
	if err != nil {
		// handle err
	}

	defer resp.Body.Close()

	return true, nil
}

func (inst *Instance) getGateway() (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/%s/checkouts/%s?previous_step=shipping_method&step=payment_method", inst.Domain, inst.Store.Code, inst.Tokens.ShopifyCheckoutToken), nil)
	if err != nil {
		inst.Logger.Error("Error creating request", zap.Error(err))
		// handle err
	}
	// TODO: Change host to be taken from inst.Store
	req.Host = inst.Domain
	req.Header.Set("User-Agent", inst.Session.Useragent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	//req.Header.Set("Content-Type", "text/html; charset=utf-8")
	//req.Header.Set("Accept-Encoding", "deflate, br")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", inst.Domain))
	//req.Header.Set("Origin", fmt.Sprintf("https://%s/", inst.Domain))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err := inst.Session.Client.Do(req)
	if err != nil {
		// handle err
	}

	respDump, err := io.ReadAll(resp.Body)
	respStr := string(respDump)

	defer resp.Body.Close()

	// Extract token with regex `data-select-gateway="(.+)"` from respStr
	r, _ := regexp.Compile(`data-select-gateway="(.+)"`)
	match := r.FindStringSubmatch(respStr)

	if len(match) < 2 {
		inst.Logger.Info("Could not regex match checkout gateway")
		return false, errors.New("Could not regex match checkout gateway")
	}

	inst.Tokens.CheckoutGateway = match[1]

	// Extract token with regex `data-select-gateway="(.+)"` from respStr
	r, _ = regexp.Compile(`Shopify.Checkout.totalPrice = (.+);`)
	match = r.FindStringSubmatch(respStr)

	if len(match) < 2 {
		inst.Logger.Info("Could not regex match total price")
		return false, errors.New("Could not regex match checkout gateway")
	}

	inst.TotalPrice, err = strconv.ParseFloat(match[1], 64)
	if err != nil {
		inst.Logger.Error("Error parsing total price", zap.Error(err))
	}

	// Extract token with regex `name="authenticity_token" value="([a-zA-Z0-9_-]+)"` from respStr
	r, _ = regexp.Compile(`name="authenticity_token" value="([a-zA-Z0-9_-]+)"`)
	match = r.FindStringSubmatch(respStr)

	if len(match) < 2 {
		inst.Logger.Info("Could not regex match checkout authenticity_token")
		return false, errors.New("Could not regex match checkout authenticity_token")
	}

	inst.Tokens.CheckoutToken = match[1]

	return true, nil
}

func (inst *Instance) createPaymentSession() (bool, error) {
	type Data struct {
		CreditCard          data_handling.CardDetails `json:"credit_card"`
		PaymentSessionScope string                    `json:"payment_session_scope"`
	}

	data := Data{
		CreditCard:          inst.Profile.Card,
		PaymentSessionScope: inst.Domain,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		inst.Logger.Error("Error marshalling data", zap.Error(err))
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s", inst.Store.DepositDomain), bytes.NewBuffer(jsonData))
	if err != nil {
		inst.Logger.Error("Error creating request", zap.Error(err))
	}

	req.Host = inst.Domain
	req.Header.Set("User-Agent", inst.Session.Useragent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", fmt.Sprintf("https://%s", inst.Domain))
	req.Header.Set("Origin", fmt.Sprintf("https://%s/", inst.Domain))
	req.Header.Set("Accept-Encoding", "br")

	resp, err := inst.Session.Client.Do(req)
	if err != nil {
		inst.Logger.Error("Error sending request", zap.Error(err))
	}

	respDump, err := io.ReadAll(resp.Body)
	respStr := string(respDump)

	defer resp.Body.Close()

	type PaymentId struct {
		ID string `json:"id"`
	}

	inst.Logger.Info("Payment session", zap.String("Payment session resp", respStr))

	var paymentID PaymentId
	if err := json.Unmarshal([]byte(respDump), &paymentID); err != nil {
	}

	inst.PaymentGateway = paymentID.ID

	return true, nil
}

func (inst *Instance) submitPayment() (bool, error) {
	params := url.Values{}
	params.Add("_method", `patch`)
	params.Add("authenticity_token", inst.Tokens.CheckoutToken)
	params.Add("previous_step", `payment_method`)
	params.Add("step", ``)
	params.Add("s", inst.PaymentGateway)
	params.Add("checkout[payment_gateway]", inst.Tokens.CheckoutGateway)
	params.Add("checkout[credit_card][vault]", `false`)
	params.Add("checkout[different_billing_address]", `false`)
	params.Add("checkout[remember_me]", `false`)
	params.Add("checkout[remember_me]", `0`)
	params.Add("checkout[vault_phone]", fmt.Sprintf("+44%s", inst.Profile.Phone))
	params.Add("checkout[total_price]", fmt.Sprintf("%v", int(inst.TotalPrice*100)))
	params.Add("complete", "1")
	params.Add("checkout[client_details][browser_width]", strconv.Itoa(rand.Intn(2000-1000)+1000))
	params.Add("checkout[client_details][browser_height]", strconv.Itoa(rand.Intn(2000-1000)+1000))
	params.Add("checkout[client_details][javascript_enabled]", `1`)
	params.Add("checkout[client_details][color_depth]", `30`)
	params.Add("checkout[client_details][java_enabled]", `false`)
	params.Add("checkout[client_details][browser_tz]", `-60`)

	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/%s/checkouts/%s", inst.Domain, inst.Store.Code, inst.Tokens.ShopifyCheckoutToken), body)
	if err != nil {
		// handle err
	}
	// TODO: Change host to be taken from inst.Store
	req.Host = inst.Domain
	req.Header.Set("User-Agent", inst.Session.Useragent)
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", inst.Domain))
	req.Header.Set("Origin", fmt.Sprintf("https://%s/", inst.Domain))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := inst.Session.Client.Do(req)
	if err != nil {
		inst.Logger.Error("Error sending payment request", zap.Error(err))
		return false, err
	}

	respDump, err := io.ReadAll(resp.Body)
	respStr := string(respDump)

	if resp.StatusCode != 302 {
		if err != nil {
			inst.Logger.Error("Error reading payment request body", zap.Error(err))
		}
		inst.Logger.Info("Potential error", zap.String("Payment request status code", strconv.Itoa(resp.StatusCode)), zap.String("Response body", respStr))
	}
	defer resp.Body.Close()

	// ? If 302, check if location contains /processing
	if strings.Contains(resp.Header.Get("Location"), "/processing") {

		fmt.Print("\nProcessing")

		for true {
			fmt.Print(".")
			parsedUrl, err := url.Parse(resp.Header.Get("Location"))
			if err != nil {
				inst.Logger.Error("Error parsing redirect URL", zap.Error(err))
			}
			req.Header.Del("Content-Type")
			req.URL = parsedUrl

			resp, err := inst.Session.Client.Do(req)
			if err != nil {
				inst.Logger.Error("Error sending payment request", zap.Error(err))
				return false, err
			}

			respDump, err := io.ReadAll(resp.Body)
			respStr := string(respDump)

			if resp.StatusCode != 302 {
				if err != nil {
					inst.Logger.Error("Error reading payment request body", zap.Error(err))
				}
				inst.Logger.Info("Potential error", zap.String("Payment request status code", strconv.Itoa(resp.StatusCode)), zap.String("Response body", respStr))
			}
			defer resp.Body.Close()
		}
	}

	if !strings.Contains(resp.Header.Get("Location"), "processing") || !strings.Contains(resp.Header.Get("Location"), "/thank_you") {
		inst.Logger.Info("Potential 3DS", zap.String("Status code", strconv.Itoa(resp.StatusCode)), zap.String("Link", resp.Header.Get("Location")))
		fmt.Println("Potential 3DS", resp.Header.Get("Location"))
		time.Sleep(2 * time.Minute)
	}

	success := false
	for success {
		time.Sleep(2 * time.Second)
		req.URL, err = url.Parse(resp.Header.Get("Location"))
		req.Body = nil

		if err != nil {
			inst.Logger.Error("Error parsing redirect URL", zap.Error(err))
		}
		resp, err := inst.Session.Client.Do(req)
		if err != nil {
			inst.Logger.Error("Error checking checkout progress", zap.Error(err))
		}

		if strings.Contains(resp.Header.Get("Location"), "/thank_you") {
			inst.Logger.Info("Successfully Checked Out!!")
			success = true
		}
		if strings.Contains(resp.Header.Get("Location"), "/processing") {
			inst.Logger.Info("Processing payment")
		}

		switch resp.StatusCode {
		case 302:
			inst.Logger.Info("Payment request status code", zap.String("Payment request status code", strconv.Itoa(resp.StatusCode)), zap.String("Response body", respStr))
			break
		case 200:
			inst.Logger.Info("Payment request status code", zap.String("Payment request status code", strconv.Itoa(resp.StatusCode)), zap.String("Response body", respStr))
			break
		}
	}

	fmt.Println("3DS URL: ", resp.Header.Get("Location"))

	return true, nil
}

func (inst *Instance) printStatus(text string) {
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

func (inst *Instance) run() {
	for true {
		inst.Status = "Getting variants"
		_, err := inst.wrap(inst.getVariants)
		if err != nil {
			continue
		}

		inst.Status = "Carting variants"
		_, err = inst.wrap(inst.cartVariant)
		if err != nil {
			continue
		}

		inst.Status = "Initializing checkout"
		_, err = inst.wrap(inst.initCheckout)
		if err != nil {
			continue
		}

		inst.Status = "Getting authorization token"
		_, err = inst.wrap(inst.authToken)
		if err != nil {
			continue
		}

		inst.Status = "Submitting address"
		_, err = inst.wrap(inst.submitAddress)
		if err != nil {
			continue
		}

		inst.Status = "Getting delivery token"
		_, err = inst.wrap(inst.deliveryToken)
		if err != nil {
			continue
		}

		inst.Status = "Getting shipping rates"
		_, err = inst.wrap(inst.getShippingRates)
		if err != nil {
			continue
		}

		inst.Status = "Submitting delivery"
		_, err = inst.wrap(inst.submitDelivery)
		if err != nil {
			continue
		}

		inst.Status = "Getting gateway"
		_, err = inst.wrap(inst.getGateway)
		if err != nil {
			continue
		}

		inst.Status = "Creating payment session"
		_, err = inst.wrap(inst.createPaymentSession)
		if err != nil {
			continue
		}

		inst.Status = "Submitting payment"
		_, err = inst.wrap(inst.submitPayment)
		if err != nil {
			continue
		}

	}
}

func Test() {
	//bd-eu.porterproxies.com:8888::4p61al0m
	//bd-eu.porterproxies.com:8888:user-PP_A46I34N-country-gb-plan-luminati-session-77201727:4p61al0m
	//bd-eu.porterproxies.com:8888:user-PP_A46I34N-country-gb-plan-luminati-session-47144407:4p61al0m
	//bd-eu.porterproxies.com:8888:user-PP_A46I34N-country-gb-plan-luminati-session-62726645:4p61al0m
	//bd-eu.porterproxies.com:8888:user-PP_A46I34N-country-gb-plan-luminati-session-65273914:4p61al0m
	//bd-eu.porterproxies.com:8888:user-PP_A46I34N-country-gb-plan-luminati-session-13220225:4p61al0m
	//bd-eu.porterproxies.com:8888:user-PP_A46I34N-country-gb-plan-luminati-session-25219854:4p61al0m
	//bd-eu.porterproxies.com:8888:user-PP_A46I34N-country-gb-plan-luminati-session-94389856:4p61al0m
	//bd-eu.porterproxies.com:8888:user-PP_A46I34N-country-gb-plan-luminati-session-36275924:4p61al0m
	//bd-eu.porterproxies.com:8888:user-PP_A46I34N-country-gb-plan-luminati-session-49296777:4p61al0m
	//bd-eu.porterproxies.com:8888:user-PP_A46I34N-country-gb-plan-luminati-session-54498097:4p61al0m
	prof := data_handling.Dylan()
	opt := data_handling.Options{
		TaskID:    0,
		URL:       "https://releases.flatspot.com/products/nike-sb-x-albino-preto-dunk-low-pro-shoes-fossil-black-sail",
		VariantID: "40747855872061",
		UseProxy:  false,
		Proxy: data_handling.ProxyDefiniton{
			Protocol: "http",
			Host:     "bd-eu.porterproxies.com",
			Port:     "8888",
			Username: "user-PP_A46I34N-country-gb-plan-luminati-session-31382305",
			Password: "4p61al0m",
		},
		Profile: prof,
		Size:    "UK 10",
	}

	inst, err := NewShopifyInstance(
		opt,
	)

	if err != nil {
	}

	inst.run()
	fmt.Print("INST: %v", inst)
}

type fn func()

// Creates wrapper function and sets it to the passed pointer to function
func (inst *Instance) wrap(function func() (bool, error)) (bool, error) {
	fmt.Println(inst.Status)
	return function()
}
