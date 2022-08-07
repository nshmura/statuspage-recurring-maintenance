package maintenance

import (
	"testing"
	"time"
)

func dateOf(year int, month time.Month, day int) time.Time {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func timeOf(year int, month time.Month, day, hour, min int) time.Time {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	return time.Date(year, month, day, hour, min, 0, 0, loc)
}

func TestCreateSchedule(t *testing.T) {
	patterns := []struct {
		schedule    RecurringCommand       // input
		maintenance []RecurringMaintenance // input
		exp         []ScheduledTerm        // expected
	}{
		// simple days
		{
			RecurringCommand{
				FromDate: dateOf(2020, 1, 1),
				ToDate:   dateOf(2020, 1, 2),
			},
			[]RecurringMaintenance{
				{
					Service: "test",
					RecurringSchedules: []RecurringSchedules{
						{
							Day:   "everyday",
							Start: "23h50m",
							Time:  "20m",
						},
					},
				},
			},
			[]ScheduledTerm{
				{
					Service: "test",
					Start:   timeOf(2020, 1, 1, 23, 50),
					End:     timeOf(2020, 1, 2, 0, 10),
				},
				{
					Service: "test",
					Start:   timeOf(2020, 1, 2, 23, 50),
					End:     timeOf(2020, 1, 3, 0, 10),
				},
			},
		},

		// overlapped days
		{
			RecurringCommand{
				FromDate: dateOf(2020, 1, 1),
				ToDate:   dateOf(2020, 1, 3),
			},
			[]RecurringMaintenance{
				{
					Service: "test",
					RecurringSchedules: []RecurringSchedules{
						{
							Day:   "everyday",
							Start: "23h50m",
							Time:  "20m",
						},
						{
							Day:   "1st thursday",
							Start: "22h00m",
							Time:  "2h",
						},
					},
				},
			},
			[]ScheduledTerm{
				{
					Service: "test",
					Start:   timeOf(2020, 1, 1, 23, 50),
					End:     timeOf(2020, 1, 2, 0, 10),
				},
				{
					Service: "test",
					Start:   timeOf(2020, 1, 2, 22, 00),
					End:     timeOf(2020, 1, 3, 0, 10),
				},
				{
					Service: "test",
					Start:   timeOf(2020, 1, 3, 23, 50),
					End:     timeOf(2020, 1, 4, 0, 10),
				},
			},
		},

		// 2 Services
		{
			RecurringCommand{
				FromDate: dateOf(2020, 1, 1),
				ToDate:   dateOf(2020, 1, 2),
			},
			[]RecurringMaintenance{
				{
					Service: "test1",
					RecurringSchedules: []RecurringSchedules{
						{
							Day:   "everyday",
							Start: "23h50m",
							Time:  "20m",
						},
					},
				},
				{
					Service: "test2",
					RecurringSchedules: []RecurringSchedules{
						{
							Day:   "1st wednesday",
							Start: "10h00m",
							Time:  "20m",
						},
					},
				},
			},
			[]ScheduledTerm{
				{
					Service: "test1",
					Start:   timeOf(2020, 1, 1, 23, 50),
					End:     timeOf(2020, 1, 2, 0, 10),
				},
				{
					Service: "test1",
					Start:   timeOf(2020, 1, 2, 23, 50),
					End:     timeOf(2020, 1, 3, 0, 10),
				},
				{
					Service: "test2",
					Start:   timeOf(2020, 1, 1, 10, 00),
					End:     timeOf(2020, 1, 1, 10, 20),
				},
			},
		},
	}

	for idx, row := range patterns {
		result := row.schedule.CreateSchedule(row.maintenance)

		if len(row.exp) != len(result) {
			t.Errorf(`test(%v): %v.CreateSchedule(%v)'s len is %v. exp is %v`,
				idx+1, row.schedule, row.maintenance, len(result), len(row.exp))
			continue
		}

		for i := 0; i < len(row.exp); i++ {
			if row.exp[i].Service != result[i].Service {
				t.Errorf(`test(%v): %v.CreateSchedule(%v) [idx:%v]'s Service is %v. exp is %v`,
					idx+1, row.schedule, row.maintenance, i+1, result[i].Service, row.exp[i].Service)
			}
			if !row.exp[i].Start.Equal(result[i].Start) {
				t.Errorf(`test(%v): %v.CreateSchedule(%v) [idx:%v]'s Start is %v. exp is %v`,
					idx+1, row.schedule, row.maintenance, i+1, result[i].Start, row.exp[i].Start)
			}
			if !row.exp[i].End.Equal(result[i].End) {
				t.Errorf(`test(%v): %v.CreateSchedule(%v) [idx:%v]'s End is %v. exp is %v`,
					idx+1, row.schedule, row.maintenance, i+1, result[i].End, row.exp[i].End)
			}
		}
	}
}

func TestIsCreateTerms(t *testing.T) {
	patterns := []struct {
		schedule RecurringSchedules // input
		start    time.Time          // input
		end      time.Time          // input
		exp      []*Term            // expected
	}{
		// everyday
		{
			RecurringSchedules{
				Day:   "everyday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 1),
			dateOf(2020, 1, 2),
			[]*Term{
				{
					timeOf(2020, 1, 1, 23, 50),
					timeOf(2020, 1, 2, 0, 10),
				},
				{
					timeOf(2020, 1, 2, 23, 50),
					timeOf(2020, 1, 3, 0, 10),
				},
			},
		},

		// out of range
		{
			RecurringSchedules{
				Day:   "1st sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 1),
			dateOf(2020, 1, 1),
			[]*Term{},
		},

		// 1st sunday
		{
			RecurringSchedules{
				Day:   "1st sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 1),
			dateOf(2020, 1, 6),
			[]*Term{
				{
					timeOf(2020, 1, 5, 23, 50),
					timeOf(2020, 1, 6, 0, 10),
				},
			},
		},

		// every sunday
		{
			RecurringSchedules{
				Day:   "every sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 1),
			dateOf(2020, 2, 29),
			[]*Term{
				{
					timeOf(2020, 1, 5, 23, 50),
					timeOf(2020, 1, 6, 0, 10),
				},
				{
					timeOf(2020, 1, 12, 23, 50),
					timeOf(2020, 1, 13, 0, 10),
				},
				{
					timeOf(2020, 1, 19, 23, 50),
					timeOf(2020, 1, 20, 0, 10),
				},
				{
					timeOf(2020, 1, 26, 23, 50),
					timeOf(2020, 1, 27, 0, 10),
				},
				{
					timeOf(2020, 2, 2, 23, 50),
					timeOf(2020, 2, 3, 0, 10),
				},
				{
					timeOf(2020, 2, 9, 23, 50),
					timeOf(2020, 2, 10, 0, 10),
				},
				{
					timeOf(2020, 2, 16, 23, 50),
					timeOf(2020, 2, 17, 0, 10),
				},
				{
					timeOf(2020, 2, 23, 23, 50),
					timeOf(2020, 2, 24, 0, 10),
				},
			},
		},
	}

	for idx, row := range patterns {
		result := row.schedule.CreateTerms(row.start, row.end)

		if len(row.exp) != len(result) {
			t.Errorf(`test(%v): %v.CreateTerms(%v, %v)'s len is %v. exp is %v`,
				idx+1, row.schedule, row.start, row.end, len(result), len(row.exp))
			continue
		}

		for i := 0; i < len(row.exp); i++ {
			if !row.exp[i].Start.Equal(result[i].Start) {
				t.Errorf(`test(%v): %v.CreateTerms(%v, %v)'s Start is %v. exp is %v`,
					idx+1, row.schedule, row.start, row.end, result[i].Start, row.exp[i].Start)
			}
			if !row.exp[i].End.Equal(result[i].End) {
				t.Errorf(`test(%v): %v.CreateTerms(%v, %v)'s End is %v. exp is %v`,
					idx+1, row.schedule, row.start, row.end, result[i].End, row.exp[i].End)
			}
		}
	}
}

func TestIsMaintenanceDay(t *testing.T) {

	patterns := []struct {
		exp      bool               // expected
		schedule RecurringSchedules // input
		time     time.Time          // input
	}{
		// everyday
		{
			true,
			RecurringSchedules{
				Day:   "everyday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 1),
		},

		// 1st sunday
		{
			true,
			RecurringSchedules{
				Day:   "1st sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 5),
		},

		// 1st monday
		{
			true,
			RecurringSchedules{
				Day:   "1st monday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 6),
		},

		// 1st tuesday
		{
			true,
			RecurringSchedules{
				Day:   "1st tuesday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 7),
		},

		// 1st wednesday
		{
			true,
			RecurringSchedules{
				Day:   "1st wednesday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 1),
		},

		// 1st thursday
		{
			true,
			RecurringSchedules{
				Day:   "1st thursday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 2),
		},

		// 1st friday
		{
			true,
			RecurringSchedules{
				Day:   "1st friday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 3),
		},

		// 1st saturday
		{
			true,
			RecurringSchedules{
				Day:   "1st saturday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 4),
		},

		// 2nd sunday
		{
			true,
			RecurringSchedules{
				Day:   "2nd sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 12),
		},

		// 3rd sunday
		{
			true,
			RecurringSchedules{
				Day:   "3rd sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 19),
		},

		// 4th sunday
		{
			true,
			RecurringSchedules{
				Day:   "4th sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 26),
		},

		// invalid day
		{
			false,
			RecurringSchedules{
				Day:   "1st sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 4),
		},

		// invalid day
		{
			false,
			RecurringSchedules{
				Day:   "1st sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 6),
		},

		// 2 month
		{
			true,
			RecurringSchedules{
				Day:   "1st sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 2, 2),
		},

		// 12 month
		{
			true,
			RecurringSchedules{
				Day:   "1st tuesday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 12, 1),
		},

		// 12/31
		{
			true,
			RecurringSchedules{
				Day:   "5th thursday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 12, 31),
		},

		// every sunday
		{
			true,
			RecurringSchedules{
				Day:   "every sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 5),
		},
		{
			true,
			RecurringSchedules{
				Day:   "every sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 12),
		},
		{
			true,
			RecurringSchedules{
				Day:   "every sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 19),
		},
		{
			true,
			RecurringSchedules{
				Day:   "every sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 1, 26),
		},
		{
			true,
			RecurringSchedules{
				Day:   "every sunday",
				Start: "23h50m",
				Time:  "20m",
			},
			dateOf(2020, 2, 2),
		},
	}

	for idx, row := range patterns {
		if row.exp != row.schedule.IsMaintenanceDay(row.time) {
			t.Errorf(`test(%v): %v.IsMaintenanceDay(%v) is %v`, idx+1, row.schedule, row.time, !row.exp)
		}
	}
}

func TestAdjustIncients(t *testing.T) {
	commmand := RecurringCommand{
		FromDate: dateOf(2020, 1, 1),
		ToDate:   dateOf(2020, 1, 5),
	}

	config := StatuspageConfig{
		StatuspagePageId: "testPageId",
		StatuspageServices: []StatuspageService{
			{
				Service:      "testService",
				ComponentIds: []string{"testComponentId"},
			},
			{
				Service:      "testService2",
				ComponentIds: []string{"testComponentId1", "testComponentId2"},
			},
		},
	}

	patterns := []struct {
		incidents     []StatuspageIncident // input
		schduledTerms []ScheduledTerm      // input
		exp           []ScheduledTerm      // expected
	}{
		// overlapped && covered
		{
			[]StatuspageIncident{
				{
					Id: "1",
					Components: []StatuspageComponnet{
						{
							Id: "testComponentId",
						},
					},
					ScheduledFor:   timeOf(2020, 1, 1, 1, 0),
					ScheduledUntil: timeOf(2020, 1, 1, 1, 10),
				},
			},
			[]ScheduledTerm{
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 1, 0),
					End:     timeOf(2020, 1, 1, 1, 10),
				},
			},
			[]ScheduledTerm{},
		},

		// overlapped && covered
		{
			[]StatuspageIncident{
				{
					Id: "1",
					Components: []StatuspageComponnet{
						{
							Id: "testComponentId",
						},
					},
					ScheduledFor:   timeOf(2020, 1, 1, 1, 0),
					ScheduledUntil: timeOf(2020, 1, 1, 1, 10),
				},
			},
			[]ScheduledTerm{
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 1, 1),
					End:     timeOf(2020, 1, 1, 1, 9),
				},
			},
			[]ScheduledTerm{},
		},

		// overlapped && !covered
		{
			[]StatuspageIncident{
				{
					Id:     "1",
					PageId: "page1",
					Components: []StatuspageComponnet{
						{
							Id: "testComponentId",
						},
					},
					ScheduledFor:   timeOf(2020, 1, 1, 1, 0),
					ScheduledUntil: timeOf(2020, 1, 1, 1, 10),
				},
			},
			[]ScheduledTerm{
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 1, 5),
					End:     timeOf(2020, 1, 1, 1, 15),
				},
			},
			[]ScheduledTerm{},
		},

		// overlapped && !covered
		{
			[]StatuspageIncident{
				{
					Id:     "1",
					PageId: "page1",
					Components: []StatuspageComponnet{
						{
							Id: "testComponentId",
						},
					},
					ScheduledFor:   timeOf(2020, 1, 1, 1, 0),
					ScheduledUntil: timeOf(2020, 1, 1, 1, 10),
				},
			},
			[]ScheduledTerm{
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 1, 10),
					End:     timeOf(2020, 1, 1, 1, 15),
				},
			},
			[]ScheduledTerm{},
		},

		// !overlapped
		{
			[]StatuspageIncident{
				{
					Id:     "1",
					PageId: "page1",
					Components: []StatuspageComponnet{
						{
							Id: "testComponentId",
						},
					},
					ScheduledFor:   timeOf(2020, 1, 1, 1, 0),
					ScheduledUntil: timeOf(2020, 1, 1, 1, 11),
				},
			},
			[]ScheduledTerm{
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 1, 12),
					End:     timeOf(2020, 1, 1, 1, 15),
				},
			},
			[]ScheduledTerm{
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 1, 12),
					End:     timeOf(2020, 1, 1, 1, 15),
				},
			},
		},

		// not a same component
		{
			[]StatuspageIncident{
				{
					Id:     "1",
					PageId: "page1",
					Components: []StatuspageComponnet{
						{
							Id: "testComponentId2",
						},
					},
					ScheduledFor:   timeOf(2020, 1, 1, 1, 0),
					ScheduledUntil: timeOf(2020, 1, 1, 1, 10),
				},
			},
			[]ScheduledTerm{
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 1, 0),
					End:     timeOf(2020, 1, 1, 1, 10),
				},
			},
			[]ScheduledTerm{
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 1, 0),
					End:     timeOf(2020, 1, 1, 1, 10),
				},
			},
		},

		// many copmonents
		{
			[]StatuspageIncident{
				{
					Id:     "1",
					PageId: "page1",
					Components: []StatuspageComponnet{
						{
							Id: "testComponentId",
						},
					},
					ScheduledFor:   timeOf(2020, 1, 1, 1, 0),
					ScheduledUntil: timeOf(2020, 1, 1, 1, 10),
				},
				{
					Id: "2",
					Components: []StatuspageComponnet{
						{
							Id: "testComponentId1",
						},
						{
							Id: "testComponentId2",
						},
					},
					ScheduledFor:   timeOf(2020, 1, 1, 1, 0),
					ScheduledUntil: timeOf(2020, 1, 1, 1, 10),
				},
			},
			[]ScheduledTerm{
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 1, 0),
					End:     timeOf(2020, 1, 1, 1, 10),
				},
				{
					Service: "testService2",
					Start:   timeOf(2020, 1, 1, 1, 0),
					End:     timeOf(2020, 1, 1, 1, 10),
				},
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 2, 0),
					End:     timeOf(2020, 1, 1, 2, 10),
				},
			},
			[]ScheduledTerm{
				{
					Service: "testService",
					Start:   timeOf(2020, 1, 1, 2, 0),
					End:     timeOf(2020, 1, 1, 2, 10),
				},
			},
		},
	}

	for idx, row := range patterns {
		actual := commmand.adjustIncients(row.incidents, row.schduledTerms, config)
		if len(actual) != len(row.exp) {
			t.Errorf("test(%v): exp:%v, actual:%v", idx+1, row.exp, actual)
		} else {
			for i := 0; i < len(row.exp); i++ {
				if !row.exp[i].Start.Equal(actual[i].Start) {
					t.Errorf("test(%v): exp:%v, actual:%v", idx+1, row.exp, actual)
				} else if !row.exp[i].End.Equal(actual[i].End) {
					t.Errorf("test(%v): exp:%v, actual:%v", idx+1, row.exp, actual)
				}
			}
		}
	}
}
