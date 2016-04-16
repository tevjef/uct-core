package main

type MyJsonName struct {
	Data struct {
		CampusDescription     string    `json:"campusDescription"`
		CourseNumber          string    `json:"courseNumber"`
		CourseReferenceNumber string    `json:"courseReferenceNumber"`
		CourseTitle           string    `json:"courseTitle"`
		CreditHourHigh        string    `json:"creditHourHigh"`
		CreditHourIndicator   string    `json:"creditHourIndicator"`
		CreditHourLow         int       `json:"creditHourLow"`
		CreditHours           string    `json:"creditHours"`
		CrossList             string    `json:"crossList"`
		CrossListAvailable    string    `json:"crossListAvailable"`
		CrossListCapacity     string    `json:"crossListCapacity"`
		CrossListCount        string    `json:"crossListCount"`
		Enrollment            int       `json:"enrollment"`
		Faculty               []Faculty `json:"faculty"`
		ID                    int       `json:"id"`
		IsSectionLinked       bool      `json:"isSectionLinked"`
		LinkIdentifier        string    `json:"linkIdentifier"`
		MaximumEnrollment     int       `json:"maximumEnrollment"`
		MeetingsFaculty       []struct {
			Category              string `json:"category"`
			Class                 string `json:"class"`
			CourseReferenceNumber string `json:"courseReferenceNumber"`
			Faculty               []struct {
				BannerID              string `json:"bannerId"`
				Category              string `json:"category"`
				Class                 string `json:"class"`
				CourseReferenceNumber string `json:"courseReferenceNumber"`
				DisplayName           string `json:"displayName"`
				EmailAddress          string `json:"emailAddress"`
				PrimaryIndicator      bool   `json:"primaryIndicator"`
				Term                  string `json:"term"`
			} `json:"faculty"`
			MeetingTime struct {
				BeginTime             string `json:"beginTime"`
				Building              string `json:"building"`
				BuildingDescription   string `json:"buildingDescription"`
				Campus                string `json:"campus"`
				CampusDescription     string `json:"campusDescription"`
				Category              string `json:"category"`
				Class                 string `json:"class"`
				CourseReferenceNumber string `json:"courseReferenceNumber"`
				CreditHourSession     int    `json:"creditHourSession"`
				EndDate               string `json:"endDate"`
				EndTime               string `json:"endTime"`
				Friday                bool   `json:"friday"`
				HoursWeek             int    `json:"hoursWeek"`
				MeetingScheduleType   string `json:"meetingScheduleType"`
				Monday                bool   `json:"monday"`
				Room                  string `json:"room"`
				Saturday              bool   `json:"saturday"`
				StartDate             string `json:"startDate"`
				Sunday                bool   `json:"sunday"`
				Term                  string `json:"term"`
				Thursday              bool   `json:"thursday"`
				Tuesday               bool   `json:"tuesday"`
				Wednesday             bool   `json:"wednesday"`
			} `json:"meetingTime"`
			Term string `json:"term"`
		} `json:"meetingsFaculty"`
		OpenSection             bool   `json:"openSection"`
		PartOfTerm              string `json:"partOfTerm"`
		ScheduleTypeDescription string `json:"scheduleTypeDescription"`
		SeatsAvailable          int    `json:"seatsAvailable"`
		SequenceNumber          string `json:"sequenceNumber"`
		Subject                 string `json:"subject"`
		SubjectCourse           string `json:"subjectCourse"`
		SubjectDescription      string `json:"subjectDescription"`
		Term                    string `json:"term"`
		TermDesc                string `json:"termDesc"`
		WaitAvailable           int    `json:"waitAvailable"`
		WaitCapacity            int    `json:"waitCapacity"`
		WaitCount               int    `json:"waitCount"`
	} `json:"data"`
	PageMaxSize          int    `json:"pageMaxSize"`
	PageOffset           int    `json:"pageOffset"`
	PathMode             string `json:"pathMode"`
	SectionsFetchedCount int    `json:"sectionsFetchedCount"`
	Success              bool   `json:"success"`
	TotalCount           int    `json:"totalCount"`

	Faculty struct {
		BannerID              string `json:"bannerId"`
		Category              string `json:"category"`
		Class                 string `json:"class"`
		CourseReferenceNumber string `json:"courseReferenceNumber"`
		DisplayName           string `json:"displayName"`
		EmailAddress          string `json:"emailAddress"`
		PrimaryIndicator      bool   `json:"primaryIndicator"`
		Term                  string `json:"term"`
	}
}
