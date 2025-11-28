package kafka

type UserCreatedEvent struct {
	UserID    int64  `json:"userId"`
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
}

type UserUpdatedEvent struct {
	UserID    int64  `json:"userId"`
	Field     string `json:"field"`
	OldValue  string `json:"oldValue"`
	NewValue  string `json:"newValue"`
	Timestamp int64  `json:"timestamp"`
}
