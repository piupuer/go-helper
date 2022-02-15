package resp

type DelayExportHistory struct {
	Base
	Uuid     string `json:"uuid"`
	Category string `json:"category"`
	Name     string `json:"name"`
	Progress string `json:"progress"`
	End      uint   `json:"end"`
	Url      string `json:"url"`
}
