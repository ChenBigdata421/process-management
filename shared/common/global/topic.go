package global

const (
	LoginLog   = "login_log_queue"
	OperateLog = "operate_log_queue"
	ApiCheck   = "api_check_queue"

	// 领域事件Topic
	TopicCaseEvents                  = "case.events"
	TopicMediaEvents                 = "media.events"
	TopicIncidentRecordEvents        = "incident.events"
	TopicArchiveEvents               = "archive.events"
	TopicWritEvents                  = "writ.events" // 文书事件Topic
	TopicCaseMediaRelationEvents     = "case.media.relation.events"
	TopicIncidentMediaRelationEvents = "incident.media.relation.events"
	TopicArchiveMediaRelationEvents  = "archive.media.relation.events"
	TopicWritMediaRelationEvents     = "writ.media.relation.events" // 文书媒体关联事件Topic
)
