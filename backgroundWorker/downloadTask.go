package backgroundWorker

import "encoding/json"

type DownloadTask struct {
	Hash string `json:"hash"`
}

func (dt *DownloadTask) ToJSON() string {
	str, _ := json.Marshal(dt)
	return string(str)
}
