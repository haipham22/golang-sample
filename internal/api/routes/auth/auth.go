package auth

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"golang-sample/internal/api/storage"
)

type Auth interface {

	// PostLogin godoc
	//
	//	@Summary		Login
	//	@Description	Login to get access token
	//	@Tags			auth
	//	@Accept			json
	//	@Produce		json
	//	@Param			campaign_id	path		int	true	"Campaign ID"
	//	@Success		200			{object}	schemas.Response{data=models.CampaignData}
	//	@Router			/campaign/{campaign_id} [get]
	PostLogin(c echo.Context) error
	PostRegister(c echo.Context) error
}

type authHandler struct {
	log     *zap.SugaredLogger
	storage storage.Storage
}

func NewAuthHandler(log *zap.SugaredLogger, storage storage.Storage) Auth {
	return &authHandler{
		log:     log,
		storage: storage,
	}
}
