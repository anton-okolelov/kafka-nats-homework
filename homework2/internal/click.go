package internal

type Click struct {
	SiteId     int32  `json:"site_id"`
	PriceCents int32  `json:"price_cents"`
	ClickedAt  string `json:"clicked_at"`
}
