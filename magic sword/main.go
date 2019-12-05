package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type status struct {
	HP           int   `json:"hp"`
	UnderAttack  bool  `json:"underAttack"`
	Timestamp    int64 `json:"timestamp"`
	DamageAlert  int   `json:"damageAlert"`
	InCombat     bool  `json:"inCombat"`
	CombatTarget int32 `json:"combatTarget"`
	NumWin       int   `json:"numWin"`
	NumLose      int   `json:"numLose"`
}

type player struct {
	ID     int32  `json:"id"`
	Name   string `json:"name"`
	Geo    string `json:"geo"`
	Status status `json:"status"`
}

type request struct {
	User        player   `json:"user"`
	StartCombat bool     `json:"startCombat"`
	Target      int32    `json:"target"`
	Damage      int      `json:"damage"`
	Res         response `json:"res"`
}

const reactTime int64 = 15

var userMAP map[int32]player
var combactMAP map[string][]player

func main() {
	combactMAP = make(map[string][]player)
	userMAP = make(map[int32]player)
	r := gin.Default()
	r.GET("/v1/normal", normal)
	r.GET("/v1/combat", combat)
	r.GET("/v1/statics", getStats)
	r.Run(":11000")
}

func combat(c *gin.Context) {
	var r request
	err := c.BindJSON(&r)
	if err != nil {
		c.JSON(500, request{Res: response{Code: 0, Message: "Invalid request"}})
		return
	}
	if r.Target == r.User.ID {
		c.JSON(500, response{Message: "cannot target self"})
		return
	}

	if !userMAP[r.Target].Status.InCombat {
		c.JSON(500, response{Message: "target not in battle"})
		return
	}

	if !userMAP[r.User.ID].Status.InCombat {
		c.JSON(500, response{Message: "user not in battle"})
		return
	}

	if r.Damage > 0 {
		p := userMAP[r.Target]
		p.Status.UnderAttack = true
		p.Status.Timestamp = time.Now().Unix()
		p.Status.DamageAlert = p.Status.DamageAlert + r.Damage
		userMAP[r.Target] = p
	} else if r.Damage < 0 {
		p := userMAP[r.User.ID]
		p.Status.UnderAttack = false
		p.Status.Timestamp = time.Now().Unix()
		p.Status.HP = p.Status.HP - (p.Status.DamageAlert + r.Damage)
		p.Status.DamageAlert = 0
		if p.Status.HP <= 0 {
			p.Status.InCombat = false
			r.Res = response{Code: 1, Message: "you lose"}
			target := userMAP[r.Target]
			target.Status.InCombat = false
			target.Status.NumWin = target.Status.NumWin + 1
			userMAP[r.Target] = target
			p.Status.NumLose = p.Status.NumLose + 1
		}
		userMAP[r.User.ID] = p
		r.User = p
		c.JSON(200, r)
		return
	}

	p2 := userMAP[r.User.ID]
	if p2.Status.UnderAttack && p2.Status.Timestamp+reactTime > time.Now().Unix() {
		fmt.Println("---------------------------------------------")
		p2.Status.UnderAttack = false
		p2.Status.Timestamp = time.Now().Unix()
		p2.Status.HP = p2.Status.HP - p2.Status.DamageAlert
		p2.Status.DamageAlert = 0
		if p2.Status.HP <= 0 {
			p2.Status.InCombat = false
			r.Res = response{Code: 1, Message: "you lose"}
			p2.Status.NumLose = p2.Status.NumLose + 1
			target := userMAP[r.Target]
			target.Status.InCombat = false
			target.Status.NumWin = target.Status.NumWin + 1
			userMAP[r.Target] = target
		}
		userMAP[r.User.ID] = p2
	}
	r.User = p2
	fmt.Println(userMAP)
	c.JSON(200, r)
}

func normal(c *gin.Context) {
	var r request
	err := c.BindJSON(&r)
	if err != nil {
		c.JSON(500, request{Res: response{Code: 0, Message: "Cannot Parse Request"}})
		return
	}
	// check if is in battle status
	p, ok2 := userMAP[r.User.ID]
	if p.Status.InCombat {
		r.User = p
		r.Res = response{Code: 1, Message: "battle start"}
		c.JSON(200, r)
		return
	}
	// check if the device wants to start combat
	if r.StartCombat {
		player, ok := userMAP[r.Target]
		if ok {
			player.Status.InCombat = true
			player.Status.CombatTarget = r.User.ID
			userMAP[r.Target] = player
			p.Status.InCombat = true
			p.Status.CombatTarget = r.Target
			userMAP[r.User.ID] = p
			r.Res = response{Code: 1, Message: "battle start"}
			r.User = p
			c.JSON(200, r)
		} else {
			c.JSON(200, request{Res: response{Code: 0, Message: "invalid target"}})
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
		r.User.Status.NumWin = p.Status.NumWin
		r.User.Status.NumLose = p.Status.NumLose
	}
	r.User.Status.HP = 100
	r.User.Status.UnderAttack = false
	r.User.Status.InCombat = false
	userMAP[r.User.ID] = r.User
	combactMAP[r.User.Geo] = updatedPlayers
	c.JSON(200, updatedPlayers)
}

func getStats(c *gin.Context) {
	var r request
	err := c.BindJSON(&r)
	if err != nil {
		c.JSON(500, request{Res: response{Code: 0, Message: "Cannot Parse Request"}})
		return
	}
	c.JSON(200, userMAP[r.User.ID])
}
