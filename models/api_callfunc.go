package models

type ClientApiCallFunc func(ClientRequest) (int, []byte, error)
