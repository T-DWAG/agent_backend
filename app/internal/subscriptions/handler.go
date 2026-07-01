package subscriptions

import (
	"model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mszlu521/thunder/res"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetUserSubscription(c *gin.Context) {
	res.Success(c, &SubscriptionResponse{
		Configs: &model.PlanConfig{
			MaxAgents:            10,
			MaxKnowledgeBaseSize: 10,
			MaxWorkflows:         10,
		},
		ID:            uuid.New(),
		UserID:        uuid.New(),
		Plan:          string(model.FreePlan),
		Duration:      string(model.Yearly),
		PaymentMethod: string(model.WeChatPay),
		StartDate:     time.Now().Format(time.DateTime),
		EndDate:       time.Now().Add(365 * 24 * time.Hour).Format(time.DateTime),
		CreatedAt:     time.Now().Format(time.DateTime),
		UpdatedAt:     time.Now().Format(time.DateTime),
	})
}
