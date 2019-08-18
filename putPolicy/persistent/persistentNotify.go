package persistent

type NotifyRequestBody struct {
	Hash    string   `json:"hash"`    // 原始文件的 CID
	Results []Result `json:"results"` // 持久化的结果
}

type Result struct {
	Code           int    `json:"code"`         // 状态码。0 表示成功；1 表示失败
	Desc           string `json:"desc"`         // 状态对应的描述
	PersistentOp   string `json:"persistentOp"` // 持久化操作的名称
	DstHash        string `json:"dstHash"`      // 生成的目标文件的 CID
	outputFilePath string `json:"-"`            // 临时文件输出路径
}
