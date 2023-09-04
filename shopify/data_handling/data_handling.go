package data_handling

type CheckoutProfile struct {
	Email       string
	Country     string
	Fname       string
	Lname       string
	Address1    string
	Address2    string
	City        string
	Zipcode     string
	Phone       string
	CC_NUM      string
	CC_EXP_M    string
	CC_EXP_YYYY string
	CC_CVV      string
	CC_NAME     string
}

func NewProfile() CheckoutProfile {
	return CheckoutProfile{
		Email:       "workywork899@gmail.com",
		Country:     "United Kingdom",
		Fname:       "Oskar",
		Lname:       "Urbaniak",
		Address1:    "16 Winchester Square",
		Address2:    "",
		City:        "Chester",
		Zipcode:     "CH48NN",
		Phone:       "07928988720",
		CC_NUM:      "123",
		CC_EXP_M:    "456",
		CC_EXP_YYYY: "789",
		CC_CVV:      "123",
		CC_NAME:     "456",
	}
}

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
