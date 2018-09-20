package main

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func serveStaticFromGithub(c *gin.Context) {
	var branch = c.Query("branch")
	if branch != "" {
		branch = branch + "/"
	}

	var file = c.Param("file")
	resp, err := http.Get("https://raw.githubusercontent.com/tevjef/uct-backend/" + branch + "static/" + file)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	io.Copy(c.Writer, resp.Body)

	c.Status(200)
}
