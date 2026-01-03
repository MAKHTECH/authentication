package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main() {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	privHex := hex.EncodeToString(priv)
	pubHex := hex.EncodeToString(priv.Public().(ed25519.PublicKey))

	fmt.Println("=== Ed25519 Key Pair ===")
	fmt.Println("")
	fmt.Println("PRIVATE_KEY:")
	fmt.Println(privHex)
	fmt.Println("")
	fmt.Println("PUBLIC_KEY:")
	fmt.Println(pubHex)
}
