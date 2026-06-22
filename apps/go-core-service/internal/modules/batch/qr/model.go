package qr

type ProductItemLabel struct {
	ItemCode string
	Token    string
}

type BatchPDFInput struct {
	BatchCode string
	Items     []ProductItemLabel
}
