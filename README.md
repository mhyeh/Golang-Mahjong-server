# Mahjong server

A mahjong server based on Golang

## Files

| File Name | Description |
| --- | --- |
| mahjong / Action.go | Action made by player |
| mahjong / Broadcast.go | Broadcast message to player in same room |
| mahjong / GameLogic.go | Main Mahjong logic |
| mahjong / GameManager.go | Room management , player matching, login/logout, etc. |
| mahjong / InputChecker.go | Check player's input |
| mahjong / Player.go | Struct of player |
| mahjong / PlayerManager.go | Manage player list |
| mahjong / Room.go | Struct of room |
| mahjong / RoomInfo.go | Recover game state |
| mahjong / SocketEvent.go | Handle socket event |
| mahjong / SSJ.go | Check Hu |
| mahjong / Suit.go | Struct of Mahjong suit |
| mahjong / SuitSet.go | A set of Mahjong suit |
| mahjong / Tile.go | Struct of Mahjong tile |
| mahjong / Util.go | Useful function |
| server.go | main program |

## TODO

- ~~Disconnect handle~~
- ~~Check Hu logic~~
- ~~Check PonGon and OnGon broadcast function~~
- ~~Check data send from socket~~
- ~~Error Handling~~
- Optimize
- Refactor
- Account System
- Game log storing
- AI