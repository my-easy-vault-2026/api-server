package entities

type DownloadEtherfiReportOptionVO struct {
	Account []*DownloadEtherfiReportOptionAccount `json:"account"`
}

type DownloadEtherfiReportOptionAccount struct {
	Id    uint64 `json:"id"`
	Value string `json:"value"`
}
