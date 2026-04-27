package models

import (
	"database/sql"
	"log"
)

// Guest represents a guest in the database.
type Guest struct {
	GuestID      int    `json:"guestId"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	RsvpStatus   string `json:"rsvpStatus"`
	EmailAddress string `json:"emailAddress"`
	PhoneNumber  string `json:"phoneNumber"`
	SongRequest  string `json:"songRequest"`
	IsPlusOne    bool   `json:"isPlusOne"`
}

type Family struct {
	FamilyID      int     `json:"familyId"`
	FamilyName    string  `json:"familyName"`
	IsTraveling   bool    `json:"isTraveling"`
	MailingAddr   string  `json:"mailingAddr"`
	Invite        string  `json:"invitationCode"`
	HasResponded  bool    `json:"hasResponded"`
	FamilyMembers []Guest `json:"familyMembers"`
}

type NewFamilyPayload struct {
	Families []Family `json:"families"`
}

type RsvpResponse struct {
	FamilyID  int        `json:"familyId"`
	Responses []Response `json:"responses"`
}

type Response struct {
	GuestID          int    `json:"guestId"`
	Answer           int    `json:"answer"`
	SongRequest      string `json:"songRequest"`
	EmailAddress     string `json:"emailAddress"`
	IsPlusOne        bool   `json:"isPlusOne"`
	PlusOneFirstName string `json:"plusOneFirstName"`
	PlusOneLastName  string `json:"plusOneLastName"`
}

// GuestModel holds database connection.
type GuestModel struct {
	DB *sql.DB
}

// NewGuestModel creates a new GuestModel.
func NewGuestModel(db *sql.DB) *GuestModel {
	return &GuestModel{DB: db}
}

// GetFamilyByInvitationCode fetches a guest by invitation code.
func (m *GuestModel) GetFamilyByInvitationCode(invitationCode string) (Family, error) {
	log.Printf("Getting Family by invitation code %s", invitationCode)
	query := `
SELECT families.family_id,
       families.family_name,
       families.mailing_address,
       families.is_traveling,
       invitationCodes.invitation_code,
       families.has_responded,
       guestsInfo.guest_id,
       guestsInfo.first_name,
       guestsInfo.last_name,
       guestsInfo.email_addr,
       guestsInfo.phone_number,
       guestsInfo.song_request,
       guestsInfo.is_plus_one,
       rsvpCodes.status_name
FROM families
	join invitationCodes
    	on families.family_id = invitationCodes.family_id
	join guestsInfo
		on guestsInfo.family_id = families.family_id
	left join rsvpList
		on rsvpList.guest_id = guestsInfo.guest_id
	left join rsvpCodes
		on rsvpCodes.status_id = rsvpList.status_id
WHERE invitationCodes.invitation_code = $1
ORDER BY families.family_id, guestsInfo.guest_id;
`

	var family Family

	rows, err := m.DB.Query(query, invitationCode)
	if err != nil {
		return Family{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var guest Guest
		err := rows.Scan(
			&family.FamilyID,
			&family.FamilyName,
			&family.MailingAddr,
			&family.IsTraveling,
			&family.Invite,
			&family.HasResponded,
			&guest.GuestID,
			&guest.FirstName,
			&guest.LastName,
			&guest.EmailAddress,
			&guest.PhoneNumber,
			&guest.SongRequest,
			&guest.IsPlusOne,
			&guest.RsvpStatus,
		)
		if err != nil {
			return Family{}, err
		}
		family.FamilyMembers = append(family.FamilyMembers, guest)
	}
	if err := rows.Err(); err != nil {
		return Family{}, err
	}

	if family.FamilyID == 0 {
		return Family{}, sql.ErrNoRows
	}
	return family, nil
}

const searchGuestByID = `
SELECT guestsInfo.first_name,
       guestsInfo.last_name
FROM guestsInfo
WHERE guestsInfo.guest_id = $1
`

// GetGuestInfoByGuestID fetches a guest by invitation code.
func (m *GuestModel) GetGuestInfoByGuestID(guestID int) (string, string, error) {
	log.Printf("Getting guest by guestID %d", guestID)

	rows, err := m.DB.Query(searchGuestByID, guestID)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()
	var guest Guest
	for rows.Next() {

		err := rows.Scan(
			&guest.FirstName,
			&guest.LastName,
		)
		if err != nil {
			return "", "", err
		}
	}
	if err := rows.Err(); err != nil {
		return "", "", err
	}

	return guest.FirstName, guest.LastName, nil
}

func (m *GuestModel) InsertNewFamiliesAndGuests(newFamily Family) (string, int, error) {

	log.Printf("Inserting new family: %+v", newFamily)

	tx, err := m.DB.Begin()
	if err != nil {
		return "", 0, err
	}
	defer tx.Rollback()

	// Insert into families
	familyID, err := m.insertNewFamily(tx, newFamily)
	if err != nil {
		return "", 0, err
	}
	log.Printf("Family inserted. ID: %d", familyID)

	// Insert invitation code
	err = m.insertNewInvitationCode(tx, newFamily.Invite, familyID)
	if err != nil {
		return "", 0, err
	}

	// Insert guests
	err = m.insertNewGuests(tx, newFamily, familyID)
	if err != nil {
		return "", 0, err
	}

	// Commit after both families and guestsInfo inserts have completed
	err = tx.Commit()
	if err != nil {
		return "", 0, err
	}

	return newFamily.FamilyName, familyID, nil
}

func (m *GuestModel) insertNewFamily(tx *sql.Tx, newFamily Family) (int, error) {
	log.Println("In insertNewFamily")

	var familyID int

	// Insert and return familyID
	err := tx.QueryRow(`
		INSERT INTO families (family_name, is_traveling, mailing_address, has_responded)
		VALUES ($1,$2,$3,$4)
		RETURNING family_id
	`,
		newFamily.FamilyName,
		newFamily.IsTraveling,
		newFamily.MailingAddr,
		newFamily.HasResponded,
	).Scan(&familyID)
	if err != nil {
		return 0, err
	}
	return familyID, nil
}

const insertInvitationCodes = `
INSERT INTO invitationCodes
(invitation_code, family_id)
values ($1, $2)
`

func (m *GuestModel) insertNewInvitationCode(tx *sql.Tx, invitationCode string, familyID int) error {
	log.Println("In insertNewInvitationCode")
	stmt, err := m.prepareStatements(tx, insertInvitationCodes)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(stmt)

	_, err = stmt.Exec(invitationCode, familyID)
	if err != nil {
		return err
	}
	return nil
}

const insertGuestsInfo = `
		INSERT INTO guestsInfo
		(family_id, first_name, last_name, email_addr, phone_number, song_request, is_plus_one)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING guest_id
	`
const insertRsvpList = `
insert into rsvpList
(guest_id, status_id)
values ($1, $2)
`

func (m *GuestModel) insertNewGuests(tx *sql.Tx, newFamily Family, familyID int) error {
	log.Println("In insertNewGuests")
	guestsInfoStmt, err := m.prepareStatements(tx, insertGuestsInfo)
	if err != nil {
		return err
	}
	rsvpListStmt, err := m.prepareStatements(tx, insertRsvpList)
	if err != nil {
		return err
	}
	defer func(guestsInfoStmt *sql.Stmt, rsvpListStmt *sql.Stmt) {
		err := guestsInfoStmt.Close()
		if err != nil {
			log.Fatal(err)
		}
		err = rsvpListStmt.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(guestsInfoStmt, rsvpListStmt)

	for _, familyMember := range newFamily.FamilyMembers {
		var guestID int64

		err := guestsInfoStmt.QueryRow(
			familyID,
			familyMember.FirstName,
			familyMember.LastName,
			familyMember.EmailAddress,
			familyMember.PhoneNumber,
			familyMember.SongRequest,
			familyMember.IsPlusOne,
		).Scan(&guestID)
		if err != nil {
			return err
		}
		_, err = rsvpListStmt.Exec(
			guestID,
			1,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

const updateRsvpList = `
update rsvpList
set status_id = $1
where guest_id = $2
`
const updatePlusOnesInfo = `
update guestsInfo
set email_addr = $1,
first_name = $2,
last_name = $3,
song_request = $4
where guest_id = $5
`
const updateGuestsInfo = `
update guestsInfo
set email_addr = $1,
song_request = $2
where guest_id = $3
`
const updateFamilyHasResponded = `
update families
set has_responded = true
where family_id = $1
`

func (m *GuestModel) RespondRsvp(RsvpResponses RsvpResponse) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rsvpStmt, err := tx.Prepare(updateRsvpList)
	if err != nil {
		return err
	}
	defer rsvpStmt.Close()

	plusOneStmt, err := tx.Prepare(updatePlusOnesInfo)
	if err != nil {
		return err
	}
	defer plusOneStmt.Close()

	guestStmt, err := tx.Prepare(updateGuestsInfo)
	if err != nil {
		return err
	}
	defer guestStmt.Close()

	familyStmt, err := tx.Prepare(updateFamilyHasResponded)
	if err != nil {
		return err
	}
	defer familyStmt.Close()

	for _, guestResponse := range RsvpResponses.Responses {
		_, err := rsvpStmt.Exec(guestResponse.Answer, guestResponse.GuestID)
		if err != nil {
			log.Println("Issue in rsvpStmt")
			return err
		}

		if guestResponse.IsPlusOne {
			_, err = plusOneStmt.Exec(
				guestResponse.EmailAddress,
				guestResponse.PlusOneFirstName,
				guestResponse.PlusOneLastName,
				guestResponse.SongRequest,
				guestResponse.GuestID,
			)
			if err != nil {
				log.Println("Issue in plusOneStmt")
				return err
			}
		} else {
			_, err = guestStmt.Exec(
				guestResponse.EmailAddress,
				guestResponse.SongRequest,
				guestResponse.GuestID,
			)
			if err != nil {
				log.Println("Issue in guestStmt")
				return err
			}
		}
	}
	_, err = familyStmt.Exec(
		RsvpResponses.FamilyID,
	)
	if err != nil {
		log.Println("Issue in rsvpStmt")
		return err
	}

	return tx.Commit()
}

func (m *GuestModel) prepareStatements(tx *sql.Tx, query string) (*sql.Stmt, error) {
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt, err
}
