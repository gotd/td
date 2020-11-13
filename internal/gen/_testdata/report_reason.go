package _testdata

/*
inputReportReasonSpam#58dbcab8 = ReportReason;
inputReportReasonViolence#1e22c78d = ReportReason;
inputReportReasonPornography#2e59d922 = ReportReason;
inputReportReasonChildAbuse#adf44ee3 = ReportReason;
inputReportReasonOther#e1746d0a text:string = ReportReason;
inputReportReasonCopyright#9b89f93a = ReportReason;
inputReportReasonGeoIrrelevant#dbd4feed = ReportReason;
*/

type ReportReason interface {
	reportReason()
}

type InputReportReasonSpam struct{}

func (InputReportReasonSpam) reportReason() {}

func (InputReportReasonSpam) encode(raw []byte) []byte {
	return raw
}

type InputReportReasonOther struct {
	Text string
}

func (InputReportReasonOther) reportReason() {}
