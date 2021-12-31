package resp

type Dict struct {
	Base
	Name      string     `json:"name"`
	Desc      string     `json:"desc"`
	Status    uint       `json:"status"`
	DictDatas []DictData `json:"dictDatas"`
}

type DictData struct {
	Base
	Key      string `json:"key"`
	Val      string `json:"val"`
	Addition string `json:"addition"`
	Sort     uint   `json:"sort"`
	Status   uint   `json:"status"`
	DictId   uint   `json:"dictId"`
	Dict     Dict   `json:"dict"`
}
