package entities

import (
	"shared-modules/common"
	"time"
)

type FileDownload struct {
	Token string `json:"token" validate:"required"`
}

type FileDownloadCache struct {
	Type string `json:"type"`
	Id   uint64 `json:"id"`
}

type UploadFileToS3ByPipeForm struct {
	FileName    string             `json:"fileName" validate:"required"`
	Extention   string             `json:"extention" validate:"required"`
	FilePath    string             `json:"filePath" validate:"required"`
	MimeType    string             `json:"mimeType" validate:"required"`
	ContentType string             `json:"contentType" validate:"required"`
	FileType    common.FileType    `json:"fileType" validate:"required"`
	FilePurpose common.FilePurpose `json:"filePurpose" validate:"required"`
	ExpiredAt   *time.Time         `json:"expiredAt"`
	Expires     *time.Duration     `json:"expires"`
}
