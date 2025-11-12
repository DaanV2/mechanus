package xcrypto_test

import (
	"fmt"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"
)

func ExampleHashPassword() {
	password := []byte("mySecurePassword123")
	hash, err := xcrypto.HashPassword(password)
	if err != nil {
		fmt.Println("Error:", err)

		return
	}

	// Hash will be different each time due to salt
	fmt.Printf("Hash length: %d\n", len(hash))
	fmt.Printf("Hash starts with $2a: %t\n", len(hash) > 3 && string(hash[:3]) == "$2a")
	// Output:
	// Hash length: 60
	// Hash starts with $2a: true
}

func ExampleComparePassword() {
	password := []byte("mySecurePassword123")
	hash, _ := xcrypto.HashPassword(password)

	// Correct password
	match1, err1 := xcrypto.ComparePassword(hash, password)
	fmt.Printf("Correct password: match=%t, err=%v\n", match1, err1)

	// Incorrect password
	match2, err2 := xcrypto.ComparePassword(hash, []byte("wrongPassword"))
	fmt.Printf("Wrong password: match=%t, err=%v\n", match2, err2)
	// Output:
	// Correct password: match=true, err=<nil>
	// Wrong password: match=false, err=<nil>
}

func ExampleGenerateRSAKeys() {
	key, err := xcrypto.GenerateRSAKeys()
	if err != nil {
		fmt.Println("Error:", err)

		return
	}

	fmt.Printf("ID length: %d\n", len(key.ID()))
	fmt.Printf("Has private key: %t\n", key.Private() != nil)
	fmt.Printf("Has public key: %t\n", key.Public() != nil)
	// Output:
	// ID length: 28
	// Has private key: true
	// Has public key: true
}
