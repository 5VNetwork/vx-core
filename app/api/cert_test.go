package api

import (
	context "context"
	"fmt"
	"testing"
)

func TestGenerateCert(t *testing.T) {
	a := Api{}
	resp, err := a.GenerateCert(context.Background(), &GenerateCertRequest{
		Domain: "www.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(resp.Cert))
	fmt.Println(string(resp.Key))
}
