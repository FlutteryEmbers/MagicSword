package main

import (
	"fmt"
	"sync"
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
	Result       int   `json:"result"`
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

const reactTime int64 = 2

var userMAP map[int32]player
var combactMAP map[string][]player
var l sync.Mutex
var l2 sync.Mutex

func main() {
	combactMAP = make(map[string][]player)
	userMAP = make(map[int32]player)
	r := gin.Default()
	r.GET("/v1/normal", normal)
	r.GET("/v1/combat", combat)
	r.GET("/v1/statics", getStats)
	r.GET("/v1/restart", restart)
	r.StaticFile("/", "./project_website/project_website/index.html")
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
		r.Res = response{Message: "cannot target self"}
		c.JSON(500, r)
		return
	}

	if !userMAP[r.User.ID].Status.InCombat {
		r.User = userMAP[r.User.ID]
		r.Res = response{Message: "user not in battle"}
		c.JSON(500, r)
		return
	}

	if !userMAP[r.Target].Status.InCombat {
		r.User = userMAP[r.User.ID]
		r.Res = response{Message: "user not in battle"}
		c.JSON(500, r)
		return
	}

	if r.Damage > 0 {
		if r.Damage == 3 {
			l.Lock()
			p := userMAP[r.Target]
			l.Unlock()
			p.Status.HP = p.Status.HP - 3
			l.Lock()
			userMAP[r.Target] = p
			l.Unlock()
		} else {
			l.Lock()
			p := userMAP[r.Target]
			l.Unlock()
			p.Status.UnderAttack = true
			p.Status.Timestamp = time.Now().Unix()
			// p.Status.DamageAlert = p.Status.DamageAlert + r.Damage
			p.Status.DamageAlert = r.Damage
			l.Lock()
			userMAP[r.Target] = p
			l.Unlock()
		}

	} else if r.Damage < 0 {
		l.Lock()
		p := userMAP[r.User.ID]
		l.Unlock()
		p.Status.UnderAttack = false
		p.Status.Timestamp = time.Now().Unix()
		if p.Status.DamageAlert > 0 {
			p.Status.HP = p.Status.HP - (p.Status.DamageAlert + r.Damage)
		}
		p.Status.DamageAlert = 0
		if p.Status.HP <= 0 {
			p.Status.InCombat = false
			r.Res = response{Code: 1, Message: "you lose"}
			l.Lock()
			target := userMAP[r.Target]
			l.Unlock()
			target.Status.InCombat = false
			target.Status.NumWin = target.Status.NumWin + 1
			l.Lock()
			userMAP[r.Target] = target
			l.Unlock()
			p.Status.NumLose = p.Status.NumLose + 1
		}
		l.Lock()
		userMAP[r.User.ID] = p
		l.Unlock()
		r.User = p
		c.JSON(200, r)
		return
	}

	l.Lock()
	p2 := userMAP[r.User.ID]
	l.Unlock()

	if p2.Status.UnderAttack && p2.Status.Timestamp+reactTime < time.Now().Unix() {
		fmt.Println("---------------------------------------------")
		p2.Status.UnderAttack = false
		p2.Status.Timestamp = time.Now().Unix()
		p2.Status.HP = p2.Status.HP - p2.Status.DamageAlert
		p2.Status.DamageAlert = 0
		if p2.Status.HP <= 0 {
			p2.Status.InCombat = false
			r.Res = response{Code: 1, Message: "you lose"}
			p2.Status.NumLose = p2.Status.NumLose + 1
			l.Lock()
			target := userMAP[r.Target]
			l.Unlock()

			target.Status.InCombat = false
			target.Status.NumWin = target.Status.NumWin + 1
			l.Lock()
			userMAP[r.Target] = target
			l.Unlock()
		}
		l.Lock()
		userMAP[r.User.ID] = p2
		l.Unlock()
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
	l.Lock()
	p, ok2 := userMAP[r.User.ID]
	l.Unlock()
	if p.Status.InCombat {
		r.User = p
		r.Res = response{Code: 1, Message: "battle start"}
		c.JSON(200, r)
		return
	}
	// check if the device wants to start combat
	if r.StartCombat {
		if r.Target == r.User.ID {
			r.Res = response{Code: 1, Message: "cannot target-self"}
			c.JSON(200, r)
			return
		}
		l.Lock()
		player, ok := userMAP[r.Target]
		l.Unlock()
		if ok {
			player.Status.InCombat = true
			player.Status.CombatTarget = r.User.ID
			l.Lock()
			userMAP[r.Target] = player
			l.Unlock()
			p.Status.InCombat = true
			p.Status.CombatTarget = r.Target
			l.Lock()
			userMAP[r.User.ID] = p
			l.Unlock()
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
	l2.Lock()
	players, ok := combactMAP[r.User.Geo]
	l2.Unlock()
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
		l2.Lock()
		players2 := combactMAP[p.Geo]
		l2.Unlock()
		var playerList2 []player
		for i := range players2 {
			if players2[i].ID != p.ID {
				playerList2 = append(playerList2, players2[i])
			}
		}
		l2.Lock()
		combactMAP[p.Geo] = playerList2
		l2.Unlock()
		r.User.Status.NumWin = p.Status.NumWin
		r.User.Status.NumLose = p.Status.NumLose
	}
	r.User.Status.HP = 100
	r.User.Status.UnderAttack = false
	r.User.Status.InCombat = false
	r.User.Status.Result = -1
	l.Lock()
	userMAP[r.User.ID] = r.User
	l.Unlock()
	l2.Lock()
	combactMAP[r.User.Geo] = updatedPlayers
	l2.Unlock()
	c.JSON(200, updatedPlayers)
}

func getStats(c *gin.Context) {
	var r request
	err := c.BindJSON(&r)
	if err != nil {
		c.JSON(500, request{Res: response{Code: 0, Message: "Cannot Parse Request"}})
		return
	}
	player, ok := userMAP[r.User.ID]
	if ok {
		c.JSON(200, player)
	} else {
		c.JSON(500, request{Res: response{Code: 0, Message: "Target not in map"}})
	}

}

func restart(c *gin.Context) {
	for k := range userMAP {
		delete(userMAP, k)
	}

	for k := range combactMAP {
		delete(combactMAP, k)
	}

	c.JSON(200, response{Code: 1, Message: "clean done !!!!!"})
}
