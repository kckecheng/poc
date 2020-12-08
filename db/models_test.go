package db

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const DB string = "test.db"

var database *gorm.DB

func init() {
	var err error
	database, err = gorm.Open(sqlite.Open(DB), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}

func TestModles(t *testing.T) {
	database.AutoMigrate(&Command{})
	database.AutoMigrate(&Result{})
}

func TestCreate(t *testing.T) {
	commands := []Command{}
	for i := 0; i < 100; i++ {
		cmd := Command{
			Cli:    fmt.Sprintf("Command %d", i),
			Target: uuid.New().String(),
		}
		commands = append(commands, cmd)
	}
	database.Create(&commands)

	results := []Result{}
	for i := 0; i < 1000; i++ {
		cid := rand.Intn(101)
		wid := rand.Intn(11)
		var sflag bool
		if rand.Intn(2) > 0 {
			sflag = true
		}
		result := Result{
			CommandID: cid,
			Output:    fmt.Sprintf("Command %d output", cid),
			Where:     fmt.Sprintf("Node %d", wid),
			Success:   sflag,
		}
		results = append(results, result)
	}
	database.Create(&results)
}

func TestRead(t *testing.T) {
	// Select 10 commands randomly
	cids := []int{}
	for i := 0; i < 10; i++ {
		cid := rand.Intn(101)
		cids = append(cids, cid)
	}

	t.Log("Commands ...")
	commands := []Command{}
	database.Find(&commands, cids)
	for _, cmd := range commands {
		t.Log(cmd)
	}

	t.Log("Results ...")
	results := []Result{}
	database.Where("command_id IN ?", cids).Find(&results)
	for _, result := range results {
		t.Log(result)
	}
}
