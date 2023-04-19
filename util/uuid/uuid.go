package uuid

import (
	"fmt"
	"github.com/google/uuid"
)

func NewUUID() string {
	uuid := uuid.New()
	return fmt.Sprintf("%s", uuid)
}
