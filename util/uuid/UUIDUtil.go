package uuid

import "github.com/satori/go.uuid"

func GetRandomUUID() string {
	uid := uuid.NewV4()
	return uid.String()
}
