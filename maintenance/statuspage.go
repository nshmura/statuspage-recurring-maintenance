package maintenance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const key_toolNamespace = "statuspage_register_tool"
const key_scheduleType = "scheduleType"
const key_scheduleKey = "scheduleKey"
const ScheduleType_Recurring = "recurring"

// Request body of creating Incident
type StatuspageCreateIncidentRequest struct {
	Incident StatuspageIncidentRequest `json:"incident"`
}

// Incident Data of creating Incident
type StatuspageIncidentRequest struct {
	Name                                      string                 `json:"name"`
	Status                                    string                 `json:"status"`
	ImpactOverride                            string                 `json:"impact_override"`
	ScheduledFor                              time.Time              `json:"scheduled_for"`
	ScheduledUntil                            time.Time              `json:"scheduled_until"`
	ScheduledRemindPrior                      bool                   `json:"scheduled_remind_prior"`
	ScheduledAutoInProgress                   bool                   `json:"scheduled_auto_in_progress"`
	ScheduledAutoCompleted                    bool                   `json:"scheduled_auto_completed"`
	Metadata                                  map[string]interface{} `json:"metadata"`
	DeliverNotifications                      bool                   `json:"deliver_notifications"`
	AutoTransitionDeliverNotificationsAtEnd   bool                   `json:"auto_transition_deliver_notifications_at_end"`
	AutoTransitionDeliverNotificationsAtStart bool                   `json:"auto_transition_deliver_notifications_at_start"`
	AutoTransitionToMaintenanceState          bool                   `json:"auto_transition_to_maintenance_state"`
	AutoTransitionToOperationalState          bool                   `json:"auto_transition_to_operational_state"`
	AutoTweetAtBeginning                      bool                   `json:"auto_tweet_at_beginning"`
	AutoTweetOnCompletion                     bool                   `json:"auto_tweet_on_completion"`
	AutoTweetOnCreation                       bool                   `json:"auto_tweet_on_creation"`
	AutoTweetOneHourBefore                    bool                   `json:"auto_tweet_one_hour_before"`
	BackfillDate                              interface{}            `json:"backfill_date"`
	Backfilled                                bool                   `json:"backfilled"`
	Body                                      string                 `json:"body"`
	Components                                map[string]interface{} `json:"components"`
	ComponentIds                              []string               `json:"component_ids"`
	ScheduledAutoTransition                   bool                   `json:"scheduled_auto_transition"`
}

// Incident Data of Statuspage
type StatuspageIncident struct {
	Id                      string                            `json:"id"`
	Components              []StatuspageComponnet             `json:"components"`
	CreatedAt               time.Time                         `json:"created_at"`
	Impact                  string                            `json:"impact"`
	ImpactOverride          string                            `json:"impact_override"`
	IncidentUpdates         []interface{}                     `json:"incident_updates"`
	Metadata                map[string]map[string]interface{} `json:"metadata"`
	MonitoringAt            time.Time                         `json:"monitoring_at"`
	Name                    string                            `json:"name"`
	PageId                  string                            `json:"page_id"`
	ResolvedAt              time.Time                         `json:"resolved_at"`
	ScheduledAutoCompleted  bool                              `json:"scheduled_auto_completed"`
	ScheduledAutoInProgress bool                              `json:"scheduled_auto_in_progress"`
	ScheduledFor            time.Time                         `json:"scheduled_for"`
	ScheduledRemindPrior    bool                              `json:"scheduled_remind_prior"`
	ScheduledRemindAt       time.Time                         `json:"scheduled_remind_at"`
	ScheduledUntil          time.Time                         `json:"scheduled_until"`
	Shortlink               string                            `json:"shortlink"`
	Status                  string                            `json:"status"`
	UpdatedAt               time.Time                         `json:"updated_at"`
}

// return true if StatuspageIncident is scheduled by this tool and is recurring schedule
func (incident StatuspageIncident) isRecurringSchedule() bool {
	data, ok := incident.Metadata[key_toolNamespace]
	if !ok {
		return false
	}

	if t, ok := data[key_scheduleType]; !ok || t == ScheduleType_Recurring {
		return true
	} else {
		return false
	}
}

// return true if StatuspageIncident has ids components
func (incident StatuspageIncident) isSameComponentIds(ids []string) bool {
	if len(ids) != len(incident.Components) {
		return false
	}

	for _, c := range incident.Components {
		found := false
		for _, id := range ids {
			if c.Id == id {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (incident StatuspageIncident) isSameTerm(start time.Time, end time.Time) bool {
	return start.Equal(incident.ScheduledFor) && end.Equal(incident.ScheduledUntil)
}

type StatuspageComponnet struct {
	Id                 string    `json:"id"`
	PageId             string    `json:"page_id"`
	GroupId            string    `json:"group_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Group              bool      `json:"group"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Position           int       `json:"position"`
	Status             string    `json:"status"`
	Showcase           bool      `json:"showcase"`
	OnlyShowIfDegraded bool      `json:"only_show_if_degraded"`
	AutomationEmail    string    `json:"automation_email"`
}

func CreateMaintenanceStatuspageData(
	title string,
	body string,
	componentIds []string,
	start time.Time,
	end time.Time,
	scheduleType string,
	scheduleKey string,
) StatuspageCreateIncidentRequest {
	return StatuspageCreateIncidentRequest{
		Incident: StatuspageIncidentRequest{
			Name:                    title,
			Status:                  "scheduled",
			ImpactOverride:          "maintenance",
			ScheduledFor:            start,
			ScheduledUntil:          end,
			ScheduledRemindPrior:    true,
			ScheduledAutoInProgress: true,
			ScheduledAutoCompleted:  true,
			Metadata: map[string]interface{}{
				key_toolNamespace: map[string]interface{}{
					"createdAt":      time.Now(),
					key_scheduleType: scheduleType,
					key_scheduleKey:  scheduleKey,
				},
			},
			DeliverNotifications:                      true,
			AutoTransitionDeliverNotificationsAtEnd:   true,
			AutoTransitionDeliverNotificationsAtStart: true,
			AutoTransitionToMaintenanceState:          true,
			AutoTransitionToOperationalState:          true,
			AutoTweetAtBeginning:                      false,
			AutoTweetOnCompletion:                     false,
			AutoTweetOnCreation:                       false,
			AutoTweetOneHourBefore:                    false,
			Backfilled:                                false,
			Body:                                      body,
			Components:                                map[string]interface{}{},
			ComponentIds:                              componentIds,
			ScheduledAutoTransition:                   true,
		},
	}
}

type StatuspageRepository interface {
	Add(data StatuspageCreateIncidentRequest) error
	FindAllScheduledIncidents(page int, perPage int) ([]StatuspageIncident, error)
	Delete(incident StatuspageIncident) error
}

// DryRun Repository
type StatuspageDryRunRepository struct {
	statuspageRESTClient StatuspageRESTClient
}

func createStatuspageDryRunRepository(pageId string, accessToken string) *StatuspageDryRunRepository {
	return &StatuspageDryRunRepository{
		statuspageRESTClient: StatuspageRESTClient{PageId: pageId, AccessToken: accessToken},
	}
}

func (s *StatuspageDryRunRepository) Add(data StatuspageCreateIncidentRequest) error {
	fmt.Printf("[dryRun]: add [%s] %s - %s\n", data.Incident.Name, data.Incident.ScheduledFor, data.Incident.ScheduledUntil)
	return nil
}

func (s *StatuspageDryRunRepository) Delete(incident StatuspageIncident) error {
	fmt.Printf("[dryRun]: delete [%s] %s - %s id:%s\n", incident.Name, incident.ScheduledFor, incident.ScheduledUntil, incident.Id)
	return nil
}

func (s *StatuspageDryRunRepository) FindAllScheduledIncidents(page int, perPage int) ([]StatuspageIncident, error) {
	return s.statuspageRESTClient.FindAllScheduledIncidents(page, perPage)
}

// REST Repository
type StatuspageRESTRepository struct {
	statuspageRESTClient StatuspageRESTClient
}

func createStatuspageRESTRepository(pageId string, accessToken string) *StatuspageRESTRepository {
	return &StatuspageRESTRepository{
		statuspageRESTClient: StatuspageRESTClient{PageId: pageId, AccessToken: accessToken},
	}
}

func (s *StatuspageRESTRepository) Add(data StatuspageCreateIncidentRequest) error {
	fmt.Printf("add [%s] %s - %s\n", data.Incident.Name, data.Incident.ScheduledFor, data.Incident.ScheduledUntil)
	return s.statuspageRESTClient.Add(data)
}

func (s *StatuspageRESTRepository) Delete(incident StatuspageIncident) error {
	fmt.Printf("[dryRun]: delete [%s] %s - %s id:%s\n", incident.Name, incident.ScheduledFor, incident.ScheduledUntil, incident.Id)
	return s.statuspageRESTClient.Delete(incident.Id)
}

func (s *StatuspageRESTRepository) FindAllScheduledIncidents(page int, perPage int) ([]StatuspageIncident, error) {
	return s.statuspageRESTClient.FindAllScheduledIncidents(page, perPage)
}

// StatuspageRESTClient
type StatuspageRESTClient struct {
	PageId      string
	AccessToken string
}

func (s *StatuspageRESTClient) Add(data StatuspageCreateIncidentRequest) error {
	url := fmt.Sprintf("https://api.statuspage.io/v1/pages/%s/incidents", s.PageId)
	bearer := "OAuth " + s.AccessToken

	body, _ := json.Marshal(&data)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 201 {
		return errors.New(string(respBody))
	}

	return nil
}

func (s *StatuspageRESTClient) Delete(incidentId string) error {
	url := fmt.Sprintf("https://api.statuspage.io/v1/pages/%s/incidents/%s", s.PageId, incidentId)
	bearer := "OAuth " + s.AccessToken

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New(string(respBody))
	}

	return nil
}

func (s *StatuspageRESTClient) FindAllScheduledIncidents(page int, perPage int) ([]StatuspageIncident, error) {
	url := fmt.Sprintf("https://api.statuspage.io/v1/pages/%s/incidents/scheduled?page=%d&per_page=%d", s.PageId, page, perPage)
	bearer := "OAuth " + s.AccessToken

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New(string(body))
	}

	var incidents []StatuspageIncident
	err = json.Unmarshal(body, &incidents)
	if err != nil {
		return nil, err
	}
	return incidents, nil
}
