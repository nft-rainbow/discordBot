package models

type UploadFilesResponse struct {
	FileUrl  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
	FileName string `json:"file_name"`
}
