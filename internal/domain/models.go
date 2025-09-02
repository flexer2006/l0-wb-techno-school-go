package domain

import "time"

type Order struct {
	DateCreated time.Time `json:"date_created" db:"date_created"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Raw         []byte    `json:"-" db:"raw"`

	Delivery *Delivery `json:"delivery"`
	Payment  *Payment  `json:"payment"`
	Items    []Item    `json:"items"`

	OrderUID          string `json:"order_uid" db:"order_uid"`
	TrackNumber       string `json:"track_number" db:"track_number"`
	Entry             string `json:"entry" db:"entry"`
	Locale            string `json:"locale" db:"locale"`
	InternalSignature string `json:"internal_signature" db:"internal_signature"`
	CustomerID        string `json:"customer_id" db:"customer_id"`
	DeliveryService   string `json:"delivery_service" db:"delivery_service"`
	Shardkey          string `json:"shardkey" db:"shardkey"`
	OofShard          string `json:"oof_shard" db:"oof_shard"`

	SmID int `json:"sm_id" db:"sm_id"`
}

type Delivery struct {
	ID int64 `json:"-" db:"id"`

	OrderUID string `json:"-" db:"order_uid"`
	Name     string `json:"name" db:"name"`
	Phone    string `json:"phone" db:"phone"`
	Zip      string `json:"zip" db:"zip"`
	City     string `json:"city" db:"city"`
	Address  string `json:"address" db:"address"`
	Region   string `json:"region" db:"region"`
	Email    string `json:"email" db:"email"`
}

type Payment struct {
	ID           int64     `json:"-" db:"id"`
	PaymentDt    int64     `json:"payment_dt" db:"payment_dt"`
	PaymentTs    time.Time `json:"-" db:"payment_ts"`
	Amount       float64   `json:"amount" db:"amount"`
	DeliveryCost float64   `json:"delivery_cost" db:"delivery_cost"`
	GoodsTotal   float64   `json:"goods_total" db:"goods_total"`
	CustomFee    float64   `json:"custom_fee" db:"custom_fee"`

	OrderUID    string `json:"-" db:"order_uid"`
	Transaction string `json:"transaction" db:"transaction"`
	RequestID   string `json:"request_id" db:"request_id"`
	Currency    string `json:"currency" db:"currency"`
	Provider    string `json:"provider" db:"provider"`
	Bank        string `json:"bank" db:"bank"`
}

type Item struct {
	ID         int64   `json:"-" db:"id"`
	ChrtID     int64   `json:"chrt_id" db:"chrt_id"`
	NmID       int64   `json:"nm_id" db:"nm_id"`
	Price      float64 `json:"price" db:"price"`
	TotalPrice float64 `json:"total_price" db:"total_price"`

	OrderUID    string `json:"-" db:"order_uid"`
	TrackNumber string `json:"track_number" db:"track_number"`
	RID         string `json:"rid" db:"rid"`
	Name        string `json:"name" db:"name"`
	Size        string `json:"size" db:"size"`
	Brand       string `json:"brand" db:"brand"`

	Sale   int `json:"sale" db:"sale"`
	Status int `json:"status" db:"status"`
}
