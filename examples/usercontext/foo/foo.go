package foo

import "fmt"

// Client is a hypothetical API client.
type Client struct {
	answer int
}

func NewClient() *Client {
	return &Client{}
}

func (fc *Client) String() string {
	return fmt.Sprint("The answer is: ", fc.answer)
}

func (fc *Client) Something() {
	fc.answer++
	if fc.answer == 42 {
		fc.answer = 0
	}
}
