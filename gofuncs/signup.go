package rollpoker

import (
)

func (player *Player) TryAllEndGame(rdata *RoomData, gc *GameCommand) *CommandResponse {
	_, has := rdata.Room.Hosts[gc.PlayerId]
	if !has {
		return CError("You are not a Host")
	}
	rdata.Room.RoomState = SIGNUP
	return CSave()
}

func (player *Player) TryAllPlayerUpdate(rdata *RoomData, gc *GameCommand) *CommandResponse {
	dname := gc.Args["DisplayName"]
	if dname != "" {
		rdata.Room.Members[gc.PlayerId] = dname
		if player != nil {
			player.DisplayName = dname
		}
	} else {
		return CError("Bad display name")
	}
	return CSave()
}

func (player *Player) TryAllJoinGroup(rdata *RoomData, gc *GameCommand) *CommandResponse {
	roompass := gc.Args["RoomPass"]
	if rdata.Room.OrigSettings.RoomPass != roompass {
		return CError("Invalid password")
	}
	dname := gc.Args["DisplayName"]
	if dname == "" {
		// TODO: Better validation. Not only whitespace, etc.
		return CError("DisplayName must be set")
	}

	if rdata.Room.Members == nil {
		rdata.Room.Members = make(map[string]string)
	}
	rdata.Room.Members[gc.PlayerId] = dname
	return CSave()
}

func (player *Player) TrySignupRegister(rdata *RoomData, gc *GameCommand) *CommandResponse {
	newPlayer := new(Player)

	newPlayer.PlayerId = gc.PlayerId
	newPlayer.DisplayName = rdata.Room.Members[gc.PlayerId]
	newPlayer.DisplayState = "Joined"

	if rdata.Room.Players == nil {
		rdata.Room.Players = make(map[string]*Player)
	}
	rdata.Room.Players[gc.PlayerId] = newPlayer
	return CSave()
}
