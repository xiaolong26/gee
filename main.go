package main

import (
	"fmt"
	"gee"
)

func main()  {
	g := gee.New()
	g.GET("/", func(c *gee.Context) {
		c.JSON(200,"asd")
	})
	err := g.Run(":9090")
	if err!=nil{
		fmt.Println(err)
	}
}

