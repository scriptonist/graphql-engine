package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/afero"

	"github.com/hasura/graphql-engine/cli/seed"

	"github.com/gin-gonic/gin"
)

// CreateSeedRequest --
type CreateSeedRequest struct {
	SQL      string `json:"sql,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// CreateSeedResponse --
type CreateSeedResponse struct {
	Message string `json:"message,omitempty"`
}

// CreateSeedAPIHandler --
func CreateSeedAPIHandler(c *gin.Context) {
	// bind request to struct var
	var requestData CreateSeedRequest
	if err := c.Bind(&requestData); err != nil {
		c.JSON(400, CreateSeedResponse{"bad request format"})
		return
	}
	if requestData.Filename == "" {
		c.JSON(400, CreateSeedResponse{"bad request format"})
		return
	}
	seedDirectory, ok := c.Get("seedDirectory")
	if !ok {
		c.JSON(http.StatusInternalServerError, CreateSeedResponse{"cannot determine seed directory"})
		return
	}
	var createSeedOpts = seed.CreateSeedOptions{DirectoryPath: seedDirectory.(string), UserProvidedSeedName: requestData.Filename}
	fs := afero.NewOsFs()
	filename, err := seed.CreateSeedFile(fs, createSeedOpts)
	if err != nil {
		c.JSON(http.StatusBadRequest, CreateSeedResponse{err.Error()})
		return
	}

	// Write contents to the file
	err = ioutil.WriteFile(*filename, []byte(requestData.SQL), 0655)
	if err != nil {
		c.JSON(http.StatusBadRequest, CreateSeedResponse{err.Error()})
		return
	}

	// create a seed file
	c.JSON(http.StatusOK, CreateSeedResponse{fmt.Sprintf("created seed file %s", *filename)})

}
