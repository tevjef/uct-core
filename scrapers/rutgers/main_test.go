package main

import (
	"testing"
	uct "uct/common"
	"github.com/stretchr/testify/assert"
)

func TestRutgersSemester(t *testing.T) {
	getRutgersSemester(uct.Semester{Season:uct.FALL, Year:2016})
	assert.Equal(t, "92016", getRutgersSemester(uct.Semester{Season:uct.FALL, Year:2016}))
}
