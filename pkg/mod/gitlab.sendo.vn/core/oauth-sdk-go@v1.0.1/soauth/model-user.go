package soauth

type UserStruct struct {
	CustomerID uint64 `json:"customer_id" bson:"customer_id"`
	FPTID      uint64 `json:"fpt_id" bson:"fpt_id"`

	Email string `json:"email" bson:"email"`
	Phone string `json:"phone" bson:"telephone"`

	CheckoutVerified bool `json:"checkout_verified" bson:"checkout_verified"`
	EmailVerified    bool `json:"email_verified" bson:"email_verified"`
	PhoneVerified    bool `json:"phone_verified" bson:"phone_verified"`

	FirstName string `json:"first_name" bson:"first_name"`
	LastName  string `json:"last_name" bson:"last_name"`

	Avatar   string `json:"avatar" bson:"avatar"`
	Birthday uint64 `json:"birthday" bson:"dob"`
	Gender   uint64 `json:"gender" bson:"gender"`

	DefaultShipping uint64 `json:"default_shipping" bson:"default_shipping"`

	CreatedAt uint64 `json:"created_at" bson:"created_at"`
	UpdatedAt uint64 `json:"updated_at" bson:"updated_at"`
}
