package jsonstore

var token = "6e2b459782b0f1b2e89d36d6424717949c795b3b63147ae2ef9cfb94c48d75f8"

/*

func TestGetBytes(t *testing.T) {
	client := NewClient(token)
	resp, err := client.GetBytes("users/1")
	if err != nil {
		t.Fatalf("client.GetBytes unexpected error: %s", err)
	}
	fmt.Printf("%s\n", resp)
}

func TestGet(t *testing.T) {
	client := NewClient(token)
	err := client.Get("users/1", nil)
	fmt.Println(err)
	if err != ErrNoValue {
		t.Errorf("client.Get returned wrong error. Expected='%s' Got='%s'",
			ErrNoValue, err)
	}
}

*/
