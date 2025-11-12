package xurl_test

import (
	"fmt"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xurl"
)

func ExampleIsLocalHostOrigin() {
	fmt.Println(xurl.IsLocalHostOrigin("http://localhost:8080"))
	fmt.Println(xurl.IsLocalHostOrigin("http://127.0.0.1:3000"))
	fmt.Println(xurl.IsLocalHostOrigin("https://localhost"))
	fmt.Println(xurl.IsLocalHostOrigin("https://example.com"))
	fmt.Println(xurl.IsLocalHostOrigin(""))
	// Output:
	// true
	// true
	// true
	// false
	// false
}

func ExampleIsLocalHostOrigin_httpVsHttps() {
	// Both HTTP and HTTPS localhost are accepted
	fmt.Println(xurl.IsLocalHostOrigin("http://localhost:8080"))
	fmt.Println(xurl.IsLocalHostOrigin("https://localhost:8080"))
	fmt.Println(xurl.IsLocalHostOrigin("http://127.0.0.1:3000"))
	fmt.Println(xurl.IsLocalHostOrigin("https://127.0.0.1:3000"))
	// Output:
	// true
	// true
	// true
	// true
}
