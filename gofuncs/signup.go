package rollpoker

import (
)

func (player *Player) TrySignupJoinGroup(rdata *RoomData, gc *GameCommand) *CommandResponse {
	roompass := gc.Args["RoomPass"]
	if rdata.Room.RoomPass != roompass {
		return CError("Invalid password")
	}
	dname := gc.Args["DisplayName"]
	if dname == "" {
		// TODO: Better validation. Not only whitespace, etc.
		return CError("DisplayName must be set")
	}

	rdata.Room.Members[gc.PlayerId] = dname
	return CSave()
}

func (player *Player) TrySignupRegister(rdata *RoomData, gc *GameCommand) *CommandResponse {
	newPlayer := new(Player)

	newPlayer.PlayerId = gc.PlayerId
	newPlayer.DisplayName = rdata.Room.Members[gc.PlayerId]
	newPlayer.DisplayState = "Joined"

	rdata.Room.Players[gc.PlayerId] = newPlayer
	return CSave()
}
