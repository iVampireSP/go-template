package auth

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"
	"go-template/internal/service/jwks"
	"gorm.io/gorm"
)

type Service struct {
	config   *conf.Config
	jwks     *jwks.JWKS
	logger   *logger.Logger
	Enforcer *casbin.Enforcer
}

func NewService(
	config *conf.Config,
	jwks *jwks.JWKS,
	logger *logger.Logger,
	db *gorm.DB,
) *Service {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		panic(err)
	}

	casbinModel, err := model.NewModelFromString(config.GetRBACModel())
	if err != nil {
		panic(err)
	}

	enforcer, err := casbin.NewEnforcer(casbinModel, adapter)
	if err != nil {
		panic(err)
	}

	enforcer.EnableAutoSave(true)
	enforcer.EnableAutoNotifyWatcher(true)

	return &Service{
		config:   config,
		jwks:     jwks,
		logger:   logger,
		Enforcer: enforcer,
	}
}
