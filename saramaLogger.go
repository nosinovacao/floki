package main

import "fmt"

type saramaLogger struct {
}

func (s saramaLogger) Print(v ...interface{}) {
    logger.Log(v)
}

func (s saramaLogger) Printf(format string, v ...interface{}) {
    logger.Log(fmt.Sprintf(format, v))
}

func (s saramaLogger) Println(v ...interface{}) {
    s.Print(v)
}
