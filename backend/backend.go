package backend

import (
	"appengine"
	"appengine/datastore"
	"appengine/mail"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Guest struct {
	FirstName    string
	LastName     string
	IsChildGuest bool
	MealOption   string
	GuestOf      string
}

type GuestRSVP struct {
	FirstName       string
	LastName        string
	Email           string
	StayingAtHotel  bool
	AdultGuestCount int
	ChildGuestCount int
	Extra           string
	Guests          []Guest `datastore:"-"`
}

func (g *GuestRSVP) parse(formData url.Values) error {
	g.FirstName = formData.Get("firstName")
	g.LastName = formData.Get("lastName")
	g.Email = formData.Get("email")
	g.StayingAtHotel, _ = strconv.ParseBool(formData.Get("hotel"))
	g.AdultGuestCount, _ = strconv.Atoi(formData.Get("adultCount"))
	g.ChildGuestCount, _ = strconv.Atoi(formData.Get("childCount"))
	mealOption := formData.Get("entree")
	if e := g.validate(); e != nil {
		return e
	}
	// Self is also a guest.
	g.Guests = make([]Guest, 1+g.AdultGuestCount+g.ChildGuestCount)
	g.Guests[0] = Guest{FirstName: g.FirstName, LastName: g.LastName, IsChildGuest: false, MealOption: mealOption}
	guestOf := g.FirstName + " " + g.LastName
	for i := 0; i < g.AdultGuestCount; i++ {
		g.Guests[i+1] = Guest{
			FirstName:    formData.Get("guestAdultFirstName" + strconv.Itoa(i)),
			LastName:     formData.Get("guestAdultLastName" + strconv.Itoa(i)),
			IsChildGuest: false,
			MealOption:   formData.Get("guestAdultEntree" + strconv.Itoa(i)),
			GuestOf:      guestOf}
	}
	for i := 0; i < g.ChildGuestCount; i++ {
		g.Guests[i+1+g.AdultGuestCount] = Guest{
			FirstName:    formData.Get("guestChildFirstName" + strconv.Itoa(i)),
			LastName:     formData.Get("guestChildLastName" + strconv.Itoa(i)),
			IsChildGuest: true,
			MealOption:   formData.Get("guestChildEntree" + strconv.Itoa(i)),
			GuestOf:      guestOf}
	}
	for _, guest := range g.Guests {
		if e := guest.validate(); e != nil {
			return e
		}
	}
	return nil
}

func (g *Guest) validate() error {
	if g.FirstName == "" {
		return errors.New("Missing First Name")
	}
	return nil
}

func (g *GuestRSVP) validate() error {
	if g.FirstName == "" {
		return errors.New("Missing First Name")
	}
	if g.LastName == "" {
		return errors.New("Missing Last Name")
	}
	if g.Email == "" {
		return errors.New("Missing Email")
	}
	if g.AdultGuestCount < 0 || g.AdultGuestCount > 1 {
		return errors.New("Invalid Adult Guest Count")
	}
	if g.ChildGuestCount < 0 || g.ChildGuestCount > 2 {
		return errors.New("Invalid Child Guest Count")
	}
	return nil
}

type EnableRSVP struct {
	Enable bool
	Email  string
	IsProd bool
}

type RSVPCode struct {
	RemainingQuota int
}

func init() {
	http.HandleFunc("/rsvp", handler)
}

type Response struct {
    Success bool   `json:"success"`
	Message string `json:"message"`
}

func respond(w http.ResponseWriter, success bool, message string) {
    resp := &Response{
	    Success: success,
		Message: message}
	jsonString, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(jsonString))
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Load value from registry.
	ctx := appengine.NewContext(r)
	keyRSVP := datastore.NewKey(ctx, "AppRegistry", "EnableRSVP", 0, nil)
	var rsvpRegistry EnableRSVP
	if e := datastore.Get(ctx, keyRSVP, &rsvpRegistry); e != nil {
		ctx.Errorf("Failed to get EnableRSVP: %s\n", e)
		respond(w, false, "Oops, something went wrong.")
		return
	}
	if !rsvpRegistry.Enable {
		respond(w, false, "Sorry! RSVP not open yet.")
		return
	}

	// Check RSVP code
	r.ParseForm()
	codeString := r.Form.Get("rsvpCode")
	if codeString == "" {
		ctx.Warningf("Request has no RSVP code.")
		respond(w, false, "Missing RSVP code\n")
		return
	}
	keyCode := datastore.NewKey(ctx, "RSVPCode", codeString, 0, nil)
	var code RSVPCode
	if e := datastore.Get(ctx, keyCode, &code); e != nil {
		ctx.Warningf("Request has wrong RSVP code.")
		respond(w, false, "Wrong code, please double check code in your RSVP email.\n")
		return
	}
	if code.RemainingQuota <= 0 {
		ctx.Warningf("Request has expired RSVP code.")
		respond(w, false, "Code has expired.\n")
		return
	}

	rsvpEntityName := "TestRSVP"
	guestEntityName := "TestGuest"
	if rsvpRegistry.IsProd {
		rsvpEntityName = "RSVP"
		guestEntityName = "Guest"
		ctx.Infof("In Prod environment.")
	}

	// Parse input form data.
	g := new(GuestRSVP)
	if e := g.parse(r.Form); e != nil {
		respond(w, false, fmt.Sprintf("%s\n", e))
		return
	}

	// Check if email has already RSVP-ed.
	q := datastore.NewQuery(rsvpEntityName).Filter("Email =", g.Email).Limit(1).KeysOnly()
	if _, e := q.Run(ctx).Next(nil); e == nil {
		respond(w, false, fmt.Sprintf("%s has already RSVP-ed", g.Email))
		return
	}

	// Decrement Quota
	code.RemainingQuota--
	if _, e := datastore.Put(ctx, keyCode, &code); e != nil {
		ctx.Errorf("Failed to write back to RSVPCode: %s\n", e)
		respond(w, false, "Oops, something went wrong.")
		return
	}

	// Save RSVP entry
	key := datastore.NewIncompleteKey(ctx, rsvpEntityName, nil)
	if _, err := datastore.Put(ctx, key, g); err != nil {
		ctx.Errorf("Failed to write RSVPEntry: %s\n", err)
		respond(w, false, "Oops, something went wrong.")
		return
	}

	// Save guest entries
	for _, guest := range g.Guests {
		key := datastore.NewIncompleteKey(ctx, guestEntityName, nil)
		if _, err := datastore.Put(ctx, key, &guest); err != nil {
			ctx.Errorf("Failed to save guest: %s\n", err)
			respond(w, false, "Oops, something went wrong.")
			return
		}
	}

	// Email guest
	msg := &mail.Message{
		Sender:  "Di and Siyu Wedding <" + rsvpRegistry.Email + ">",
		To:      []string{g.Email},
		Subject: "You've Successfully RSVP-ed",
		Body:    "Thank you for RSVP-ing, we look forward to seeing you on our wedding!",
	}
	if err := mail.Send(ctx, msg); err != nil {
		ctx.Errorf("Failed to send email: %s\n", err)
		respond(w, false, "Oops, something went wrong.")
		return
	}
	respond(w, true, "You've successfully rsvp-ed!")
}
