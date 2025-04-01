package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello Go URL-shortener!",
		})
	})

	err := r.Run(":9008")
	if err != nil {
		panic(fmt.Sprint("Failed to start web server. Error: %v", err))
	}
}
