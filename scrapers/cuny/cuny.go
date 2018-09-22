package main

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/tevjef/uct-backend/common/try"
	"net"
)

type cunyForm url.Values

func (cf cunyForm) setUniversity(university CunyUniversity) {
	cf.keyVal(universityKey, cunyUniversityId[university])
}

func (cf cunyForm) setAction(action string) {
	cf.keyVal("ICAction", action)

}

func (cf cunyForm) setICSID(icsid string) {
	cf.keyVal("ICSID", icsid)
}

func (cf cunyForm) setTerm(term string) {
	cf.keyVal(termKey, term)
}

func (cf cunyForm) setSubject(subjectId string) {
	cf.keyVal("SSR_CLSRCH_WRK_SSR_OPEN_ONLY$chk$5", "N")
	cf.keyVal("SSR_CLSRCH_WRK_CAMPUS$14", "MAIN")
	cf.keyVal(subjectKey, subjectId)
}

func (cf cunyForm) keyVal(key, val string) {
	cf[key] = []string{val}
}

type CunyFirstClient struct {
	httpClient *http.Client
	values     url.Values
}

var defaultScraper = &cunyScraper{
	university: CityCollege,
	client: &CunyFirstClient{
		values:     map[string][]string{},
		httpClient: newClient(),
	}}

func (cf *CunyFirstClient) extractValues(doc *goquery.Document) {
	cf.values = map[string][]string{}
	doc.Find("input").Each(func(index int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		value, _ := s.Attr("value")
		cf.values[name] = []string{value}
	})
}

var ErrTimeout error = errors.New("http connection timed out")

func (cf *CunyFirstClient) Post(url string, values url.Values) (*goquery.Document, error) {
	// Merge form
	formValues := cf.values
	for key, val := range values {
		formValues[key] = val
	}

	//if len(formValues[subjectKey]) > 0 {
	//	log.WithFields(log.Fields{"subject": formValues[subjectKey]}).Debugln("subject")
	//
	//}
	//log.WithFields(log.Fields{"action": formValues["ICAction"]}).Debugln("scrapeCourses")

	var resp *http.Response
	err := try.DoWithOptions(func(attempt int) (retry bool, err error) {
		resp, err = cf.httpClient.PostForm(url, formValues)

		// If request timed out do not retry
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return false, ErrTimeout
		}
		if err != nil {
			log.WithError(err).WithField("form", values).Errorln("could not reach cunyfirst")
			return true, err
		}

		return false, nil
	}, &try.Options{try.ExponentialJitterBackoff, 5})

	if doc, err := goquery.NewDocumentFromResponse(resp); err != nil {
		log.WithError(err).Errorln("error reading response body")
	} else {
		if doc != nil {
			cf.extractValues(doc)
		}

		return doc, nil
	}

	return nil, err
}

func (cf *CunyFirstClient) Get(url string) (*goquery.Document, error) {
	var resp *http.Response

	err := try.DoWithOptions(func(attempt int) (retry bool, err error) {
		resp, err = cf.httpClient.Get(url)
		// If request timed out do not retry
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return false, ErrTimeout
		}
		if err != nil {
			log.WithError(err).Errorln("could not reach cunyfirst")
			return true, err
		}

		return false, nil
	}, &try.Options{BackoffStrategy: try.ExponentialJitterBackoff, MaxRetries: 5})

	if doc, err := goquery.NewDocumentFromResponse(resp); err != nil {
		log.WithError(err).Errorln("error reading erpsonse body")
	} else {
		if doc != nil {
			cf.extractValues(doc)
		}

		return doc, err
	}

	return nil, err
}

const (
	initalPage = "https://hrsa.cunyfirst.cuny.edu/psc/cnyhcprd/GUEST/HRMS/c/COMMUNITY_ACCESS.CLASS_SEARCH.GBL"
	termSearch = "https://hrsa.cunyfirst.cuny.edu/psc/cnyhcprd/GUEST/HRMS/c/COMMUNITY_ACCESS.CLASS_SEARCH.GBL"
)

const (
	searchAction       = "CLASS_SRCH_WRK2_SSR_PB_CLASS_SRCH"
	initAction         = "CLASS_SRCH_WRK2_SSR_PB_CLASS_SRCH"
	modifySearchAction = "CLASS_SRCH_WRK2_SSR_PB_MODIFY"
	newSearchAction    = "CLASS_SRCH_WRK2_SSR_PB_NEW_SEARCH"
	sectionBackAction  = "CLASS_SRCH_WRK2_SSR_PB_BACK"
)

const (
	universityKey = "CLASS_SRCH_WRK2_INSTITUTION$31$"
	termKey       = "CLASS_SRCH_WRK2_STRM$35$"
	subjectKey    = "SSR_CLSRCH_WRK_SUBJECT_SRCH$0"
)

var (
	selectSubjects        = "select[id*=" + subjectKey[:len(subjectKey)-2] + "]"
	selectCourses         = "div[id*=win0divSSR_CLSRSLT_WRK_GROUPBOX2]"
	selectClassAttributes = "span#SSR_CLS_DTL_WRK_SSR_CRSE_ATTR_LONG"
	selectDesignation     = "span#SSR_CLS_DTL_WRK_DESCRFORMAL"
	selectRequirements    = "span#SSR_CLS_DTL_WRK_SSR_REQUISITE_LONG"
	selectClassComponents = "#win0divSSR_CLS_DTL_WRK_SSR_COMPONENT_LONG"
	selectInstructionMode = "span#INSTRUCT_MODE_DESCR"
	selectUnits           = "span#SSR_CLS_DTL_WRK_UNITS_RANGE"
	selectDesc            = "span#DERIVED_CLSRCH_DESCRLONG"
	selectDescTop         = "span#DERIVED_CLSRCH_DESCR200"
	selectEnrollmentCap   = "SSR_CLS_DTL_WRK_ENRL_CAP"
	selectEnrollmentTotal = "SSR_CLS_DTL_WRK_ENRL_TOT"
	selectAvailableSeats  = "SSR_CLS_DTL_WRK_AVAILABLE_SEATS"
	selectWaitCap         = "SSR_CLS_DTL_WRK_WAIT_CAP"
	selectWaitTotal       = "SSR_CLS_DTL_WRK_WAIT_TOT"
	selectGridRow         = ".PSLEVEL3GRIDROW"
	selectSectionLink     = "a.PSHYPERLINK"
)

func parseDay(day string) string {
	switch day {
	case "Mo":
		return "Monday"
	case "Tu":
		return "Tuesday"
	case "We":
		return "Wednesday"
	case "Th":
		return "Thursday"
	case "Fr":
		return "Friday"
	case "Sa":
		return "Saturday"
	default:
		return ""
	}
}

func splitMeeting(meeting string) [3]string {
	tups := [3]string{}
	parts := strings.Split(strings.TrimSpace(meeting), " ")

	if len(parts) == 4 {
		tups[0] = parts[0]
		tups[1] = parseTime(parts[1])
		tups[2] = parseTime(parts[3])
	}

	return tups
}

func parseTime(s string) string {
	t := s[:len(s)-2]
	meridian := s[len(s)-2:]

	return t + " " + meridian
}

func expandMeeting(meeting string) []string {
	if len(meeting) < 10 {
		return []string{meeting}
	}
	spaceIndex := strings.Index(meeting, " ")

	day := meeting[:spaceIndex]
	time := meeting[spaceIndex+1:]

	var meetings []string

	for i := 0; i < len(day)/2; i++ {
		d := parseDay(day[i*2 : i*2+2])
		if d != "" {
			meetings = append(meetings, d+" "+time)
		}
	}

	return meetings
}

var classType = []string{
	"Clinical",
	"Conference Hour",
	"Continuance",
	"Discussion",
	"Dissertation",
	"Field Studies",
	"Independent Study",
	"Internship",
	"Laboratory",
	"Lecture",
	"Message",
	"Practicum",
	"Recitation",
	"Research",
	"Seminar",
	"Supplemental Instruction",
	"Test Section",
	"Thesis Research",
	"Tutorial",
}

var cunyUniversity = []string{
	"Baruch College",
	"Borough of Manhattan CC",
	"Bronx Community College",
	"Brooklyn College",
	"CUNY School of Law",
	"CUNY School of Medicine",
	"CUNY School of Public Health",
	"City College",
	"College of Staten Island",
	"Guttman Community College",
	"Hostos Community College",
	"Hunter College",
	"John Jay College",
	"Kingsborough CC",
	"LaGuardia Community College",
	"Lehman College",
	"Medgar Evers College",
	"NYC College of Technology",
	"Queens College",
	"Queensborough CC",
	"School of Professional Studies",
	"The Graduate Center",
	"York College",
}

type CunyUniversity int

const (
	BaruchCollege CunyUniversity = iota
	BoroughofManhattanCC
	BronxCommunityCollege
	BrooklynCollege
	CUNYSchoolofLaw
	CUNYSchoolofMedicine
	CUNYSchoolofPublicHealth
	CityCollege
	CollegeofStatenIsland
	GuttmanCommunityCollege
	HostosCommunityCollege
	HunterCollege
	JohnJayCollege
	KingsboroughCC
	LaGuardiaCommunityCollege
	LehmanCollege
	MedgarEversCollege
	NYCCollegeofTechnology
	QueensCollege
	QueensboroughCC
	SchoolofProfessionalStudies
	TheGraduateCenter
	YorkCollege
)

var cunyUniversityId = []string{
	"BAR01",
	"BMC01",
	"BCC01",
	"BKL01",
	"LAW01",
	"MED01",
	"SPH01",
	"CTY01",
	"CSI01",
	"NCC01",
	"HOS01",
	"HTR01",
	"JJC01",
	"KCC01",
	"LAG01",
	"LEH01",
	"MEC01",
	"NYT01",
	"QNS01",
	"QCC01",
	"SPS01",
	"GRD01",
	"YRK01",
}

var cunyUniversityAbbr = []string{
	"BAR",
	"BMC",
	"BCC",
	"BKL",
	"LAW",
	"MED",
	"SPH",
	"CTY",
	"CSI",
	"NCC",
	"HOS",
	"HTR",
	"JJC",
	"KCC",
	"LAG",
	"LEH",
	"MEC",
	"NYT",
	"QNS",
	"QCC",
	"SPS",
	"GRD",
	"YRK",
}

func abbrToCunyUniversity(abbr string) CunyUniversity {
	for i, val := range cunyUniversityAbbr {
		if abbr == val {
			return CunyUniversity(i)
		}
	}

	return 0
}

func abbrMap() string {
	var buffer bytes.Buffer
	buffer.WriteString("\n")
	for i, val := range cunyUniversityAbbr {
		buffer.WriteString("\t")
		buffer.WriteString(val)
		buffer.WriteString("=")
		buffer.WriteString(cunyUniversity[i])
		buffer.WriteString("\n")
	}

	buffer.WriteString("\n")

	return buffer.String()
}
