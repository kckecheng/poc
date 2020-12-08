package db

import (
	"gorm.io/gorm"
)

// Command model definition
type Command struct {
	gorm.Model
	Cli    string
	Target string // where to run the command: all, or specified servers (identified by UUID)
}

// Result model definition
type Result struct {
	gorm.Model
	CommandID int
	Command   Command `gorm:"foreignKey:CommandID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Output    string
	Success   bool
	Where     string
}
