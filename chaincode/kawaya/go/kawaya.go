type User struct {
  Id              string `json:"id"`
  Password        string `json:"password"`
  Balance         int    `json:"balance"`
  ReservedRoomId  string `json:"reserved_room_id"`
}

type Status int
const (
	StatusOk Status = 200
	StatusCreated Status = 201
	StatusConflict Status = 409
)

type ResultUser struct {
  Status  Status  `json:"status"`
  User    User    `json:"user"`
}

const DateTimeFormat = "2006-01-02 15:04:05 UTC"
type Room struct {
  Id              string    `json:"id"`
  StatusOfUse     string    `json:"status_of_use"`
  UnreservingTime time.Time `json:"unreserving_time"`
}
