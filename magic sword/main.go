package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type user struct {
	ID  int32  `json:"id"`
	HP  int    `json:"hp"`
	Geo string `json:"geo"`
}

type request struct {
	ID     int32  `json:"id"`
	HP     int    `json:"hp"`
	Geo    string `json:"geo"`
	Mode   bool   `json:"mode"`
	Damage int    `json:"damage"`
	Shield int    `json:"shield"`
}

var map_user map[int32][]user
var map_combact map[string][]user

// map_user = make(map[string]user)

func main() {
	r := gin.Default()
	r.POST("/v1/combat", combat)
	r.Run(":11000")
}

func combat(c *gin.Context) {
	var r request
	err := c.BindJSON(&r)
	if err != nil {
		fmt.Println("error")
	}

}
