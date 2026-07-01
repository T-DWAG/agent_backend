package model

// SubscriptionPlan 是订阅计划类型，定义了不同的订阅级别
type SubscriptionPlan string

const (
	FreePlan       SubscriptionPlan = "free"       // 免费版
	BasicPlan      SubscriptionPlan = "basic"      // 基础版
	ProPlan        SubscriptionPlan = "pro"        // 专业版
	EnterprisePlan SubscriptionPlan = "enterprise" // 企业版
)

// PlanConfig 结构体，定义每种订阅计划对应的资源配额
type PlanConfig struct {
	MaxAgents            int64 `json:"maxAgents"`            // 最大可创建Agent数量
	MaxWorkflows         int64 `json:"maxWorkflows"`         // 最大可创建工作流数量
	MaxKnowledgeBaseSize int64 `json:"maxKnowledgeBaseSize"` // 知识库最大容量
}

// PaymentDuration 是付款周期类型，定义了支持的计费周期
type PaymentDuration string

const (
	Monthly   PaymentDuration = "month"   // 按月付费
	Quarterly PaymentDuration = "quarter" // 按季度付费
	Yearly    PaymentDuration = "year"    // 按年付费
)

// PaymentMethod 是支付方式类型，目前仅支持微信支付
type PaymentMethod string

const (
	WeChatPay PaymentMethod = "wechat" // 微信支付
)
