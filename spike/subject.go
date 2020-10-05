package spike

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	uctfirestore "github.com/tevjef/uct-backend/common/firestore"
	"github.com/tevjef/uct-backend/common/middleware"
	"github.com/tevjef/uct-backend/common/middleware/cache"
	"github.com/tevjef/uct-backend/common/middleware/httperror"
	"github.com/tevjef/uct-backend/common/model"
)

func subjectHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		subjectTopicName := strings.ToLower(c.Param("topic"))
		firestore := uctfirestore.FromContext(c)

		if sub, err := firestore.GetSubject(c, subjectTopicName); err != nil {
			httperror.ServerError(c, err)
			return
		} else {
			response := model.Response{
				Data: &model.Data{Subject: sub},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}

func subjectsHandler(expire time.Duration) gin.HandlerFunc {
	return cache.CachePage(func(c *gin.Context) {
		season := strings.ToLower(c.Param("season"))
		year, err := strconv.Atoi(c.Param("year"))
		if err != nil {
			httperror.BadRequest(c, err)
			return
		}

		uniTopicName := strings.ToLower(c.Param("topic"))

		firestore := uctfirestore.FromContext(c)

		semester := &model.Semester{
			Year:   int32(year),
			Season: season,
		}
		if subjects, err := firestore.GetSubjectsBySemester(c, uniTopicName, semester); err != nil {
			httperror.ServerError(c, err)
			return
		} else {
			response := model.Response{
				Data: &model.Data{Subjects: subjects},
			}
			c.Set(middleware.ResponseKey, response)
		}
	}, expire)
}
