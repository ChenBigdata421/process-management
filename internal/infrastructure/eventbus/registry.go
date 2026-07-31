package eventbus

import (
	"sort"
)

// Subscription 描述一个 topic 声明(v4 动作1:process-management 侧手列表 → 1 注册表)。
type Subscription struct {
	Topic string
}

// Registry 收集本服务全部 topic 声明,派生出预订阅 / 存在性断言列表。
type Registry struct {
	subs []Subscription
}

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Register(topic string) {
	r.subs = append(r.subs, Subscription{Topic: topic})
}

// Topics 去重后的全部 topic —— 预订阅 / 存在性断言对象。
// process-management 为纯发布(不消费),无 consumed 概念;原 Owned 字段已删(配置集中化后无区分意义)。
func (r *Registry) Topics() []string {
	return r.distinct(func(s Subscription) bool { return true })
}

func (r *Registry) distinct(pred func(Subscription) bool) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, s := range r.subs {
		if pred(s) {
			if _, ok := seen[s.Topic]; !ok {
				seen[s.Topic] = struct{}{}
				out = append(out, s.Topic)
			}
		}
	}
	sort.Strings(out)
	return out
}
