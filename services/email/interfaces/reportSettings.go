package interfaces

import (
	"database/sql"
	"inv_app/services/reports"
	"sync"
)

type CustomerSet map[CustomerID]struct{}

type Email string
type CustomerID int
type RepresentativeEmail Email
type ReportRow reports.CustomerUsageRep
type Report []ReportRow
type CustomerEmails []Email
type Status string

type CustomerSettings struct {
	sync.Mutex
	Report
	CustomerEmails
	RepresentativeEmail
}

type ReportSettings struct {
	sync.RWMutex
	CustomerSettings map[CustomerID]*CustomerSettings
}

// NewReport initializes a new ReportSettings with a non-nil map.
func NewReport() *ReportSettings {
	return &ReportSettings{
		CustomerSettings: make(map[CustomerID]*CustomerSettings),
	}
}

// AddReport safely adds a report row for a customer.
func (r *ReportSettings) AddReport(customerId CustomerID, report reports.CustomerUsageRep) {
	r.Lock()
	defer r.Unlock()

	cs, ok := r.CustomerSettings[customerId]
	if !ok {
		cs = &CustomerSettings{}
		r.CustomerSettings[customerId] = cs
	}

	cs.Lock()
	defer cs.Unlock()
	cs.Report = append(cs.Report, ReportRow(report))
}

// SetCustomerEmails fetches and sets customer emails from the DB.
func (r *ReportSettings) SetCustomerEmails(db *sql.DB, customerId CustomerID) error {
	r.Lock()
	defer r.Unlock()

	rows, err := db.Query(`
		SELECT
			c.customer_id, ce.email
		FROM
			customers c
		RIGHT JOIN customer_emails ce ON ce.customer_id = c.customer_id
		WHERE
			c.is_connected_to_reports = true
			AND ($1 = 0 OR c.customer_id = $1);
	`, customerId)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid CustomerID
		var email Email
		if err := rows.Scan(&cid, &email); err != nil {
			return err
		}

		cs, ok := r.CustomerSettings[cid]
		if !ok {
			cs = &CustomerSettings{}
			r.CustomerSettings[cid] = cs
		}

		cs.Lock()
		cs.CustomerEmails = append(cs.CustomerEmails, email)
		cs.Unlock()
	}

	return rows.Err()
}

// SetRepresentativeEmail fetches and sets the representative email from the DB.
func (r *ReportSettings) SetRepresentativeEmail(db *sql.DB, customerId CustomerID) error {
	r.Lock()
	defer r.Unlock()

	rows, err := db.Query(`
		SELECT
			c.customer_id, u.email
		FROM
			customers c
		RIGHT JOIN users u ON u.user_id = c.user_id
		WHERE
			c.is_connected_to_reports = true
			AND ($1 = 0 OR c.customer_id = $1);
	`, customerId)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid CustomerID
		var email Email
		if err := rows.Scan(&cid, &email); err != nil {
			return err
		}

		cs, ok := r.CustomerSettings[cid]
		if !ok {
			cs = &CustomerSettings{}
			r.CustomerSettings[cid] = cs
		}

		cs.Lock()
		cs.RepresentativeEmail = RepresentativeEmail(email)
		cs.Unlock()
	}

	return rows.Err()
}

func (r *ReportSettings) GetCustomerSettings(customerId CustomerID) (*CustomerSettings, bool) {
	r.RLock()
	defer r.RUnlock()

	cs, ok := r.CustomerSettings[customerId]
	return cs, ok
}

// IsCustomerSettingsEmpty returns true if there are no customer settings.
func (r *ReportSettings) IsCustomerSettingsEmpty() bool {
	r.RLock()
	defer r.RUnlock()

	return len(r.CustomerSettings) == 0
}

// IsCustomerEmailsEmpty returns true if the specified customer has no emails.
func (r *ReportSettings) IsCustomerEmailsEmpty(customerId CustomerID) bool {
	cs, ok := r.GetCustomerSettings(customerId)
	if !ok {
		return true
	}

	cs.Lock()
	defer cs.Unlock()
	return len(cs.CustomerEmails) == 0
}

// IsRepresentativeEmailEmpty returns true if the specified customer has no representative email.
func (r *ReportSettings) IsRepresentativeEmailEmpty(customerId CustomerID) bool {
	r.RLock()
	defer r.RUnlock()

	cs, ok := r.CustomerSettings[customerId]
	if !ok {
		return true
	}

	cs.Lock()
	defer cs.Unlock()

	return cs.RepresentativeEmail == ""
}
