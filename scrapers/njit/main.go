package main

import (
	uct "uct/common"
	"time"
"net/http"
)

var (
	rutgersNB *uct.University
	host = "http://sis.rutgers.edu/soc"
)

func main() {

	rutgersNB = &uct.University {
		Name: "Rutgers University–New Brunswick",
		Abbr: "RU-NB",
		MainColor: "F44336",
		AccentColor: "607D8B",
		HomePage: "http://newbrunswick.rutgers.edu/",
		RegistrationPage: "https://sims.rutgers.edu/webreg/",
		Registrations: []*uct.Registration{
			&uct.Registration{
				Period: uct.SEM_FALL,
				PeriodDate: time.Date(0000, time.September, 6, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.SEM_SPRING,
				PeriodDate: time.Date(0000, time.January, 17, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.SEM_SUMMER,
				PeriodDate: time.Date(0000, time.May, 30, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.SEM_WINTER,
				PeriodDate: time.Date(0000, time.December, 23, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.START_FALL,
				PeriodDate: time.Date(0000, time.March, 20, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.START_SPRING,
				PeriodDate: time.Date(0000, time.October, 18, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.START_SUMMER,
				PeriodDate: time.Date(0000, time.January, 14, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.START_WINTER,
				PeriodDate: time.Date(0000, time.September, 21, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.END_FALL,
				PeriodDate: time.Date(0000, time.September, 13, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.END_SPRING,
				PeriodDate: time.Date(0000, time.January, 27, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.END_SUMMER,
				PeriodDate: time.Date(0000, time.August, 15, 0, 0, 0, 0, time.UTC),
			},
			&uct.Registration{
				Period: uct.END_WINTER,
				PeriodDate: time.Date(0000, time.December, 22, 0, 0, 0, 0, time.UTC),
			},
		},
		Metadata: []uct.Metadata{
			uct.Metadata{
				"About", `<p><b>Rutgers University–New Brunswick</b> is the oldest campus of <a href="/wiki/Rutgers_Uni
				versity" title="Rutgers University">Rutgers University</a>, the others being in <a href="/wiki/Rutgers%
				E2%80%93Camden" title="Rutgers–Camden" class="mw-redirect">Camden</a> and <a href="/wiki/Rutgers%E2%80%
				93Newark" title="Rutgers–Newark" class="mw-redirect">Newark</a>. It is primarily located in the <a href
				="/wiki/New_Brunswick,_New_Jersey" title="New Brunswick, New Jersey">City of New Brunswick</a> and <a h
				ref="/wiki/Piscataway_Township,_New_Jersey" title="Piscataway Township, New Jersey" class="mw-redirect"
				>Piscataway Township</a>. The campus is composed of several smaller campuses: <i>College Avenue</i>, <i
				><a href="/wiki/Busch_Campus_(Rutgers_University)" title="Busch Campus (Rutgers University)" class="mw-
				redirect">Busch</a></i>, <i>Livingston,</i> <i>Cook</i>, and <i>Douglass</i>, the latter two sometimes
				referred to as "Cook/Douglass," as they are adjacent to each other. Rutgers–New Brunswick also includes
				 several buildings in downtown New Brunswick. The New Brunswick campuses include 19 undergraduate, grad
				 uate and professional schools, including the School of Arts and Sciences, School of Environmental and
				 Biological Sciences, School of Communication, Information and Library Studies, the <a href="/wiki/Edwa
				 rd_J._Bloustein_School_of_Planning_and_Public_Policy" title="Edward J. Bloustein School of Planning an
				 d Public Policy">Edward J. Bloustein School of Planning and Public Policy</a>, <a href="/wiki/School_o
				 f_Engineering_(Rutgers_University)" title="School of Engineering (Rutgers University)" class="mw-redir
				 ect">School of Engineering</a>, the <a href="/wiki/Ernest_Mario_School_of_Pharmacy" title="Ernest Mari
				 o School of Pharmacy">Ernest Mario School of Pharmacy</a>, the Graduate School, the Graduate School of
				  Applied and Professional Psychology, the Graduate School of Education, <a href="/wiki/School_of_Manag
				  ement_and_Labor_Relations" title="School of Management and Labor Relations" class="mw-redirect">School
				   of Management and Labor Relations</a>, the <a href="/wiki/Mason_Gross_School_of_the_Arts" title="Maso
				   n Gross School of the Arts">Mason Gross School of the Arts</a>, the College of Nursing, the <a href="
				   /wiki/Rutgers_Business_School" title="Rutgers Business School" class="mw-redirect">Rutgers Business
				   School</a> and the <a href="/wiki/School_of_Social_Work_(Rutgers_University)" title="School of Social
				    Work (Rutgers University)" class="mw-redirect">School of Social Work</a>.</p>`,
			},
		},
	}

	Semesters := uct.ResolveSemesters(time.Now(), rutgersNB.Registrations)

	LastSemester := Semesters[0]
	CurrentSemester := Semesters[1]
	NextSemester := Semesters[2]


}


func getSubjects(semester uct.Semester) {
	resp, err := http.Get(host + "/subjects.json?semester=92016&campus=NB&level=U")
	if err != nil {
		// handle error
	}
}

func getRutgersSemester(semester uct.Semester) {
	if (semester.Season == FALL)
	}
}