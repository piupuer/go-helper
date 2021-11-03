package resp

type Dict struct {
	Base
	Name      string     `json:"name"`
	Desc      string     `json:"desc"`
	Status    uint       `json:"status"`
	Remark    string     `json:"remark"`
	DictDatas []DictData `json:"dictDatas"`
}

type DictData struct {
	Base
	Key      string `json:"key"`
	Val      string `json:"val"`
	Attr     string `json:"attr"`
	Addition string `json:"addition"`
	Sort     uint   `json:"sort"`
	Status   uint   `json:"status"`
	Remark   string `json:"remark"`
	DictId   uint   `json:"dictId"`
	Dict     Dict   `json:"dict"`
}
