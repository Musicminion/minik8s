package stringutil

import (
    "math/rand"
    "time"
)

func GenerateRandomStr(length int) string {
    source := rand.NewSource(time.Now().UnixNano())
    rng := rand.New(source)
    letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    b := make([]rune, length)
    for i := range b {
        b[i] = letterRunes[rng.Intn(len(letterRunes))]
    }
    return string(b)
}