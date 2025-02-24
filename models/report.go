// SPDX-FileCopyrightText: 2023 froggie <legal@frogg.ie>
//
// SPDX-License-Identifier: OSL-3.0

package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/effectindex/tripreporter/types"
	"github.com/georgysavva/scany/v2/pgxscan"
)

type Report struct {
	types.Context
	Unique
	Account      uuid.UUID      `json:"account_id" db:"account_id"`       // References the account that created this report.
	User         UserPublic     `json:"user"`                             // References the user belonging to Account.
	Created      Timestamp      `json:"creation_time" db:"creation_time"` // Required, set when creating a report.
	LastModified Timestamp      `json:"modified_time" db:"modified_time"` // Required, defaults to Created and set when modifying a report.
	Title        string         `json:"title" db:"title"`                 // Required.
	Date         Timestamp      `json:"report_date" db:"report_date"`     // Optional.
	Setting      string         `json:"setting,omitempty" db:"setting"`   // Optional.
	Sources      ReportSources  `json:"report_sources,omitempty"`         // Optional. Saved in the report_sources table,  appended manually.
	Subject      *ReportSubject `json:"report_subject,omitempty"`         // Optional. Saved in the report_subjects table, appended manually.
	Events       ReportEvents   `json:"report_events,omitempty"`          // Optional. Saved in the report_events table, appended manually.
	//Effects      []Effect  // TODO: #118
}

func (r *Report) Get() (*Report, error) {
	r.InitType(r)
	db := r.DB()

	if r.NilUUID() {
		return r, types.ErrorReportNotSpecified
	}

	var r1 []*Report
	if err := pgxscan.Select(context.Background(), db, &r1,
		`select * from reports where id=$1`, r.ID,
	); err != nil {
		r.Logger.Warnw("Failed to get report from DB", zap.Error(err))
		return r, err
	} else if len(r1) == 0 {
		return r, types.ErrorReportNotFound
	} else if len(r1) > 1 { // This shouldn't happen
		r.Logger.Errorw("Multiple reports found for parameters", "reports", r1)
		return r, types.ErrorReportNotSpecified
	}

	var r2 ReportEvents
	if err := pgxscan.Select(context.Background(), db, &r2,
		`select * from report_events where report_id=$1`, r.ID,
	); err != nil {
		r.Logger.Warnw("Failed to get report_events from DB", zap.Error(err))
		return r, err
	}

	_ = db.Commit(context.Background()) // Finish tx

	for n, i := range r2 {
		if i.Type == ReportEventDrug && i.DrugID != uuid.Nil {
			if drug, err := (&Drug{Context: r.Context, Unique: Unique{ID: i.DrugID}}).Get(); err != nil {
				return r, err
			} else {
				r2[n].Drug = *drug
			}
		}
	}

	sources, err := (&ReportSource{Context: r.Context, Report: r.ID}).Get()
	if err != nil {
		return r, err
	}

	subject, err := (&ReportSubject{Context: r.Context, Report: r.ID}).Get()
	if err != nil {
		return r, err
	}

	user, err := (&User{Context: r.Context, Unique: Unique{ID: r1[0].Account}}).Get()
	if err != nil {
		r.Logger.Infow("Failed to get user", "report", r1[0].Account, "user", user)
		return r, err
	}

	r2.Sort()
	r.FromData(r1[0])
	r.Sources = sources
	r.Subject = subject
	r.Events = r2
	r.User = *user.CopyPublic()

	return r, nil
}

func (r *Report) Post() (*Report, error) {
	r.InitType(r)
	db := r.DB()
	defer db.Commit(context.Background())

	// Init report UUID
	if r.NilUUID() {
		if err := r.InitUUID(r.Logger); err != nil {
			return r, err
		}

		// We don't need to fix r.Events IDs because we just use r.ID when inserting
	}

	// Insert report
	if _, err := db.Exec(context.Background(),
		`insert into reports(
			id,
			account_id,
			creation_time,
			modified_time,
			report_date,
			title,
			setting
		) values($1, $2, $3, $4, $5, $6, $7);`,
		r.ID, r.Account, r.Created.String(), r.LastModified.String(), r.Date.String(), r.Title, r.Setting,
	); err != nil {
		r.Logger.Warnw("Failed to write report to DB", zap.Error(err))
		_ = db.Rollback(context.Background())
		return r, err
	}

	// Insert report sources
	for _, s := range r.Sources {
		if _, err := s.Post(db); err != nil {
			r.Logger.Warnw("Failed to write report to DB", zap.Error(err))
			_ = db.Rollback(context.Background())
			return r, err
		}
	}

	// Insert report subject
	if r.Subject != nil {
		if _, err := r.Subject.Post(db); err != nil {
			r.Logger.Warnw("Failed to write report to DB", zap.Error(err))
			_ = db.Rollback(context.Background())
			return r, err
		}
	}

	// Insert report drugs
	for _, e := range r.Events {
		if e.Type == ReportEventDrug {
			if _, err := e.Drug.Post(db); err != nil {
				r.Logger.Warnw("Failed to write report to DB", zap.Error(err))
				_ = db.Rollback(context.Background())
				return r, err
			}
		}
	}

	// Finally, insert report events
	for _, e := range r.Events {
		// Create our query dynamically. This really only exists to append `drug_uuid` when needed, a better solution would be nicer.
		insertFields := make([]interface{}, 0)

		insertQuery := []string{"report_id", "event_index", "event_timestamp", "event_type", "event_section", "event_content"}
		insertFields = append(insertFields, r.ID, e.Index, e.Timestamp.String(), e.Type, e.Section, e.Content)

		if e.Drug.ID != uuid.Nil {
			insertQuery = append(insertQuery, "event_drug")
			insertFields = append(insertFields, e.Drug.ID)
		}

		insertValues := make([]string, 0)
		for n, _ := range insertQuery {
			insertValues = append(insertValues, fmt.Sprintf("$%v", n+1))
		}

		query := fmt.Sprintf(
			`insert into report_events(%s) values(%s);`,
			strings.Join(insertQuery, ", "), strings.Join(insertValues, ", "),
		)

		if _, err := db.Exec(context.Background(),
			query,
			insertFields...,
		); err != nil {
			r.Logger.Warnw("Failed to write report to DB", zap.Error(err))
			_ = db.Rollback(context.Background())
			return r, err
		}
	}

	return r, nil
}

func (r *Report) FromBody(r1 *http.Request) (*Report, error) {
	r.InitType(r)

	type ReportFormMedication struct {
		DrugName   string `json:"drug_name,omitempty"`
		DrugDosage string `json:"drug_dosage,omitempty"`
		RoA        int64  `json:"roa,string,omitempty"`
		Prescribed int64  `json:"prescribed,string,omitempty"`
	}

	type ReportFormSubject struct {
		Age           int64                           `json:"subject_age,string,omitempty"`
		Gender        []string                        `json:"subject_gender,omitempty"`
		ImperialUnits bool                            `json:"use_imperial,omitempty"`
		HeightCm      Decimal                         `json:"subject_height_cm,omitempty"`
		HeightFt      Decimal                         `json:"subject_height_ft,omitempty"`
		HeightIn      Decimal                         `json:"subject_height_in,omitempty"`
		WeightKg      Decimal                         `json:"subject_weight_kg,omitempty"`
		WeightLbs     Decimal                         `json:"subject_weight_lbs,omitempty"`
		Medications   map[string]ReportFormMedication `json:"medications,omitempty"`
	}

	type ReportFormEvent struct {
		Timestamp  string `json:"timestamp,omitempty"`
		IsDrug     bool   `json:"is_drug,omitempty"`
		Section    int64  `json:"section,string,omitempty"`
		Content    string `json:"content,omitempty"`
		DrugName   string `json:"drug_name,omitempty"`
		DrugDosage string `json:"drug_dosage,omitempty"`
		RoA        int64  `json:"roa,string,omitempty"`
		Prescribed int64  `json:"prescribed,string,omitempty"`
	}

	type ReportForm struct {
		Title      string            `json:"title"`
		Setting    string            `json:"setting,omitempty"`
		ReportDate string            `json:"report_date"`
		Subject    ReportFormSubject `json:"report_subject,omitempty"`
		Events     []ReportFormEvent `json:"report_sections,omitempty"`
	}

	// We need a report ID in order to parse the report sections.
	// Anything calling this method should init an ID manually, unless making a new report.
	if r.NilUUID() {
		err := r.InitUUID(r.Logger)
		if err != nil {
			return r, err
		}
	}

	if r1.Body == nil {
		return r, types.ErrorStringEmpty.PrefixedError("Request body")
	}

	defer r1.Body.Close()
	body, err := io.ReadAll(r1.Body)
	if err != nil {
		return r, err
	}

	if len(body) == 0 {
		return r, types.ErrorStringEmpty.PrefixedError("Request body")
	}

	var rf *ReportForm
	err = json.Unmarshal(body, &rf)
	if err != nil {
		return r, err
	}

	//
	// Now we should have all the data, we need to turn some types into Go types to make sense.

	// First lets ensure we have a title in the report form, as they are required
	if len(rf.Title) == 0 {
		return r, types.ErrorStringEmpty.PrefixedError("Report title")
	}

	// Next, lets fix the create and last modified timestamps, if they're not valid
	if !r.Created.Valid() {
		r.Created.Now()
	}

	if !r.LastModified.Valid() {
		r.LastModified = r.Created
	}

	// TODO: Sources

	//
	// Parse report subject info

	// Parse gender
	gender := ""
	if len(rf.Subject.Gender) > 0 {
		caser := cases.Title(language.English)
		gender = caser.String(rf.Subject.Gender[0])
	}

	// Parse unit, height and weight
	displayUnit := UnitMetric
	if rf.Subject.ImperialUnits {
		displayUnit = UnitImperial
	}

	height := &Decimal{}
	weight := &Decimal{}

	if displayUnit == UnitMetric {
		height = &rf.Subject.HeightCm
		weight = &rf.Subject.WeightKg
	} else {
		height.Zero()
		height.Add(rf.Subject.HeightFt.Mul(30.48)) // Convert Ft to Cm and add
		height.Add(rf.Subject.HeightIn.Div(2.54))  // Convert In to Cm and add

		weight.Zero()
		weight.Add(rf.Subject.WeightLbs.Div(2.205)) // Convert Lbs to Kg and add
	}

	// Create report subject
	r.Subject = &ReportSubject{
		Context:     r.Context,
		Report:      r.Unique.ID,
		Age:         rf.Subject.Age,
		Gender:      gender,
		DisplayUnit: displayUnit,
		HeightCm:    *height,
		WeightKg:    *weight,
		Medications: make([]Drug, 0),
	}

	// Parse medications
	for _, medication := range rf.Subject.Medications {
		if len(medication.DrugName) == 0 && len(medication.DrugDosage) == 0 && medication.RoA == 0 && medication.Prescribed == 0 {
			// Keep only non-empty medications
			continue
		}

		drug := &Drug{ // TODO: Frequency is not parsed here
			Account:    r.Account,
			Name:       medication.DrugName,
			RoA:        RouteOfAdministration(medication.RoA),
			Prescribed: DrugPrescribed(medication.Prescribed),
		}
		drug.ParseDose(medication.DrugDosage)

		err = drug.Unique.InitUUID(r.Logger)
		if err != nil {
			return r, err
		}

		r.Subject.Medications = append(r.Subject.Medications, *drug)
	}

	// Now we want to trim empty sections from the array
	formEvents := rf.Events[:0]
	for _, s := range rf.Events {
		sectionEmpty := false

		if s.IsDrug {
			if s.Section == 0 && len(s.DrugName) == 0 && len(s.DrugDosage) == 0 && s.RoA == 0 && s.Prescribed == 0 {
				sectionEmpty = true
			}
		} else {
			if s.Section == 0 && len(s.Content) == 0 {
				sectionEmpty = true
			}
		}

		// Keep non-empty sections
		if !sectionEmpty {
			formEvents = append(formEvents, s)
		}
	}

	// Only keep non-empty sections
	rf.Events = formEvents

	// Try to find if any of the timestamps have been set
	firstTimestamp := "T00:00:00Z"
	for _, s := range rf.Events {
		if len(s.Timestamp) > 0 {
			firstTimestamp = "T" + s.Timestamp + ":00Z"
			break
		}
	}

	// Now lets parse the rf.ReportDate as an actual Timestamp, if we have one
	if len(rf.ReportDate) > 0 {
		date, err := r.Date.Parse(rf.ReportDate + firstTimestamp)
		if err != nil {
			return r, err
		}

		r.Date = *date
	}

	// Now we can parse the sections properly
	sections := make(ReportEvents, 0)

	for n, s := range rf.Events {
		// First we parse each event's timestamp to add to the event
		// If the user didn't set a date, these each default to 0001-01-01, as r.Date is not set
		timestamp, err := r.Date.ParseTime("T" + s.Timestamp + ":00Z")
		if err != nil && len(s.Timestamp) > 0 {
			return r, err
		}

		// Create default event
		event := &ReportEvent{
			Report:  r.Unique.ID,
			Index:   int64(n),
			Type:    ReportEventNote,
			Section: ReportEventSection(s.Section),
			Content: s.Content,
		}

		// If we parsed a timestamp, add it
		if timestamp != nil {
			event.Timestamp = *timestamp
		}

		// Add drug info if it's a drug
		if s.IsDrug {
			event.Type = ReportEventDrug
			event.Drug = Drug{ // TODO: Frequency is not parsed here
				Account:    r.Account,
				Name:       s.DrugName,
				RoA:        RouteOfAdministration(s.RoA),
				Prescribed: DrugPrescribed(s.Prescribed),
			}
			event.Drug.ParseDose(s.DrugDosage)

			err = event.Drug.Unique.InitUUID(r.Logger)
			if err != nil {
				return r, err
			}
		}

		// Add this to the parsed sections type
		sections = append(sections, event)
	}

	// Set title, sections and events now that we have the sections parsed
	r.Title = rf.Title
	r.Setting = rf.Setting
	r.Events = sections

	// We should have a completely parsed Report now
	return r, nil
}

func (r *Report) FromData(r1 *Report) {
	r.InitType(r)
	r.ID = r1.ID
	r.Account = r1.Account
	r.Created = r1.Created
	r.LastModified = r1.LastModified
	r.Date = r1.Date
	r.Title = r1.Title
	r.Setting = r1.Setting
	r.Events = r1.Events
}
