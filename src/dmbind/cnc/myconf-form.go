package cnc

var listQuery = `{
	"recordType":"object",
	"pageInfo":{"pageSize":%d,"pageNum":%d,"totalRowNum":-1,"totalPageNum":1,"startRowNum":%d,"endRowNum":%d},
	"columnInfo":[
		{"id":"isSelect","header":"选择","fieldName":"domainId","fieldIndex":"domainId","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"domainName","header":"域名","fieldName":"domainName","fieldIndex":"domainName","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"domainId","header":"domainId","fieldName":"domainId","fieldIndex":"domainId","sortOrder":null,"hidden":true,"exportable":true,"printable":true}
	],
	"sortInfo":[],"filterInfo":[],"remotePaging":true,"parameters":{},"action":"load"
}`

var versionListForm = `{
	"recordType":"object",
	"pageInfo": {"pageSize":5000,"pageNum":1,"totalRowNum":-1,"totalPageNum":1,"startRowNum":1,"endRowNum":0},
	"columnInfo":[
		{"id":"customerDetailId","header":"customerDetailId","fieldName":"customerDetailId","fieldIndex":"customerDetailId","sortOrder":null,"hidden":true,"exportable":true,"printable":true},
		{"id":"submitId","header":"submitId","fieldName":"submitId","fieldIndex":"submitId","sortOrder":null,"hidden":true,"exportable":true,"printable":true},
		{"id":"emails","header":"emails","fieldName":"emails","fieldIndex":"emails","sortOrder":null,"hidden":true,"exportable":true,"printable":true},
		{"id":"deployTime","header":"deployTime","fieldName":"deployTime","fieldIndex":"deployTime","sortOrder":null,"hidden":true,"exportable":true,"printable":true},
		{"id":"requirementId","header":"requirementId","fieldName":"requirementId","fieldIndex":"requirementId","sortOrder":null,"hidden":true,"exportable":true,"printable":true},
		{"id":"relevanceDomain","header":"relevanceDomain","fieldName":"relevanceDomain","fieldIndex":"relevanceDomain","sortOrder":null,"hidden":true,"exportable":true,"printable":true},
		{"id":"remark","header":"remark","fieldName":"remark","fieldIndex":"remark","sortOrder":null,"hidden":true,"exportable":true,"printable":true},
		{"id":"version","header":"版本号","fieldName":"version","fieldIndex":"version","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"status","header":"任务状态","fieldName":"status","fieldIndex":"status","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"submitName","header":"提交人","fieldName":"submitName","fieldIndex":"submitName","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"startTime","header":"提交时间","fieldName":"startTime","fieldIndex":"startTime","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"endTime","header":"完成时间","fieldName":"endTime","fieldIndex":"endTime","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"op","header":"操作","fieldName":"op","fieldIndex":"op","sortOrder":null,"hidden":false,"exportable":true,"printable":true}
	],
	"sortInfo":[],"filterInfo":[],"remotePaging":true,"parameters":{},"action":"load"
}`

var versionDetailForm = `{
	"recordType":"object",
	"pageInfo":{"pageSize":5000,"pageNum":1,"totalRowNum":-1,"totalPageNum":1,"startRowNum":0,"endRowNum":0},
	"columnInfo":[
		{"id":"icp","header":"备案号","fieldName":"icp","fieldIndex":"icp","sortOrder":null,"hidden":true,"exportable":true,"printable":true},
		{"id":"customerVersion","header":"customerVersion","fieldName":"customerVersion","fieldIndex":"customerVersion","sortOrder":null,"hidden":true,"exportable":true,"printable":true},
		{"id":"domainName","header":"域名","fieldName":"domainName","fieldIndex":"domainName","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"customerDomainDetailId","header":"服务域名","fieldName":"customerDomainDetailId","fieldIndex":"customerDomainDetailId","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"srcIp","header":"回源域名/源IP","fieldName":"srcIp","fieldIndex":"srcIp","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"clientIp","header":"客户端IP","fieldName":"clientIp","fieldIndex":"clientIp","sortOrder":null,"hidden":false,"exportable":true,"printable":true},
		{"id":"testUrl","header":"检测URL","fieldName":"testUrl","fieldIndex":"testUrl","sortOrder":null,"hidden":false,"exportable":true,"printable":true}
	],"sortInfo":[],"filterInfo":[],"remotePaging":true,"parameters":{},"action":"load"}
`
