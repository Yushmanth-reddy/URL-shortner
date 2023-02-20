package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

// struct to map json file to Go object
type Config struct {
	Server struct {
		Port string `json:"port"` // this tag indicates that in json encoding Port field name is written as port
	} `json:"server"`

	Redis struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Password string `json:"password"`
	} `json:"redis"`

	Options struct {
		Prefix string `json:"prefix"`
		Schema string `json:"schema"`
	} `json:"options"`
}

func ReadFromFile(filePath string) (*Config, error) {
	b, err := ioutil.ReadFile(filePath)

	if err != nil {
		return nil, errors.New("Couldn't read file. Error : " + err.Error())
	}

	var cfg Config
	json.Unmarshal(b, &cfg) //unmarshal json bytes in b to variable referenced i.e. reference of cfg
	//passing &cfg, since structs are passed by value

	return &cfg, err
}
