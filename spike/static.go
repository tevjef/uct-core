package main

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tevjef/uct-backend/spike/middleware/httperror"
)

func serveStaticFromGithub(c *gin.Context) {
	var branch = c.Query("branch")
	if branch != "" {
		branch = branch + "production/"
	}

	var file = c.Param("file")
	resp, err := http.Get("https://raw.githubusercontent.com/tevjef/uct-backend/" + branch + "static/" + file)

	defer resp.Body.Close()

	if err != nil {
		httperror.BadRequest(c, err)
		return
	}

	io.Copy(c.Writer, resp.Body)

	c.Status(200)
}
