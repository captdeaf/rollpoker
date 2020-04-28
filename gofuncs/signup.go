package rollpoker

import (
)

func (player *Player) TrySignupJoinGroup(rdata *RoomData, gc *GameCommand) *CommandResponse {
	roompass := gc.Args["RoomPass"]
	if rdata.Room.RoomPass != roompass {
		return CError("Invalid password")
	}

	rdata.Room.Members = append(rdata.Room.Members, gc.PlayerId)
	return CSave()
}

func (player *Player) TrySignupRegister(rdata *RoomData, gc *GameCommand) *CommandResponse {
	dname := gc.Args["DisplayName"]
	if dname == "" {
		// TODO: Better validation. Not only whitespace, etc.
		return CError("DisplayName must be set")
	}

	newPlayer := new(Player)

	newPlayer.PlayerId = gc.PlayerId
	newPlayer.DisplayName = dname
	newPlayer.DisplayState = "Joined"

	rdata.Room.Players[gc.PlayerId] = newPlayer
	return CSave()
}
