package data

import "time"

//
type EVar struct {
	AppID     string    `json:"app_id"`     //
	CreatedAt time.Time `json:"created_at"` //
	ID        string    `json:"id"`         //
	Internal  bool      `json:"internal"`   //
	ServiceID string    `json:"service_id"` //
	Title     string    `json:"title"`      //
	UpdatedAt time.Time `json:"updated_at"` //
	Value     string    `json:"value"`      //
}

//
func (e *EVar) Collection() string {
	return "evars"
}

//
func (e *EVar) Id() string {
	return e.ID
}
