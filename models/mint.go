package models

import (
	"gorm.io/gorm"
	"math/big"
	"time"
)

type EasyMintMetaDto struct {
	Chain         string `form:"chain" json:"chain" binding:"required" oneof:"conflux conflux_test"`
	Name          string `form:"name" json:"name" binding:"required"`
	Description   string `form:"description" json:"description" binding:"required"`
	FileUrl       string `form:"file_url" json:"file_url" binding:"required,uri"`
	MintToAddress string `form:"mint_to_address" json:"mint_to_address" binding:"required"`
}

type CustomMintDto struct {
	ContractInfoDto
	MintItemDto
}

type ContractInfoDto struct {
	Chain           string `form:"chain" json:"chain" binding:"required,oneof=conflux conflux_test"`
	ContractType    string `form:"contract_type" json:"contract_type" binding:"required,oneof=erc721 erc1155" `
	ContractAddress string `form:"contract_address" json:"contract_address" binding:"required"`
}

type MintItemDto struct {
	MintToAddress string   `form:"mint_to_address" json:"mint_to_address" binding:"required"`
	TokenId       *big.Int `form:"token_id" json:"token_id"`
	Amount        *big.Int `form:"amount" json:"amount"`
	MetadataUri   string   `form:"metadata_uri" json:"metadata_uri" binding:"required,uri"`
}

type MintResp struct {
	UserAddress string `form:"user_address" json:"user_address"`
	NFTAddress string `form:"nft_address" json:"nft_address"`
	Advertise string `form:"advertise" json:"advertise"`
}


type MintTask struct {
	BaseModel
	AppId     uint   `gorm:"index" json:"app_id"`
	ChainType uint   `gorm:"type:int" json:"chain_type"`
	ChainId   uint   `gorm:"type:int" json:"chain_id"`
	Contract  string `gorm:"type:varchar(256);index" json:"contract"`
	MintTo    string `gorm:"type:varchar(256);index" json:"mint_to"`
	TokenURI  string `gorm:"type:varchar(256)" json:"token_uri"`
	TokenId   uint64 `gorm:"index" json:"token_id"`
	Amount    uint   `json:"amount"`
	Status    uint   `json:"status"` // 0-pending, 1-success, 2-failed
	Hash      string `gorm:"type:varchar(256)" json:"hash"`
	TxId      uint   `gorm:"index" json:"tx_id"`
	Error     string `gorm:"type:text" json:"error"`
}

type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}