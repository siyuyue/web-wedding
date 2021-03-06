package backend

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/sendgrid/sendgrid-go.v2"
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
	ConfirmationSent bool
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
	SendGridAPIKey string
}

type RSVPCode struct {
	RemainingQuota int
}

func init() {
	http.HandleFunc("/rsvp", handler)
	http.HandleFunc("/admin/send_email", sendEmailHandler)
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

func emailWithSendGrid(sg *sendgrid.SGClient, from string, to string, subject string, body string) error {
	message := sendgrid.NewMail()
	message.AddTo(to)
	message.AddBcc(from)
	message.SetFrom(from)
	message.SetSubject(subject)
	message.SetText(body)
	return sg.Send(message)
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
	sg := sendgrid.NewSendGridClientWithApiKey(rsvpRegistry.SendGridAPIKey)
	sg.Client = urlfetch.Client(ctx)
	if err := emailWithSendGrid(sg, rsvpRegistry.Email, g.Email, "Thank you for your RSVP to Di and Siyu's Wedding!", "We look forward to seeing you on August 19, 2017.\n\nFor more information please see http://www.diandsiyu.wedding\n\nDi & Siyu"); err != nil {
		// Already succeeded, no need to show failure message.
		ctx.Errorf("Failed to send email: %s\n", err)
	}
	respond(w, true, "You've successfully rsvp-ed!")
}

func sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	keyRSVP := datastore.NewKey(ctx, "AppRegistry", "EnableRSVP", 0, nil)
	var rsvpRegistry EnableRSVP
	if e := datastore.Get(ctx, keyRSVP, &rsvpRegistry); e != nil {
		ctx.Errorf("Failed to get EnableRSVP: %s\n", e)
		respond(w, false, "Oops, something went wrong.")
		return
	}
	
	guests := make([]Guest, 0, 150)
	guestQuery := datastore.NewQuery("Guest")
	for t := guestQuery.Run(ctx);; {
		var guest Guest
		_, err := t.Next(&guest)
		if err == datastore.Done {
            break
        }
        if err != nil {
            fmt.Fprintf(w, "Datastore read failed: %s\n", err)
            return
        }
		
		guests = append(guests, guest)
	}
	
	q := datastore.NewQuery("RSVP")
	for t := q.Run(ctx);; {
		var rsvp GuestRSVP
		key, err := t.Next(&rsvp)
		if err == datastore.Done {
            break
        }
        if err != nil {
            fmt.Fprintf(w, "Datastore read failed: %s\n", err)
            return
        }
		if rsvp.ConfirmationSent {
			continue
		}
		
		message := "This email is to confirm your following RSVP at Di & Siyu's wedding on August 19, 2017.\n\n\n\n"
		
		fmt.Fprintf(w, "For guest %s %s (%s):\n", rsvp.FirstName, rsvp.LastName, rsvp.Email)
		
		for _, guest := range guests {
			if (guest.FirstName == rsvp.FirstName && guest.LastName == rsvp.LastName) || (guest.GuestOf == rsvp.FirstName + " " +  rsvp.LastName) {
				var guestType string
				if guest.IsChildGuest {
					guestType = "Child"
				} else {
					guestType = "Adult"
				}
				message += fmt.Sprintf("%s guest: %s %s, meal option: %s\n\n", guestType, guest.FirstName, guest.LastName, guest.MealOption)
			}
		}
		message += "\nPlease reply if you need to change or cancel the RSVP. \n\n\n\nBest,\n\nDi & Siyu\n\n"
		fmt.Fprintf(w, "%s\n", message)
		// Email guest
		ctx.Infof("Sending confirmation email to %s", rsvp.Email)
		sg := sendgrid.NewSendGridClientWithApiKey(rsvpRegistry.SendGridAPIKey)
		sg.Client = urlfetch.Client(ctx)
		if err := emailWithSendGrid(sg, rsvpRegistry.Email, rsvp.Email, "RSVP Confirmation", message); err != nil {
			ctx.Infof("Failed to send confirmation email to %s", rsvp.Email)
			fmt.Fprintf(w, "Failed to send email to %s: %s\n", rsvp.Email, err)
		} else {
			ctx.Infof("Succeeded sending confirmation email to %s", rsvp.Email)
			rsvp.ConfirmationSent = true
			if _, e := datastore.Put(ctx, key, &rsvp); e != nil {
				fmt.Fprintf(w, "Failed to write back to RSVP: %s\n", e)
				return
			}
		}
	}
	fmt.Fprintf(w, "success!")
}
