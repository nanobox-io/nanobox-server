package models

type (

	//
	Model interface {
		Get(id string)
		Create(body []byte)
		Update(id string, body []byte)
		Destroy(id string)
		Save()
		Count()
	}
)
