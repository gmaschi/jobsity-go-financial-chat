package chatmodel

type (
	GetRoom struct {
		RoomID string `uri:"roomId" binding:"required,numeric"`
	}
)
