package command

import (
	"go.uber.org/zap"

	config2 "paperback-vbook-converter/config"
	"paperback-vbook-converter/internal/pkg/http"
)

type CheckForUpdateSource struct {
	logger *zap.SugaredLogger
	config *config2.Config
}

var (
	CheckForUpdateCommand = "CheckForUpdate"
)

//type checkForUpdateWorker func(ctx context.Context, repositoryUrl string)

func NewCheckForUpdateCmd(logger *zap.SugaredLogger, config *config2.Config) *CheckForUpdateSource {
	return &CheckForUpdateSource{
		logger: logger,
		config: config,
	}
}

func (c CheckForUpdateSource) Run(repositories []string) error {

	c.logger.Infof("%s: Start check for update....", CheckForUpdateCommand)

	c.logger.Debug(repositories)

	if len(repositories) == 0 {
		c.logger.Warnf("%s: empty repository", CheckForUpdateCommand)
	}

	httpClient := http.NewHttpClient(c.logger, c.config)

	var err error

	for _, repository := range repositories {
		err := c.checkForUpdate(repository, httpClient)
		if err != nil {
			return err
		}
	}

	return err
}

func (c CheckForUpdateSource) checkForUpdate(repositoryUrl string, httpClient *http.ClientRequestHandler) error {

	repositoryInfo := httpClient.GetRepositoryInfo(repositoryUrl)

	c.logger.Debug(repositoryInfo.Metadata, repositoryInfo.Sources)

	//for len(repositoryInfo.Sources) > 0 {
	//	ctxCancelFn()
	//}
	//
	//ctxCancelFn()

	return nil
}
