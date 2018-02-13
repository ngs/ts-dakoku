package main

import "github.com/ngs/ts-dakoku/app"

func main() {
	if _, err := app.Run(); err != nil {
		panic(err)
	}
}
