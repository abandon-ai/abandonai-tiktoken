package main

import (
	"fmt"
	"github.com/pkoukk/tiktoken-go"
)

func main() {
	text := "Hello, world!"
	encoding := "gpt-3.5-turbo"

	tkm, err := tiktoken.EncodingForModel(encoding)
	if err != nil {
		err = fmt.Errorf("getEncoding: %v", err)
		return
	}

	// encode
	token := tkm.Encode(text, nil, nil)

	// tokens
	fmt.Println(token)
	// num_tokens
	fmt.Println(len(token))
}
