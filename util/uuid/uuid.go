package uuid

import (
	"fmt"
	"github.com/google/uuid"
)

// 这个对应的是Google V4的UUID，和K8s保持一致
func NewUUID() string {
	uuid := uuid.New()
	return fmt.Sprintf("%s", uuid)
}
