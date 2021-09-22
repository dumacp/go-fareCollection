package mediums

type PaymentMedium struct {
	ID             string        `json:"id"`
	Active         bool          `json:"active"`
	Name           string        `json:"name"`
	Code           string        `json:"code"`
	Mappings       []interface{} `json:"mappings"`
	OrganizationID string        `json:"organizationId"`
	Metadata       *Metadata     `json:"metadata"`
}

type Metadata struct {
	CreatedBy string `json:"createdBy"`
	CreatedAt int    `json:"createAt"`
	UpdateBy  string `json:"updatedBy"`
	UpdateAt  int    `json:"updatedAt"`
}
