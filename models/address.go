package models

type BindCFXAddress struct {
	UserId string `gorm:"type:varchar(256)" json:"user_id" binding:"required"`
	UserAddress string `gorm:"type:varchar(256)" json:"user_address" binding:"required"`
}

type GetBindCFXAddressResp struct{
	CFXAddress string `json:"cfx_address"`
	UserId string `json:"user_id"`
}



