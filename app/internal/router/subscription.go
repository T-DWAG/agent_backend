package router

import (
	"app/internal/subscriptions"

	"github.com/gin-gonic/gin"
)

type SubscriptionRouter struct{}

func (s *SubscriptionRouter) Register(engine *gin.Engine) {
	h := subscriptions.NewHandler()
	g := engine.Group("/api/v1/subscription")
	{
		g.GET("/current", h.GetUserSubscription)
	}
}
