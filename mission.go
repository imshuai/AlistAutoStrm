package main

import (
	"sync"

	sdk "github.com/imshuai/alistsdk-go"
)

// Mission is a struct that holds the mission data
type Mission struct {
	ID                   int
	RootPath             string
	MissionPath          string
	LocalDir             string
	IsCreateSubDirectory bool
	IsRecursive          bool
	Exts                 []string
	save                 chan<- File
	wg                   *sync.WaitGroup
	c                    *sdk.Client
}

// NewMission creates a new Mission
func NewMission(id int, rootPath, missionPath, localDir string, isCreateSubDirectory, isRecursive bool, exts []string, save chan<- File, wg *sync.WaitGroup, c *sdk.Client) *Mission {
	return &Mission{
		ID:                   id,
		RootPath:             rootPath,
		MissionPath:          missionPath,
		LocalDir:             localDir,
		IsCreateSubDirectory: isCreateSubDirectory,
		IsRecursive:          isRecursive,
		Exts:                 exts,
		save:                 save,
		wg:                   wg,
		c:                    c,
	}
}
