package resp

type Api struct {
	Base
	Method   string `json:"method"`
	Path     string `json:"path"`
	Category string `json:"category"`
	Desc     string `json:"desc"`
	Title    string `json:"title"`
}

type ApiGroupByCategory struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Children []Api  `json:"children"`
}

type ApiTreeWithAccess struct {
	List      []ApiGroupByCategory `json:"list"`
	AccessIds []uint               `json:"accessIds"`
}
