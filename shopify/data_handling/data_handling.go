package data_handling

type ProxyDefiniton struct {
	Protocol string
	Host     string
	Port     string
	Username string
	Password string
}

type Options struct {
	TaskID    int
	URL       string
	VariantID string
	UseProxy  bool
	Proxy     ProxyDefiniton
	Profile   CheckoutProfile
	Size      string
}

type CardDetails struct {
	Number            string      `json:"number"`
	Name              string      `json:"name"`
	StartMonth        interface{} `json:"start_month"`
	StartYear         interface{} `json:"start_year"`
	Month             string      `json:"month"`
	Year              string      `json:"year"`
	VerificationValue string      `json:"verification_value"`
	IssueNumber       string      `json:"issue_number"`
}

type CheckoutProfile struct {
	Email    string
	Country  string
	Fname    string
	Lname    string
	Address1 string
	Address2 string
	City     string
	Zipcode  string
	Phone    string
	Card     CardDetails
}

func NewProfile() CheckoutProfile {
	cc := CardDetails{
		Number:            "5375 9000 3932 8312",
		Name:              "MR O J URBANIAK",
		StartMonth:        nil,
		StartYear:         nil,
		Month:             "05",
		Year:              "2025",
		VerificationValue: "315",
		IssueNumber:       "",
	}
	//106 originally

	return CheckoutProfile{
		Email:    "markbania111@gmail.com",
		Country:  "United Kingdom",
		Fname:    "Marek",
		Lname:    "Urbaniak",
		Address1: "36 Gloucester Road",
		Address2: "",
		City:     "Shrewsbury",
		Zipcode:  "SY13PJ",
		Phone:    "7928983220",
		Card:     cc,
	}
}

func Dylan() CheckoutProfile {
	cc := CardDetails{
		Number:            "5356 7434 8171 8942",
		Name:              "OJ URBAN",
		StartMonth:        nil,
		StartYear:         nil,
		Month:             "08",
		Year:              "2032",
		VerificationValue: "450",
		IssueNumber:       "",
	}
	//106 originally

	return CheckoutProfile{
		Email:    "dylww659@gmail.com",
		Country:  "United Kingdom",
		Fname:    "Dylan",
		Lname:    "Williams",
		Address1: "11 Milner Road",
		Address2: "",
		City:     "Liverpool",
		Zipcode:  "L170AB",
		Phone:    "7928983267",
		Card:     cc,
	}
}

func Katarzyna() CheckoutProfile {
	cc := CardDetails{
		Number:            "5356 7431 5170 0766",
		Name:              "OSKAR J URBANIAK",
		StartMonth:        nil,
		StartYear:         nil,
		Month:             "08",
		Year:              "2032",
		VerificationValue: "106",
		IssueNumber:       "",
	}
	//106 originally

	return CheckoutProfile{
		Email:    "katarzyna320193gogo@gmail.com",
		Country:  "United Kingdom",
		Fname:    "Katarzyna",
		Lname:    "Michalak",
		Address1: "19 Stersacre Road",
		Address2: "",
		City:     "Shrewsbury",
		Zipcode:  "SY13PW",
		Phone:    "7928983330",
		Card:     cc,
	}
}

//func (p CheckoutProfile) CardJson() string {
//	return `{"credit_card":{"number":"` + p.CC_NUM + `","name":"` + p.CC_NAME + `","month":"` + p.CC_EXP_M + `","year":"` + p.CC_EXP_YYYY + `","verification_value":"` + p.CC_CVV + `"}}`
//	return p.Email
//}

//# TestProf = Profile(
//#     email="worky.wor.k899@gmail.com",
//#     country="United Kingdom",
//#     fname="DYLAN",
//#     lname="Williams",
//#     address1="11 Milner Road",
//#     address2="",
//#     city="Liverpool",
//#     zipcode="L170AB",
//#     phone="07928988776",
//#     ccnum="5354 5680 7405 8238",
//#     ccexp_m="8",
//#     ccexp_yyyy="2028",
//#     cvv="366",
//#     ccname="DYLAN WILLIAMS",
//# )
//# route_one_instance = Shopify(
//#     "https://launches.routeone.co.uk/products/nike-sb-yuto-dunk-low-pro-qs-skate-shoes-wolf-grey-iron-grey-sail",
//#     profile=TestProf,
//#     variant_id=42881781858475,
//# )
//# route_one_instance.test_cart()
