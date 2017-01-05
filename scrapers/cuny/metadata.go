package main

import (
	"time"

	"github.com/tevjef/uct-core/common/model"
)

var cunyMetadata = map[string]model.University{

	"BAR": {
		Name:             "Baruch College",
		Abbr:             "BAR",
		HomePage:         "http://www.baruch.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `Baruch is one of CUNY's senior colleges, and traces its roots back to the 1847
			    founding of the Free Academy, the first institution of free public higher education in the United States.
			    The New York State Literature Fund was created to serve students who could not afford to enroll in New York
			    City’s private colleges. The Fund led to the creation of the Committee of the Board of Education of the City
			    of New York, led by Townsend Harris, J.S. Bosworth, and John L. Mason, which brought about the establishment
			    of what would become the Free Academy, on Lexington Avenue in Manhattan.

			    The Free Academy became the College of the City of New York, now The City College of New York (CCNY). In 1919,
			    what would become Baruch College was established as City College School of Business and Civic Administration.
			    On December 15, 1928, the cornerstone was laid on the new building which would house the newly founded school.
			    At this point, the school did not admit women. At the time it opened it was considered the biggest such school
			    for the teaching of business education in the United States.`,
			},
		},
	},

	"BMC": {
		Name:             "Borough of Manhattan Community College",
		Abbr:             "BMCC",
		HomePage:         "http://bmcc.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `The Borough of Manhattan Community College (BMCC) is one of seven two-year colleges
			    within the City University of New York (CUNY) system. Founded in 1963, BMCC originally offered business-oriented
			    and liberal arts degrees for those intending to enter the business world or transfer to a four-year college.
			    Its original campus was scattered all over midtown Manhattan, utilizing office space wherever available.
			    In the mid-1970s CUNY began scouting for suitable property on which to erect a new campus of its own.
			    The current campus has been in use since 1983. Currently, with an enrollment of over 26,000 students,
			    BMCC grants associate's degrees in a wide variety of vocational, business, health, science, and continuing
			    education fields.

			    Advertising itself to potential students under the motto "Start Here. Go Anywhere," its student body is nearly
			    two-thirds female and has a median age of 24, with attending students hailing from over 100 different countries.
			    Another 10,000 students are enrolled in distance education programs. BMCC has a faculty of nearly 1,000 full-time
			    and adjunct professors.`,
			},
		},
	},

	"BCC": {
		Name:             "Bronx Community College",
		Abbr:             "BCC",
		HomePage:         "http://bcc.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `Bronx Community College of The City University of New York offers more than 30
			    academic programs that prepare students for careers and to continue their education at four-year colleges.
			    Located on a 45-acre tree-lined campus, BCC is home to the Hall of Fame for Great Americans, the country’s
			    first hall of fame. The College provides its approximately 11,500 students with quality academic programs,
			    outstanding faculty, and flexible class schedules. BCC is a Hispanic Serving Institution (HSI), with students
			    representing approximately 100 countries. In October 2012, the BCC campus was declared a National Historic
			    Landmark, becoming the country’s first community college campus to receive such a designation. `,
			},
		},
	},
	"BKL": {
		Name:             "Brooklyn College",
		Abbr:             "BKL",
		HomePage:         "http://www.brooklyn.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `Bronx Community College of The City University of New York offers more than 30
			    academic programs that prepare students for careers and to continue their education at four-year colleges.
			    Located on a 45-acre tree-lined campus, BCC is home to the Hall of Fame for Great Americans, the country’s
			    first hall of fame. The College provides its approximately 11,500 students with quality academic programs,
			    outstanding faculty, and flexible class schedules. BCC is a Hispanic Serving Institution (HSI), with students
			    representing approximately 100 countries. In October 2012, the BCC campus was declared a National Historic
			    Landmark, becoming the country’s first community college campus to receive such a designation. `,
			},
		},
	},

	"LAW": {
		Name:             "CUNY School of Law",
		Abbr:             "LAW",
		HomePage:         "http://www.law.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `CUNY School of Law is the premier public interest law school in the country.
			    It trains lawyers to serve the underprivileged and disempowered and to make a difference in their communities.
			    CUNY Law pioneered the model of integrating a lawyering curriculum with traditional doctrinal study. The school
			    has been praised in a study by the Carnegie Foundation for the Advancement of Teaching, "Educating Lawyers:
			    Preparation for the Profession of Law," for being one of the few law schools in the country to prepare students
			    for practice through instruction in theory, skills, and ethics. Founded in 1983, the CUNY School of Law consistently
			    ranks among the top 10 law schools in the country in clinical training. With a student-faculty ratio of 10
			    to 1, CUNY Law has been a model for other law schools for its clinical practice. All third-year students at
			    CUNY Law represent clients under the supervision of attorneys at one of the largest law firms in Queens –
			    Main Street Legal Services, Inc. – situated right on the Law School campus.`,
			},
		},
	},

	"MED": {
		Name:             "CUNY School of Medicine",
		Abbr:             "MED",
		HomePage:         "https://www.ccny.cuny.edu/csom",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `The CUNY School of Medicine is a medical school that began operations in fall 2016
			    as part of the City University of New York. The school is in Harlem on the campus of the City College of New
			    York and partners with Saint Barnabas Health System in the South Bronx for clinical medical education. The school
			    will receive students from the BS/MD program at the Sophie Davis School of Biomedical Education who undertake
			    undergraduate studies for three years and then two years of didactic study in the sciences, after which students
			    complete their clinical training at St. Barnabas Hospital, culminating in their medical degree.`,
			},
		},
	},

	"SPH": {
		Name:             "CUNY School of Public Health",
		Abbr:             "SPH",
		HomePage:         "https://sph.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `The CUNY Graduate School of Public Health and Health Policy (commonly known as the
			    CUNY School of Public Health, or CUNY SPH) is a public American research and professional college within
			    the CUNY system based in New York City. The school is situated in its new building at 55 West 125th Street
			    in Manhattan. The CUNY School of Public Health offers 4 doctoral programs, 7 master's programs, an advanced
			    certificate and faculty memberships in at least 7 of CUNY's research centers and institutes. A core faculty
			    of over 45 full-time faculty is supplemented by additional faculty members drawn from throughout CUNY.`,
			},
		},
	},

	"CTY": {
		Name:             "City College of New York",
		Abbr:             "CCNY",
		HomePage:         "http://www.ccny.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `The founding institution of the City University of New York, City College offers outstanding
				teaching, learning and research on a beautiful campus in the heart of the world's most dynamic city. Our
				classrooms are equipped with the technology for a truly interactive learning environment.Our libraries
				hold 1.5 million volumes and provide online access to the resources of the entire university. Our
				laboratories are engines of innovation, where students and faculty push the boundaries of knowledge.

				Outstanding programs in architecture, engineering, education and the liberal arts and sciences prepare
				our students for the future, and produce outstanding leaders in every field.Whether they are drawn to
				the traditional, like philosophy or sociology, or emerging fields like sonic arts or biomedical
				engineering, our baccalaureate graduates go on to graduate programs at Stanford, Columbia or MIT – or
				they stay right here in one of our 50 master's programs or our doctoral programs in engineering, the
				laboratory sciences, and psychology.`,
			},
		},
	},

	"CSI": {
		Name:             "College of Staten Island",
		Abbr:             "CSI",
		HomePage:         "http://www.csi.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `The College of Staten Island is a senior college of The City University of New York (CUNY)
				offering Doctoral programs, Advanced Certificate programs, and Master’s programs, as well as Bachelor’s
				and Associate’s degrees. The College is accredited by the Middle States Commission on Higher Education.

				CSI is home to a School of Business, School of Education, and School of Health Sciences, as well as The
				Verrazano School Honors Program, and the Teacher Education Honors Academy. CSI is also a select campus
				of the Macaulay Honors College University Scholars program. The CUNY Interdisciplinary High-Performance
				Computing Center, one of the most powerful supercomputers in the New York City region, handles big-data
				analysis for faculty researchers and their student research teams. The new luxury residence halls, Dolphin
				Cove, have increased the College’s national and international desirability.`,
			},
		},
	},

	"NCC": {
		Name:             "Guttman Community College",
		Abbr:             "NCC",
		HomePage:         "http://www.guttman.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `Established on September 20, 2011, with Governor Andrew M. Cuomo’s approval of A Master Plan
				Amendment, The New Community College at CUNY was the University’s first new community college in more
				than 40 years. The second community college in Manhattan was inspired by Chancellor Matthew Goldstein’s
				interest in improving graduation rates for CUNY’s diverse urban students with a wide range of linguistic
				and cultural backgrounds. “There is no more urgent task in higher education than to find ways to help more
				community college students succeed,” the Chancellor has said.

				The New Community College at CUNY officially opened its doors in midtown Manhattan overlooking Bryant Park
				on August 20, 2012, after four years of planning in consultation with experts from around the country and
				hundreds of faculty and staff across the University. At the college’s inaugural Convocation, CUNY Chancellor
				Matthew Goldstein awarded Mayor Michael R. Bloomberg the prestigious Chancellor’s Medal from The City University
				of New York for his support and commitment to the development of this innovative new college. In accepting
				the medal the Mayor commented, “I think this school has the potential to be a game-changing model for
				community colleges across the country.”`,
			},
		},
	},

	"HOS": {
		Name:             "Hostos Community College",
		Abbr:             "HOS",
		HomePage:         "http://www.hostos.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `Eugenio María de Hostos Community College of The City University of New York is a community
				college in the City University of New York (CUNY) system located in the South Bronx, New York City.
				Hostos Community College was created by an act of the Board of Higher Education in 1968 in response to
				demands from the Hispanic/Puerto Rican community, which was urging for the establishment of a college
				to serve the people of the South Bronx. In 1970, the college admitted its first class of 623 students
				at the site of a former tire factory. Several years later, the college moved to a larger site nearby at
				149th Street and Grand Concourse.

				Hostos is the first institution of higher education on the mainland to be named after a Puerto Rican,
				Eugenio María de Hostos, an educator, writer, and patriot.[1] A large proportion (approximately 60 percent)
				of the student population is Hispanic, thus many of the courses at Hostos are offered in Spanish, and the
				college also provides extensive English and ESL instruction to students.`,
			},
		},
	},

	"HTR": {
		Name:             "Hunter College",
		Abbr:             "HTR",
		HomePage:         "http://www.hunter.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `Hunter College, located in the heart of Manhattan, is the largest college in the City University
				of New York (CUNY). Founded in 1870, it is also one of the oldest public colleges in the country. More
				than 23,000 students currently attend Hunter, pursuing undergraduate and graduate degrees in more than
				170 areas of study.

				Hunter's student body is as diverse as New York City itself. For more than 140 years, Hunter has provided
				educational opportunities for women and minorities, and today, students from every walk of life and every
				corner of the world attend Hunter.

				In addition to offering a multitude of academic programs in its prestigious School of Arts and Sciences,
				Hunter offers a wide breadth of programs in its preeminent Schools of Education, Nursing, Social Work,
				Health Professions and Urban Public Health.`,
			},
		},
	},

	"JJC": {
		Name:             "John Jay College",
		Abbr:             "JJC",
		HomePage:         "http://www.jjay.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `There is no college anywhere in the U.S. or the world quite like John Jay. Founded in 1964,
				John Jay College of Criminal Justice – a senior college of The City University of New York – has evolved
				into the preeminent international leader in educating for justice in its many dimensions. The College
				offers a rich liberal arts and professional curriculum that prepares students to serve the public interest
				as ethical leaders and engaged citizens.

				John Jay’s academic programs balance the sciences, humanities and the arts with professional studies.
				The College is unique in its mission, providing rigorous course work, research, internships, community
				service and other learning experiences to prepare students to make a difference for themselves and others
				and to transform ideas into social action and leadership.

				The College’s community of over 15,000 students, in a diverse array of undergraduate, graduate and doctoral
				programs, is the most diverse among the City University of New York’s senior colleges. The student body —
				60% female 40% Hispanic and 25% African American — today produces leaders, scholars and heroes in policing
				and beyond, including forensic science, law, fire and emergency management, social work, teaching, private
				security, forensic psychology and corrections. John Jay is also a leader in educating our nation’s military
				veterans, with more than 500 currently enrolled. `,
			},
		},
	},

	"KCC": {
		Name:             "Kingsborough CC",
		Abbr:             "KCC",
		HomePage:         "http://www.kingsborough.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `Founded in 1963 as part of the City University of New York (CUNY), Kingsborough Community
				College remains firmly committed to its mission of “providing both liberal arts and career education…[and]
				to promoting student learning and development as well as strengthening and serving its diverse community.”
				Kingsborough combines the best of what a college can offer.  Its tranquil seaside location is a perfect
				setting for reflective academic pursuits, yet the college’s active engagement in the community provides
				students with exciting opportunities to become productive participants in a growing and vital borough.

				Kingsborough has been named one of the leading community colleges in the country by the Aspen Institute
				College Excellence Program.  It has also earned national recognition for its creative and effective use
				of learning communities, for the large number of degrees it confers, for the high percentage of graduates
				who continue their studies, and for the innovative programs that draw thousands of non-traditional students
				to its campus every year.

				Kingsborough serves a widely diverse population of approximately 14,000 students and consistently ranks
				among the leading community colleges in the country in associate degrees awarded to minority students.
				Approximately 70% of Kingsborough’s students are enrolled in a liberal arts or science degree program;
				the rest pursue degrees in more specialized, career-oriented programs such as business, communications,
				criminal justice, culinary arts, nursing and allied health careers, information technology, journalism,
				maritime technology, tourism and hospitality, and the visual arts.  It also maintains one of the most
				comprehensive adult and continuing education programs in New York City.`,
			},
		},
	},

	"LAG": {
		Name:             "LaGuardia Community College",
		Abbr:             "LAG",
		HomePage:         "http://www.lagcc.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `LaGuardia Community College located in Long Island City, Queens, was founded in 1971 as a bold
				experiment in opening the doors of higher education to all, and we proudly carry forward that legacy today.
				LaGuardia educates students through over 50 degree, certificate and continuing education programs,
				providing an inspiring place for students to achieve their dreams.`,
			},
		},
	},

	"LEH": {
		Name:             "Lehman College",
		Abbr:             "LEH",
		HomePage:         "http://www.lehman.edu/",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `Lehman College is a senior college of the City University of New York (CUNY) in New York, United
				States. Founded in 1931 as the Bronx campus of Hunter College, the school became an independent college
				within CUNY in September 1967. The college is named after Herbert H. Lehman, a former New York governor,
				United States senator, and philanthropist. It is a public, comprehensive, coeducational liberal arts
				college with more than 90 undergraduate and graduate degree programs and specializations.`,
			},
		},
	},

	"MEC": {
		Name:             "Medgar Evers College",
		Abbr:             "MEC",
		HomePage:         "http://www.mec.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `Medgar Evers College is a senior college of The City University of New York (CUNY), offering
				baccalaureate and associate degrees. It was officially established in 1970 through cooperation between
				educators and community leaders in central Brooklyn. It is named after Medgar Wiley Evers, an African
				American civil rights leader who was assassinated on June 12, 1963.

				The college is divided into four schools: the School of Business, the School of Professional and
				Community Development, the School of Liberal Arts and Education, and the School of Science, Health,
				and Technology. The college also operates several external programs and associated centers such as the
				Male Development and Empowerment Center, the Center for Women's Development, the Center for Black
				Literature, and the DuBois Bunche Center for Public Policy. The college is a member of the Thurgood
				Marshall College Fund.`,
			},
		},
	},

	"NYT": {
		Name:             "New York City College of Technology",
		Abbr:             "NYT",
		HomePage:         "http://www.citytech.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `The College was founded in 1946, under the name “The New York State Institute of
			    Applied Arts and Sciences.” The urgent mission at the time was to provide training to GIs returning from the
			    Second World War and to provide New York with the technically proficient workforce it would need to thrive
			    in the emerging post-war economy. No one in 1946 could have predicted the transformation the College has
			    experienced. From its beginnings as an Institute - to being chartered as a community college in 1957 - and
			    subsequently transitioning to senior college status during the 1980’s - it has grown from serving 246
			    students in the class of 1947, to a population today of 17,400 students (about 6,500 in baccalaureate degree
			    programs) and serving over 16,000 others who participate in the continuing education offerings of the College.

			    Our mission, however, has not changed; It is still urgent, and it is still focused on preparing a
			    technically proficient workforce and well-educated citizens. The College’s offerings encompass the pre-professional,
			    professional and technical programs that respond to regional economic needs and provide access to higher education for
			    all who seek fulfillment of career and economic goals through education. The 27 associate and 24 baccalaureate degrees
			    offered allow graduates to pursue careers in the architectural and engineering technologies, the computer, entertainment,
			    and health professions, human services, advertising and publishing, hospitality, business, and law-related professions,
			    as well as programs in career and technical teacher education.`,
			},
		},
	},

	"QNS": {
		Name:             "Queens College",
		Abbr:             "QNS",
		HomePage:         "http://www.qc.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `At its founding in 1937, Queens College was hailed by the people of the borough as “the college
				of the future.” Now part of the City University of New York (CUNY), Queens College offers a rigorous
				education in the liberal arts and sciences under the guidance of a faculty dedicated to both teaching
				and research. Students graduate with the ability to think critically, address complex problems, explore
				various cultures, and use modern technologies and information resources.

				Located in a residential area of Flushing in the borough of Queens—America’s most ethnically diverse
				county—the college has students from more than 150 nations. A member of Phi Beta Kappa, Queens College
				is consistently ranked among the leading institutions in the nation for the quality of its academic
				programs and student achievement. Recognized as one of the most affordable public colleges in the
				country, Queens College offers a first-rate education to talented people of all backgrounds and financial
				means.`,
			},
		},
	},

	"QCC": {
		Name:             "Queensborough Community College",
		Abbr:             "QCC",
		HomePage:         "http://www.qcc.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `Queensborough proudly reflects the unique character of the local Queens community, the most
				diverse county in the United States. We distinguish ourselves from other higher education institutions
				in America because of that diversity, with nearly equal populations of African Americans, Asians,
				Caucasians and Latinos. In fact, our students come from more than 140 countries and speak some 84
				different languages.

				More than 16,000 students are currently enrolled in associate degree or certificate programs, and another
				10,000 students attend continuing education programs on our campus. Accredited by the Middle States
				Commission on Higher Education, Queensborough Community College, through its 17 academic departments,
				offers transfer and degree programs, including Associate in Arts (A.A.), the Associate in Science (A.S.),
				and the Associate in Applied Science (A.A.S.) degrees. The college also offers non-credit courses and
				certificate programs.`,
			},
		},
	},

	"GRD": {
		Name:             "Graduate Center of the City University of New York",
		Abbr:             "GRD",
		HomePage:         "http://www.gc.cuny.edu",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `The Graduate Center is located in the heart of Manhattan and set within the large and
				multi-campus City University of New York. It fosters advanced graduate education, original research and
				scholarship, innovative university-wide programs, and vibrant public events that draw upon and contribute
				to the complex communities of New York City and beyond. Through a broad range of nationally prominent
				doctoral programs, the Graduate School prepares students to be scholars, teachers, experts, and leaders
				in the academy, the arts and in the private, nonprofit, and government sectors. Committed to CUNY’s
				historic mission of educating the “children of the whole people,” we work to provide access to doctoral
				education for diverse groups of highly talented students, including those who have been underrepresented
				in higher education.

				The Graduate Center (GC) is the focal point for advanced teaching and research at the City University of
				New York (CUNY), the nation's largest urban public university. Devoted exclusively to graduate education,
				the GC fosters pioneering research and scholarship in the arts and sciences, and trains students for
				careers in universities and the private, nonprofit, and government sectors. With over 35 doctoral and
				master’s programs of the highest caliber, and 20 research centers, institutes, and initiatives, the GC
				benefits from highly ambitious and diverse students and alumni—who in turn teach hundreds of thousands
				of undergraduates every year. Through its public programs, the GC enhances New York City’s intellectual
				and cultural life.`,
			},
		},
	},

	"YRK": {
		Name:             "York College of The City University of New York",
		Abbr:             "YRK",
		HomePage:         "http://www.york.cuny.edu/",
		RegistrationPage: "https://home.cunyfirst.cuny.edu/oam/Portal_Login1.html",
		Registrations:    registrations,
		Metadata: []*model.Metadata{
			{
				Title: "About",
				Content: `York College of The City University of New York is one of eleven senior colleges in the City
				University of New York (CUNY) system. It is located in Jamaica, Queens in New York City. Founded in 1966,
				York was the first senior college founded under the newly formed CUNY system, which united several
				previously independent public colleges into a single public university system in 1961. The college is a
				member-school of Thurgood Marshall College Fund.

				Today, with an enrollment of more than 8,000 students, York serves as one of CUNY's leading liberal arts
				colleges, granting bachelor's degrees in more than 40 fields, including those in the Heath Professions,
				Nursing (BS) and a combined BS/MS degree in Occupational Therapy, among others. The York College Library
				subscribes to dozens of electronic resources, as well as print journals, to support the research needs of
				the faculty and students.

				Based on a study conducted by The Institute for College Access & Success (TICAS), NerdScholar, a scholarship
				information organization and website, recently listed York College as the “US College with the lowest student
				debt in 2013.” The national survey chose York as number one on its top 20 list of colleges and universities
				both private and public.`,
			},
		},
	},
}

var registrations = []*model.Registration{
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
}
