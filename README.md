# Mahjong server

A mahjong server based on Golang

## Files

| File Name | Description |
| --- | --- |
| action / Action.go | Struct of Action |
| manager / PlayerManager.go | Manage player list |
| ssj / SSJ.go | Check Hu |
| tile / Color.go | Struct of Mahjong color |
| tile / Tile.go | Struct of Mahjong tile |
| tile / TileSet.go | A set of Mahjong color |
| util / Action.go | Action made by player |
| util / Broadcasr.go | Broadcast message to player in same room |
| util / GameLogic.go | Main Mahjong logic |
| util / GameManager.go | Room management , player matching, login/logout, etc. |
| util / InputChecker.go | Check player's input |
| util / Player.go | Struct of player |
| util / Room.go | Struct of room |
| util / RoomInfo.go | Recover game state |
| server.go | main program |

## TODO

- ~~Disconnect handle~~
- ~~Check Hu logic~~
- ~~Check PonGon and OnGon broadcast function~~
- ~~Check data send from socket~~
- Error Handling
- Optimize
- Refactor