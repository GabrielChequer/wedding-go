package models

import (
	"database/sql"
	"log"
)

// Guest represents a guest in the database.
type Guest struct {
	GuestID      string `json:"guestId"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	EmailAddress string `json:"emailAddress"`
	PhoneNumber  string `json:"phoneNumber"`
	SongRequest  string `json:"songRequest"`
	IsPlusOne    bool   `json:"isPlusOne"`
}

type Family struct {
	FamilyID      string  `json:"familyId"`
	FamilyName    string  `json:"familyName"`
	IsTraveling   bool    `json:"isTraveling"`
	MailingAddr   string  `json:"mailingAddr"`
	Invite        string  `json:"invitationCode"`
	FamilyMembers []Guest `json:"familyMembers"`
}

type NewFamilyPayload struct {
	Families []Family `json:"families"`
}

type RsvpResponse struct {
	GuestID    string `json:"guestId"`
	RsvpStatus string `json:"rsvpStatus"`
}

type NewRsvpResponse struct {
	Responses []RsvpResponse `json:"responses"`
}

// RSVP values
const pending = "pending"
const pendingValue = 1
const declined = "declined"
const declinedValue = 2
const accepted = "accepted"
const acceptedValue = 3

// GuestModel holds database connection.
type GuestModel struct {
	DB *sql.DB
}

// NewGuestModel creates a new GuestModel.
func NewGuestModel(db *sql.DB) *GuestModel {
	return &GuestModel{DB: db}
}

// GetGuestByInvitationCode fetches a guest by invitation code.
func (m *GuestModel) GetFamilyByInvitationCode(invitationCode string) (Family, error) {
	log.Printf("Getting Family by invitation code %s", invitationCode)
	query := `
SELECT families.family_id,
       families.family_name,
       families.mailing_address,
       families.is_traveling,
       invitationCodes.invitation_code,
       guestsInfo.guest_id,
       guestsInfo.first_name,
       guestsInfo.last_name,
       guestsInfo.email_addr,
       guestsInfo.phone_number,
       guestsInfo.song_request,
       guestsInfo.is_plus_one
FROM families
	join invitationCodes
    	on families.family_id = invitationCodes.family_id
	join guestsInfo
		on guestsInfo.family_id = families.family_id
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
			&guest.GuestID,
			&guest.FirstName,
			&guest.LastName,
			&guest.EmailAddress,
			&guest.PhoneNumber,
			&guest.SongRequest,
			&guest.IsPlusOne,
		)
		if err != nil {
			return Family{}, err
		}
		family.FamilyMembers = append(family.FamilyMembers, guest)
	}
	if err := rows.Err(); err != nil {
		return Family{}, err
	}

	if family.FamilyID == "0" {
		return Family{}, sql.ErrNoRows
	}
	return family, nil
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
		INSERT INTO families (family_name, is_traveling, mailing_address)
		VALUES ($1,$2,$3)
		RETURNING family_id
	`,
		newFamily.FamilyName,
		newFamily.IsTraveling,
		newFamily.MailingAddr,
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
	`

func (m *GuestModel) insertNewGuests(tx *sql.Tx, newFamily Family, familyID int) error {
	log.Println("In insertNewGuests")
	stmt, err := m.prepareStatements(tx, insertGuestsInfo)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(stmt)

	for _, familyMember := range newFamily.FamilyMembers {
		_, err := stmt.Exec(
			familyID,
			familyMember.FirstName,
			familyMember.LastName,
			familyMember.EmailAddress,
			familyMember.PhoneNumber,
			familyMember.SongRequest,
			familyMember.IsPlusOne,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

const respondRsvp = `
insert into rsvpList
(guest_id, status_id)
values ($1, $2)
`

func (m *GuestModel) RespondRsvp(RsvpResponses []RsvpResponse) error {
	log.Println("Responding RSVP")

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := m.prepareStatements(tx, respondRsvp)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(stmt)

	for _, guestResponse := range RsvpResponses {
		log.Printf("response %v", guestResponse.GuestID)
		switch guestResponse.RsvpStatus {
		case accepted:
			_, err := stmt.Exec(guestResponse.GuestID, acceptedValue)
			if err != nil {
				return err
			}
		case declined:
			_, err := stmt.Exec(guestResponse.GuestID, declinedValue)
			if err != nil {
				return err
			}
		default:
			log.Println("Option not supported")
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (m *GuestModel) prepareStatements(tx *sql.Tx, query string) (*sql.Stmt, error) {
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt, err
}
