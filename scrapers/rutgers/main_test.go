package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	uct "uct/common"
)

func TestRutgersSemester(t *testing.T) {
	getRutgersSemester(uct.Semester{Season: uct.FALL, Year: 2016})
	assert.Equal(t, "92016", getRutgersSemester(uct.Semester{Season: uct.FALL, Year: 2016}))
}
