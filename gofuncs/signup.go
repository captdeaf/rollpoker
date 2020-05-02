package rollpoker

import (
)

func (player *Player) TryAllUpdateSettings(rdata *RoomData, gc *GameCommand) *CommandResponse {
	_, has := rdata.Room.Hosts[gc.PlayerId]
	if !has {
		return CError("You are not a Host")
	}
	rdata.Room.OrigSettings = MakeGameSettings(gc.Args)
	return CSave()
}

func (player *Player) TryAllPromote(rdata *RoomData, gc *GameCommand) *CommandResponse {
	flag, has := rdata.Room.Hosts[gc.PlayerId]
	if !has || flag != "OWNER" {
		return CError("You are not the Owner")
	}
	pid, has := gc.Args["PlayerId"]
	if !has { return CError("Unknown Member") }
	_, has = rdata.Room.Members[pid]
	if !has { return CError("Unknown Member") }
	rdata.Room.Hosts[pid] = "HOST"
	return CSave()
}

func (player *Player) TryAllDemote(rdata *RoomData, gc *GameCommand) *CommandResponse {
	flag, has := rdata.Room.Hosts[gc.PlayerId]
	if !has || flag != "OWNER" {
		return CError("You are not the Owner")
	}
	pid, has := gc.Args["PlayerId"]
	if !has { return CError("Unknown Member") }
	_, has = rdata.Room.Members[pid]
	if !has { return CError("Unknown Member") }
	delete(rdata.Room.Hosts, pid)
	return CSave()
}

func (player *Player) TrySignupKickPlayer(rdata *RoomData, gc *GameCommand) *CommandResponse {
	pid, has := gc.Args["PlayerId"]
	if !has { return CError("Unknown Member") }
	_, has = rdata.Room.Members[pid]
	if !has { return CError("Unknown Member") }
	_, has = rdata.Room.Players[pid]
	if !has { return CError("Not currently signed up") }
	if (pid != gc.PlayerId) {
		_, has := rdata.Room.Hosts[gc.PlayerId]
		if !has {
			return CError("You are not a Host")
		}
	}
	delete(rdata.Room.Players, pid)
	return CSave()
}

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
