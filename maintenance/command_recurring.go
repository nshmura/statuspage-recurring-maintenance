package maintenance

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

type RecurringCommand struct {
	ScheduleFilename string
	FromDate         time.Time
	ToDate           time.Time

	isDryRun           bool
	StatuspageFilename string
	AccessToken        string
}

type RecurringMaintenance struct {
	Service            string               `yaml:"service"`
	Title              string               `yaml:"title"`
	Body               string               `yaml:"body"`
	RecurringSchedules []RecurringSchedules `yaml:"recurring"`
}

type RecurringSchedules struct {
	Day   string `yaml:"day"`
	Start string `yaml:"start"`
	Time  string `yaml:"time"`
}

type Term struct {
	Start time.Time
	End   time.Time
}

type ScheduledTerm struct {
	Service string
	Title   string
	Body    string
	Start   time.Time
	End     time.Time
}

const everyOrdinal = -1

// execute recurring command
func (c *RecurringCommand) Run() {

	maintenances := make([]RecurringMaintenance, 1)
	loadFromFile(c.ScheduleFilename, &maintenances)

	statuspageConfig := StatuspageConfig{}
	loadFromFile(c.StatuspageFilename, &statuspageConfig)

	scheduledTerms := c.CreateSchedule(maintenances)

	repository := c.getStatuspageRepository(statuspageConfig.StatuspagePageId)

	incidents, err := repository.FindAllScheduledIncidents(1, 200)
	if err != nil {
		log.Fatalf("[ERORR] FindAllScheduledIncidents err: %s", err.Error())
	}
	if len(incidents) >= 200 {
		log.Fatalf("[ERORR] too many incidents are registered: %d", len(incidents))
	}
	scheduledTerms = c.adjustIncients(incidents, scheduledTerms, statuspageConfig)

	c.deleteIncidents(repository, incidents, statuspageConfig, scheduledTerms)
	c.registerIncidents(repository, incidents, statuspageConfig, scheduledTerms)
}

//-------------------------------
// Create schedule of maintenance
//-------------------------------

func (c *RecurringCommand) CreateSchedule(maintenances []RecurringMaintenance) []ScheduledTerm {

	scheduledTerms := make([]ScheduledTerm, 0)
	for _, ps := range maintenances {
		terms := make([]*Term, 0)
		for _, s := range ps.RecurringSchedules {
			terms = c.margeTerms(terms, s.CreateTerms(c.FromDate, c.ToDate))
		}
		for _, t := range terms {
			scheduledTerms = append(scheduledTerms, ScheduledTerm{
				Service: ps.Service,
				Start:   t.Start,
				End:     t.End,
				Title:   ps.Title,
				Body:    ps.Body,
			})
		}
	}
	return scheduledTerms
}

// Marge overlapped terms
//
// e.g.　"2020/1/1 10:00:00 〜 11:00:00" and "2020/1/1 10:30:00 〜 11:30:00" will be marged to be "2020/1/1 10:00:00 〜 11:30:00
//
func (c *RecurringCommand) margeTerms(terms1 []*Term, terms2 []*Term) []*Term {
	for _, m1 := range terms2 {
		merged := false

		for _, m2 := range terms1 {
			if m1.Start.Before(m2.End) && m1.End.After(m2.Start) {
				if m1.Start.Before(m2.Start) {
					m2.Start = m1.Start
				}
				if m2.End.Before(m1.End) {
					m2.End = m1.End
				}
				merged = true
				break
			}
		}

		if !merged {
			terms1 = append(terms1, m1)
		}
	}

	return terms1
}

func (s *RecurringSchedules) CreateTerms(firstDate time.Time, lastDate time.Time) []*Term {
	terms := make([]*Term, 0)

	until := lastDate.AddDate(0, 0, 1)

	for date := firstDate; date.Before(until); date = date.AddDate(0, 0, 1) {
		if s.IsMaintenanceDay(date) {
			start, _ := time.ParseDuration(s.Start)
			t, _ := time.ParseDuration(s.Time)

			terms = append(
				terms,
				&Term{
					Start: date.Add(start),
					End:   date.Add(start).Add(t),
				},
			)
		}
	}
	return terms
}

func (s *RecurringSchedules) IsMaintenanceDay(base time.Time) bool {
	if s.isEveryDay() {
		return true
	}

	ordinal, weekday, err := s.weekday()
	if err != nil {
		panic(err)
	}

	if (ordinal == everyOrdinal || ordinal == s.ordinalOfWeekday(base)) &&
		weekday == strings.ToLower(base.Weekday().String()) {
		return true
	}
	return false
}

func (s *RecurringSchedules) isEveryDay() bool {
	return strings.ToLower(s.Day) == "everyday"
}

// return xth weekday
//
// e.g. 2020/1/1 -> 1,wednesday
// e.g. 2020/12/31 -> 5,thursday
//
func (s *RecurringSchedules) weekday() (ordinal int, weekday string, err error) {
	r := regexp.MustCompile(`^\s*(1st|2nd|3rd|[4,5]+th|every)\s+(sunday|monday|tuesday|wednesday|thursday|friday|saturday)\s*$`)
	result := r.FindAllStringSubmatch(s.Day, -1)

	// log.Printf("%v %v", s.Day, result)

	o := result[0][1]
	w := result[0][2]

	if strings.ToLower(o) == "every" {
		return everyOrdinal, w, nil
	}
	if strings.ToLower(o) == "1st" {
		return 1, w, nil
	}
	if strings.ToLower(o) == "2nd" {
		return 2, w, nil
	}
	if strings.ToLower(o) == "3rd" {
		return 3, w, nil
	}
	if strings.ToLower(o) == "4th" {
		return 4, w, nil
	}
	if strings.ToLower(o) == "5th" {
		return 5, w, nil
	}
	return -1, "", fmt.Errorf("invalid weekday: %v", s.Day)
}

// return xth of weekday
//
// e.g. 2020/1/1 -> 1
// e.g. 2020/12/31 -> 5
//
func (s *RecurringSchedules) ordinalOfWeekday(t time.Time) int {
	y := t.Year()
	m := t.Month()

	target := time.Date(y, m, t.Day(), 0, 0, 0, 0, t.Location())
	first := time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
	last := first.AddDate(0, 1, -1)
	lastDayOfMonth := last.Day()

	ordinal := 0
	for d := 1; d <= lastDayOfMonth; d++ {
		date := time.Date(y, m, d, 0, 0, 0, 0, t.Location())

		if date.Weekday() == target.Weekday() {
			ordinal++
		}
		if date == target {
			return ordinal
		}
	}

	log.Fatalf("Fail ordinalOfWeekday: %v", t)
	return -1
}

//-------------------------------
// Adjust schedule of maintenance
//-------------------------------

// If the schedule of maintenace which will be created by this tool is
//  overlapped with another schedule of maintenaces which had been creaated by other way,
// the schedule of maintenace is modified not to be overlapped to another schedule of maintenances.
func (c *RecurringCommand) adjustIncients(
	incidents []StatuspageIncident,
	schedules []ScheduledTerm,
	config StatuspageConfig,
) []ScheduledTerm {
	newSchedules := make([]ScheduledTerm, 0)
	for _, s := range schedules {
		overlaped := false
		for _, i := range incidents {
			if i.isRecurringSchedule() || !s.hasSameComponent(i, config) {
				continue
			}
			if s.isOverlapped(i) {
				overlaped = true

				if s.isCovered(i) {
					fmt.Printf("skip: [%s] maintenance(%s - %s) is converted by existsing maintenance(%s - %s)\n", s.Service, s.Start, s.End, i.ScheduledFor, i.ScheduledUntil)
				} else {
					fmt.Printf(
						"alert: %s(https://manage.statuspage.io/pages/%s/incidents/%s) should be modified to contain %s 〜 %s\n",
						i.Name, i.PageId, i.Id, s.Start.Format("2006-01-02 15:04"), s.End.Format("2006-01-02 15:04"),
					)
				}
				break
			}
		}
		if !overlaped {
			newSchedules = append(newSchedules, s)
		}
	}
	return newSchedules
}

func (s *ScheduledTerm) hasSameComponent(incident StatuspageIncident, config StatuspageConfig) bool {
	component := config.findComponentByServiceName(s.Service)
	if component == nil {
		log.Fatalf("unkown service is found: %s", s.Service)
	}
	if incident.isSameComponentIds(component.ComponentIds) {
		return true
	}
	return false
}

func (s ScheduledTerm) isOverlapped(i StatuspageIncident) bool {
	return (i.ScheduledFor.Equal(s.End) || i.ScheduledFor.Before(s.End)) &&
		(i.ScheduledUntil.Equal(s.Start) || i.ScheduledUntil.After(s.Start))
}

func (s ScheduledTerm) isCovered(i StatuspageIncident) bool {
	return (i.ScheduledFor.Equal(s.Start) || i.ScheduledFor.Before(s.Start)) &&
		(i.ScheduledUntil.Equal(s.End) || i.ScheduledUntil.After(s.End))
}

//-------------------------------
// Register schedule of maintenancee
//-------------------------------

func (c *RecurringCommand) getStatuspageRepository(pageId string) StatuspageRepository {
	if c.isDryRun {
		return createStatuspageDryRunRepository(pageId, c.AccessToken)
	} else {
		return createStatuspageRESTRepository(pageId, c.AccessToken)
	}
}

func (c *RecurringCommand) deleteIncidents(
	repository StatuspageRepository,
	incidents []StatuspageIncident,
	config StatuspageConfig,
	schedules []ScheduledTerm,
) {
	toBeDeleted := make([]StatuspageIncident, 0)
	for _, i := range incidents {
		exsists := false
		for _, s := range schedules {
			if c.existsSameIncident([]StatuspageIncident{i}, config, s) {
				exsists = true
				break
			}
		}

		if !exsists &&
			i.isRecurringSchedule() &&
			i.ScheduledFor.After(c.FromDate) &&
			i.ScheduledFor.Before(c.ToDate) {
			toBeDeleted = append(toBeDeleted, i)
		}
	}
	for _, i := range toBeDeleted {
		err := repository.Delete(i)
		if err != nil {
			log.Fatalf("failed to deleteIncidents: %v", err)
		}
	}
}

func (c *RecurringCommand) registerIncidents(
	repository StatuspageRepository,
	incidents []StatuspageIncident,
	config StatuspageConfig,
	schedules []ScheduledTerm,
) {
	toBeRegistered := make([]ScheduledTerm, 0)
	for _, s := range schedules {
		if !c.existsSameIncident(incidents, config, s) &&
			s.Start.After(c.FromDate) &&
			s.Start.Before(c.ToDate.AddDate(0, 0, 1)) &&
			s.Start.After(time.Now()) {
			toBeRegistered = append(toBeRegistered, s)
		} else {
			fmt.Printf("skip: [%s] %s - %s\n", s.Service, s.Start, s.End)
		}
	}

	for _, s := range toBeRegistered {
		component := config.findComponentByServiceName(s.Service)
		if component == nil {
			log.Fatalf("unkown service is found: %s", s.Service)
		}

		err := repository.Add(CreateMaintenanceStatuspageData(
			s.Title,
			s.Body,
			component.ComponentIds,
			s.Start,
			s.End,
			ScheduleType_Recurring,
			"",
		))
		if err != nil {
			log.Fatalf("[ERORR] %s", err.Error())
		}

		if !c.isDryRun {
			// {“error”:“Too many requests, enhance your calm”} 対応
			time.Sleep(time.Second * 5)
		}
	}
}

func (c *RecurringCommand) existsSameIncident(incidents []StatuspageIncident, config StatuspageConfig, schedule ScheduledTerm) bool {
	component := config.findComponentByServiceName(schedule.Service)
	if component == nil {
		log.Fatalf("unkown service is found: %s", schedule.Service)
	}

	for _, incident := range incidents {
		if incident.isSameComponentIds(component.ComponentIds) &&
			incident.isSameTerm(schedule.Start, schedule.End) {
			return true
		}
	}

	return false
}
