package main

import (
	"time"
	"uct/common/model"
)

func getRutgers(campus string) model.University {
	university := model.University{
		Name:             "Rutgers University–New Brunswick",
		Abbr:             "RU-NB",
		HomePage:         "http://newbrunswick.edu/",
		RegistrationPage: "https://sims.edu/webreg/",
		Registrations: []*model.Registration{
			{
				Period:     model.InFall.String(),
				PeriodDate: time.Date(2000, time.September, 6, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.InSpring.String(),
				PeriodDate: time.Date(2000, time.January, 17, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.InSummer.String(),
				PeriodDate: time.Date(2000, time.May, 30, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.InWinter.String(),
				PeriodDate: time.Date(2000, time.December, 23, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.StartFall.String(),
				PeriodDate: time.Date(2000, time.March, 20, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.StartSpring.String(),
				PeriodDate: time.Date(2000, time.October, 5, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.StartSummer.String(),
				PeriodDate: time.Date(2000, time.January, 14, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.StartWinter.String(),
				PeriodDate: time.Date(2000, time.September, 21, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.EndFall.String(),
				PeriodDate: time.Date(2000, time.September, 13, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.EndSpring.String(),
				PeriodDate: time.Date(2000, time.January, 27, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.EndSummer.String(),
				PeriodDate: time.Date(2000, time.June, 15, 0, 0, 0, 0, time.UTC).Unix(),
			},
			{
				Period:     model.EndWinter.String(),
				PeriodDate: time.Date(2000, time.December, 22, 0, 0, 0, 0, time.UTC).Unix(),
			},
		},
		Metadata: []*model.Metadata{{
			Title: "About", Content: aboutNewBrunswick,
		},
		},
	}

	if campus == "NK" {
		university.Name = "Rutgers University–Newark"
		university.Abbr = "RU-NK"
		university.HomePage = "http://www.newark.edu/"
		university.Metadata = []*model.Metadata{
			{
				Title: "About", Content: aboutNewark,
			},
		}
	}

	if campus == "CM" {
		university.Name = "Rutgers University–Camden"
		university.Abbr = "RU-CAM"
		university.HomePage = "http://www.camden.edu/"
		university.Metadata = []*model.Metadata{
			{
				Title: "About", Content: aboutCamden,
			},
		}
	}

	return university
}

const (
	aboutNewBrunswick = `<p><b>Rutgers University–New Brunswick</b> is the oldest campus of <a href="/wiki/Rutgers_Uni
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
					Work (Rutgers University)" class="mw-redirect">School of Social Work</a>.</p>`

	aboutNewark = `<p><b>Rutgers–Newark</b> is one of three regional campuses of <a href="/wiki/R
							utgers_University" title="Rutgers University">Rutgers University</a>, the <a href="/wiki/Public_universit
							y" title="Public university">public</a> research university of the <a href="/wiki/U.S._state" title="U.S
							. state">U.S. state</a> of <a href="/wiki/New_Jersey" title="New Jersey">New Jersey</a>, located in the
							 city of <a href="/wiki/Newark,_New_Jersey" title="Newark, New Jersey">Newark</a>. Rutgers, founded in 1
							 766 in <a href="/wiki/New_Brunswick,_New_Jersey" title="New Brunswick, New Jersey">New Brunswick</a>, i
							 s the <a href="/wiki/Colonial_colleges" title="Colonial colleges" class="mw-redirect">eighth oldest col
							 lege in the United States</a> and a member of the <a href="/wiki/Association_of_American_Universities"
							 title="Association of American Universities">Association of American Universities</a>. In 1945, the sta
							 te legislature voted to make Rutgers University, then a private <a href="/wiki/Liberal_arts_college" ti
							 tle="Liberal arts college">liberal arts college</a>, into the state university and the following year m
							 erged the school with the former <a href="/wiki/University_of_Newark" title="University of Newark" clas
							 s="mw-redirect">University of Newark</a> (1936–1946), which became the Rutgers–Newark campus. Rutgers a
							 lso incorporated the College of South Jersey and South Jersey Law School, in Camden, as a constituent c
							 ampus of the university and renamed it <a href="/wiki/Rutgers%E2%80%93Camden" title="Rutgers–Camden" cl
							 ass="mw-redirect">Rutgers–Camden</a> in 1950.</p> <p>Rutgers–Newark offers undergraduate (bachelors) an
							 d graduate (masters, doctoral) programs to more than 11,000 students. The campus is located on 38 acre
							 s in Newark's <a href="/wiki/University_Heights,_Newark,_New_Jersey" title="University Heights, Newark
							 , New Jersey" class="mw-redirect">University Heights</a> section. It consists of seven degree-granting
							  undergraduate, graduate, and professional schools, including the <a href="/wiki/Rutgers_Business_Schoo
							  l" title="Rutgers Business School" class="mw-redirect">Rutgers Business School</a> and <a href="/wiki/
							  Rutgers_School_of_Law_-_Newark" title="Rutgers School of Law - Newark" class="mw-redirect">Rutgers Sch
							  ool of Law - Newark</a>, and several research institutes including the <a href="/wiki/Institute_of_Ja
							  zz_Studies" title="Institute of Jazz Studies">Institute of Jazz Studies</a>. According to <i>U.S. News
							   &amp; World Report</i>, Rutgers–Newark is the most <a href="/wiki/Cultural_diversity" title="Cultural
								diversity">diverse</a> national university in the United States.</p>`

	aboutCamden = `<p><b>Rutgers University–Camden</b> is one of three regional campuses of <a
							href="/wiki/Rutgers_University" title="Rutgers University">Rutgers University</a>, the <a href="/wiki/N
							ew_Jersey" title="New Jersey">New Jersey</a>'s <a href="/wiki/Public_university" title="Public universit
							y">public</a> <a href="/wiki/Research_university" title="Research university" class="mw-redirect">resear
							ch university</a>. It is located in <a href="/wiki/Camden,_New_Jersey" title="Camden, New Jersey">Camden
							</a>, New Jersey, <a href="/wiki/United_States" title="United States">United States</a>. Founded in the
							1920s, Rutgers–Camden began as an amalgam of the South Jersey Law School and the College of South Jersey
							. It is the southernmost of the three regional campuses of Rutgers—the others being located in <a href="
							/wiki/New_Brunswick,_New_Jersey" title="New Brunswick, New Jersey">New Brunswick</a> and <a href="/wiki/
							Newark,_New_Jersey" title="Newark, New Jersey">Newark</a>.<sup id="cite_ref-3" class="reference"><a href
							="#cite_note-3"><span>[</span>3<span>]</span></a></sup> The city of Camden is located on the <a href="/w
							iki/Delaware_River" title="Delaware River">Delaware River</a> east of <a href="/wiki/Philadelphia,_Penn
							sylvania" title="Philadelphia, Pennsylvania" class="mw-redirect">Philadelphia</a>.</p>`

)
