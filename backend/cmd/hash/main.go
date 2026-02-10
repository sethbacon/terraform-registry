package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    key := "dev_qHlTX4JvjK1yVUgRukLlgiwFQmFOiHdEhHYVJNfhNXc"
    hash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(hash))
}