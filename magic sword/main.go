package main

import (
	"github.com/gin-gonic/gin"
)

type action struct {
	underAttack bool
	timestamp   int64
	damage      int
}

type player struct {
	ID     int32  `json:"id"`
	Name   string `json:"name"`
	HP     int    `json:"hp"`
	Geo    string `json:"geo"`
	Combat bool   `json:"combat"`
	status action
}

type request struct {
	User      player `json:"user"`
	Mode      bool   `json:"mode"`
	Target    int32  `json:"target"`
	Damage    int    `json:"damage"`
	Timestamp int64  `json:"timestamp"`
}

type playList struct {
	Players []string `json:"players"`
}

type response struct {
	Message string `json:"message"`
}

const reactTime = 15

var userMAP map[int32]player
var combactMAP map[string][]player

func main() {
	combactMAP = make(map[string][]player)
	userMAP = make(map[int32]player)
	r := gin.Default()
	r.GET("/v1/normal", normal)
	r.GET("/v1/combat", combat)
	r.Run(":11000")
}

func combat(c *gin.Context) {
	var r request
	err := c.BindJSON(&r)
	if err != nil {
		c.JSON(500, response{Message: "invalid request"})
		return
	}
	if !userMAP[r.Target].Combat || !r.User.Combat {
		c.JSON(500, response{Message: "invalid request"})
		return
	}

	if r.Damage > 0 {
		p := userMAP[r.Target]
		p.status.underAttack = true
		p.status.timestamp = r.Timestamp
		p.status.damage = p.status.damage + r.Damage
		userMAP[r.Target] = p
	} else {
		p := userMAP[r.User.ID]
		p.status.underAttack = false
		p.status.timestamp = r.Timestamp
		p.status.damage = p.status.damage + r.Damage
		p.HP = r.User.HP - p.status.damage
		userMAP[r.User.ID] = p
		r.User.HP = p.HP
		c.JSON(200, r)
		return
	}

	p2 := userMAP[r.User.ID]
	if p2.status.underAttack && p2.status.timestamp+reactTime > r.Timestamp {
		p2.status.underAttack = false
		p2.status.timestamp = r.Timestamp
		p2.status.damage = 0
		p2.HP = r.User.HP - p2.status.damage
		userMAP[r.User.ID] = p2
		r.User.HP = p2.HP
		c.JSON(200, r)
		return
	}

	c.JSON(200, r)
}

func normal(c *gin.Context) {
	var r request
	err := c.BindJSON(&r)
	if err != nil {
		c.JSON(500, response{Message: "invalid request"})
		return
	}
	// check if is in battle status
	p, ok2 := userMAP[r.User.ID]
	if p.Combat {
		r.User.Combat = true
		c.JSON(200, r)
		return
	}
	// check if the device wants to start combat
	if r.User.Combat {
		player, ok := userMAP[r.Target]
		if ok {
			player.Combat = true
			userMAP[r.Target] = player
			c.JSON(200, r)
		} else {
			c.JSON(200, response{Message: "Invalid Target"})
		}
		return
	}
	// find the valid combat
	var updatedPlayers []player
	players, ok := combactMAP[r.User.Geo]
	if ok {
		var playerList map[int32]player
		playerList = make(map[int32]player)
		for i := range players {
			playerList[players[i].ID] = players[i]
		}
		playerList[r.User.ID] = r.User

		for k := range playerList {
			updatedPlayers = append(updatedPlayers, playerList[k])
		}
	} else {
		updatedPlayers = []player{r.User}
	}

	if ok2 {
		players2 := combactMAP[p.Geo]
		var playerList2 []player
		for i := range players2 {
			if players2[i].ID != p.ID {
				playerList2 = append(playerList2, players2[i])
			}
		}
		combactMAP[p.Geo] = playerList2
	}
	userMAP[r.User.ID] = r.User
	combactMAP[r.User.Geo] = updatedPlayers
	c.JSON(200, updatedPlayers)
}
