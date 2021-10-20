package main

import (
	"flag"
	"github.com/LF-Engineering/insights-datasource-shared/firehose"
)

var region string


func init() {
	flag.StringVar(&region, "AWS_REGION", "", "The firehose region name")
}

func main() {
	jiraChannel := "jira"
	flag.Parse()

	// create new firehose client provider
	// you need to provide region as environment variable, or it will fall to default
	// which is us-east-1
	client, err := firehose.NewClientProvider()
	if err != nil {
		panic(err)
	}

	// create new delivery stream channel named jira
	// you will need to create channel once, and then you can use it every time
	// to check if delivery stream is already exist you may use DescribeDeliveryStream
	err = client.CreateDeliveryStream(jiraChannel)
	if err != nil {
		panic(err)
	}

	// put single data record to jira channel
	_, err = client.PutRecord(jiraChannel, b)
	if err != nil {
		panic(err)
	}
}

var b = `{
  "DataSource": {
    "Name": "Jira",
    "Slug": "jira"
  },
  "Endpoint": "https://jira.lfnetworking.org",
  "Events": [
    {
      "Issue": {
        "Activities": [
          {
            "ActivityType": "jira_issue_created",
            "Body": "obfuscated issue body",
            "CreatedAt": "2021-03-16T13:26:53.000Z",
            "CreatedTz": "Asia/Shanghai",
            "Id": "98de051275eb6dc8f865a71b64ae513858fe13e0",
            "Identity": {
              "DataSourceId": "jira",
              "Email": "[redacted]",
              "Id": "96583b7dd6cd38a24f99834d973219318c0aee02",
              "Name": "Qihui Zhao",
              "Username": "qihuiz"
            },
            "IssueId": "10202",
            "IssueKey": "28e32b09599a4471168c4943aff57574ff8d309a"
          },
          {
            "ActivityType": "jira_issue_reporter_added",
            "Body": "obfuscated issue body",
            "CreatedAt": "2021-03-16T13:26:53.000Z",
            "CreatedTz": "Asia/Shanghai",
            "Id": "2a19fa9904eed651b07853c1d81edd3436171de6",
            "Identity": {
              "DataSourceId": "jira",
              "Email": "[redacted]",
              "Id": "96583b7dd6cd38a24f99834d973219318c0aee02",
              "Name": "Qihui Zhao",
              "Username": "qihuiz"
            },
            "IssueId": "10202",
            "IssueKey": "28e32b09599a4471168c4943aff57574ff8d309a"
          }
        ],
        "CreatedAt": "2021-03-16T13:26:53.000Z",
        "CreatedTz": "Asia/Shanghai",
        "DataSourceId": "jira",
        "Id": "28e32b09599a4471168c4943aff57574ff8d309a",
        "IssueId": "10202",
        "JiraProject": {
          "Id": "10200",
          "Key": "XGVELA",
          "Name": "XGVela"
        },
        "Labels": null,
        "Releases": null,
        "Title": "5G slicing",
        "URL": "https://jira.lfnetworking.org/browse/XGVELA-2",
        "UpdatedAt": "2021-03-18T08:31:57.000Z",
        "Watchers": 1
      }
    },
    {
      "Issue": {
        "Activities": [
          {
            "ActivityType": "jira_issue_created",
            "Body": "obfuscated issue body",
            "CreatedAt": "2021-03-16T08:22:15.000Z",
            "CreatedTz": "Asia/Shanghai",
            "Id": "0a2e08031585dfdfc87e639fac684c53a4318212",
            "Identity": {
              "DataSourceId": "jira",
              "Email": "[redacted]",
              "Id": "96583b7dd6cd38a24f99834d973219318c0aee02",
              "Name": "Qihui Zhao",
              "Username": "qihuiz"
            },
            "IssueId": "10201",
            "IssueKey": "46ab9541fcbe2a73aa843f8f7e8651ac47f3dd8c"
          },
          {
            "ActivityType": "jira_issue_reporter_added",
            "Body": "obfuscated issue body",
            "CreatedAt": "2021-03-16T08:22:15.000Z",
            "CreatedTz": "Asia/Shanghai",
            "Id": "d1ed9f45272aad68bdd0a7a1de0d20f54624aedc",
            "Identity": {
              "DataSourceId": "jira",
              "Email": "[redacted]",
              "Id": "96583b7dd6cd38a24f99834d973219318c0aee02",
              "Name": "Qihui Zhao",
              "Username": "qihuiz"
            },
            "IssueId": "10201",
            "IssueKey": "46ab9541fcbe2a73aa843f8f7e8651ac47f3dd8c"
          }
        ],
        "CreatedAt": "2021-03-16T08:22:15.000Z",
        "CreatedTz": "Asia/Shanghai",
        "DataSourceId": "jira",
        "Id": "46ab9541fcbe2a73aa843f8f7e8651ac47f3dd8c",
        "IssueId": "10201",
        "JiraProject": {
          "Id": "10200",
          "Key": "XGVELA",
          "Name": "XGVela"
        },
        "Labels": null,
        "Releases": null,
        "Title": "Private 5G",
        "URL": "https://jira.lfnetworking.org/browse/XGVELA-1",
        "UpdatedAt": "2021-03-17T02:54:00.000Z",
        "Watchers": 1
      }
    },
    {
      "Issue": {
        "Activities": [
          {
            "ActivityType": "jira_issue_created",
            "Body": "obfuscated issue body",
            "CreatedAt": "2021-03-18T06:25:45.000Z",
            "CreatedTz": "Asia/Shanghai",
            "Id": "ad6c5c2a19c1021d38e64985b6b3a79636b37078",
            "Identity": {
              "DataSourceId": "jira",
              "Email": "[redacted]",
              "Id": "96583b7dd6cd38a24f99834d973219318c0aee02",
              "Name": "Qihui Zhao",
              "Username": "qihuiz"
            },
            "IssueId": "10203",
            "IssueKey": "131bb98c3651aa12c14eebce9fc39b81de8450ff"
          },
          {
            "ActivityType": "jira_issue_reporter_added",
            "Body": "obfuscated issue body",
            "CreatedAt": "2021-03-18T06:25:45.000Z",
            "CreatedTz": "Asia/Shanghai",
            "Id": "ab43ac9df8d0b44591aa8127f4453be4054e775e",
            "Identity": {
              "DataSourceId": "jira",
              "Email": "[redacted]",
              "Id": "96583b7dd6cd38a24f99834d973219318c0aee02",
              "Name": "Qihui Zhao",
              "Username": "qihuiz"
            },
            "IssueId": "10203",
            "IssueKey": "131bb98c3651aa12c14eebce9fc39b81de8450ff"
          },
          {
            "ActivityType": "jira_comment_created",
            "Body": "obfuscated comment body",
            "CreatedAt": "2021-03-23T08:00:09.000Z",
            "CreatedTz": "UTC",
            "Id": "af8f9619e1399ec1071865fbad1ead4c4e81f334",
            "Identity": {
              "DataSourceId": "jira",
              "Email": "[redacted]",
              "Id": "4fd2f3e841457938983ad1475fef8073d1b8d4fc",
              "Name": "Ulf Hallgarn",
              "Username": "ulfhallgarn"
            },
            "IssueId": "10203",
            "IssueKey": "131bb98c3651aa12c14eebce9fc39b81de8450ff"
          },
          {
            "ActivityType": "jira_comment_created",
            "Body": "obfuscated comment body",
            "CreatedAt": "2021-03-23T13:33:55.000Z",
            "CreatedTz": "Asia/Shanghai",
            "Id": "285f870e56ca618e2065479192e404d8510c2f9f",
            "Identity": {
              "DataSourceId": "jira",
              "Email": "[redacted]",
              "Id": "96583b7dd6cd38a24f99834d973219318c0aee02",
              "Name": "Qihui Zhao",
              "Username": "qihuiz"
            },
            "IssueId": "10203",
            "IssueKey": "131bb98c3651aa12c14eebce9fc39b81de8450ff"
          },
          {
            "ActivityType": "jira_comment_updated",
            "Body": "obfuscated comment body",
            "CreatedAt": "2021-03-25T08:49:56.000Z",
            "CreatedTz": "Asia/Shanghai",
            "Id": "3c79460d876d1111f5d34a1254eb1c0335e02110",
            "Identity": {
              "DataSourceId": "jira",
              "Email": "[redacted]",
              "Id": "96583b7dd6cd38a24f99834d973219318c0aee02",
              "Name": "Qihui Zhao",
              "Username": "qihuiz"
            },
            "IssueId": "10203",
            "IssueKey": "131bb98c3651aa12c14eebce9fc39b81de8450ff"
          },
          {
            "ActivityType": "jira_comment_created",
            "Body": "obfuscated comment body",
            "CreatedAt": "2021-04-22T15:12:47.000Z",
            "CreatedTz": "UTC",
            "Id": "ffcfc09b7533cac9be4979071fc7aeb51dedc955",
            "Identity": {
              "DataSourceId": "jira",
              "Email": "[redacted]",
              "Id": "1f89f7c4ce2a8d10e9193470d359176000bdb034",
              "Name": "Saad Ullah Sheikh",
              "Username": "SaadUllahSheikh"
            },
            "IssueId": "10203",
            "IssueKey": "131bb98c3651aa12c14eebce9fc39b81de8450ff"
          }
        ],
        "CreatedAt": "2021-03-18T06:25:45.000Z",
        "CreatedTz": "Asia/Shanghai",
        "DataSourceId": "jira",
        "Id": "131bb98c3651aa12c14eebce9fc39b81de8450ff",
        "IssueId": "10203",
        "JiraProject": {
          "Id": "10200",
          "Key": "XGVELA",
          "Name": "XGVela"
        },
        "Labels": [
          "Operation",
          "Topology"
        ],
        "Releases": null,
        "Title": "Resource locatoin & topology",
        "URL": "https://jira.lfnetworking.org/browse/XGVELA-3",
        "UpdatedAt": "2021-04-22T15:12:47.000Z",
        "UpdatedTz": "UTC",
        "Watchers": 3
      }
    }
  ],
  "MetaData": {
    "BackendName": "jira",
    "BackendVersion": "0.1.1",
    "Project": "ONAP",
    "Tags": [
      "e",
      "a",
      "b",
      "c",
      "d"
    ]
  }
}`
