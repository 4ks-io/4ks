package function

import "time"

// StorageObjectData contains metadata of the Cloud Storage object.
// https://stackoverflow.com/questions/65031928/how-to-upload-a-resized-image-in-google-cloud-storage
type StorageObjectData struct {
	Bucket         string    `json:"bucket,omitempty"`
	Name           string    `json:"name,omitempty"`
	Metageneration int64     `json:"metageneration,string,omitempty"`
	TimeCreated    time.Time `json:"timeCreated,omitempty"`
	Updated        time.Time `json:"updated,omitempty"`
}

// FileProps holds the parsed extension and basename of an uploaded file.
type FileProps struct {
	Extension string
	Basename  string
}

// MediaStatus represents the processing state of a media asset.
// Same as libs/go/models/recipe-media-status.go
type MediaStatus int

// MediaStatus constants.
const (
	MediaStatusRequested                  MediaStatus = 0
	MediaStatusProcessing                 MediaStatus = 1
	MediaStatusReady                      MediaStatus = 2
	MediaStatusErrorUnknown               MediaStatus = 60
	MediaStatusErrorSize                  MediaStatus = 61
	MediaStatusErrorMissingAttr           MediaStatus = 62
	MediaStatusErrorInvalidMIMEType       MediaStatus = 63
	MediaStatusErrorVision                MediaStatus = 64
	MediaStatusErrorSafeSearch            MediaStatus = 65
	MediaStatusErrorInappropriateAdult    MediaStatus = 70
	MediaStatusErrorInappropriateMedical  MediaStatus = 71
	MediaStatusErrorInappropriateViolence MediaStatus = 72
	MediaStatusErrorInappropriateRacy     MediaStatus = 73
	MediaStatusErrorFailedCopy            MediaStatus = 80
	MediaStatusErrorFailedResize          MediaStatus = 81
	MediaStatusErrorFailedVariant         MediaStatus = 82
)
