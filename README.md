# Mahjong server

A mahjong server based on Golang

## Files

| File Name | Description |
| --- | --- |
| MJCard / Card.go | Struct of Mahjong card |
| MJCard / Color.go | Struct of Mahjong color |
| MJCard / Cards.go | A set of Mahjong color |
| PManager / PlayerManager.go | Manage player list |
| SSJ / SSJ.go | Check Hu |
| util / Action.go | Action made by player |
| util / Broadcasr.go | Broadcast message to player in same room |
| util / GameLogic.go | Main Mahjong logic |
| util / GameManager.go | Room management , player matching, login/logout, etc. |
| util / Player.go | Struct of player |
| util / Room.go | Struct of room |
| util / RoomInfo.go | Recover game state |
| server.go | main program |

## TODO

- ~~Disconnect handle~~
- ~~Check Hu logic~~
- ~~Check PonGon and OnGon broadcast function~~
- Check data send from socket
- Error Handling
- Optimize
- Refactor