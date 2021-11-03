package resp

type UploadMerge struct {
	Filename   string `json:"filename"`
	PreviewUrl string `json:"previewUrl"`
}

type UploadUnZip struct {
	Files []string `json:"files"`
}
