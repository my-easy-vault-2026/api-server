package test

import (
	"api-server/services"
	"shared-modules/entities"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	cardService    *services.CardService
	accountService *services.AccountService
}

func NewCardHandler() *CardHandler {
	return &CardHandler{
		cardService:    services.NewCardService(),
		accountService: services.NewAccountService(),
	}
}

// @Param			request		body		entities.ListCardForm	true	"body"
// @Param			X-ApiKey	header		string					true	"Api key"
// @Success		0			{object}	entities.ListCardVO		"data"
// @Router			/merchant/card/list [post]
// @Description	List merchant cards.
// @Tags			merchant/card
func (ch *CardHandler) ListCard(c *gin.Context) {

	form := &entities.ListCardForm{}

	err := c.ShouldBindJSON(form)

	if err != nil {
		utils.ReError(c, err)
		return
	}

	cards, err := ch.cardService.ListCard(c, form, 0, 0)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	if len(cards) == 0 {
		utils.ReData(
			c,
			nil,
		)
	}

	utils.ReData(
		c,
		cards[0].IssueID,
	)
}
