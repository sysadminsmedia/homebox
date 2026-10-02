package services

import (
	"errors"
	"fmt"

	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/mailer"
)

// FoundService relays messages from anonymous finders of lost items to the
// item's group owner. It never exposes the owner's address to the caller.
// It does not look up FoundContact itself; the caller (the found-item HTTP
// handler) resolves the contact via repo.AllRepos.Groups.FoundContactBy* and
// passes the result in.
type FoundService struct {
	mailer *mailer.Mailer
}

// MailerReady reports whether the SMTP mailer is configured and usable.
func (svc *FoundService) MailerReady() bool {
	return svc.mailer != nil && svc.mailer.Ready()
}

// SendContact relays a finder's message to the item owner's email address.
func (svc *FoundService) SendContact(contact repo.FoundContact, message, replyTo string) error {
	if !svc.MailerReady() {
		return errors.New("mailer is not configured")
	}

	subject, body := buildFoundContactEmail(contact.ItemName, message, replyTo)

	msg := mailer.NewMessageBuilder().
		SetTo(contact.OwnerName, contact.OwnerEmail).
		SetFrom("Homebox", svc.mailer.From).
		SetSubject(subject).
		SetBody(body).
		Build()

	return svc.mailer.Send(msg)
}

// buildFoundContactEmail builds the subject and body for a found-item contact
// email. mailer.Mailer.Send always sends with Content-Type: text/html (see
// pkgs/mailer/mailer.go), so even though this content is conceptually plain
// text, every interpolated value (finder message, reply address, and item
// name) must be HTML-escaped before interpolation - otherwise an anonymous
// finder, or a group member who names an item with HTML in it, could inject
// markup (or script, depending on the recipient's mail client) into an email
// sent to the item owner. Wrapped in <pre> so newlines in the finder's
// message still render as line breaks. The subject line is not escaped this
// way; it goes through mime.QEncoding in the mailer, which already protects
// email headers.
func buildFoundContactEmail(itemName, message, replyTo string) (subject, body string) {
	subject = fmt.Sprintf("Someone found your item: %s", itemName)

	body = fmt.Sprintf(
		`<pre style="white-space: pre-wrap; font-family: inherit;">Someone scanned the label on your item "%s" and sent you a message through Homebox:

%s
</pre>`,
		htmlEscape(itemName), htmlEscape(message),
	)
	if replyTo != "" {
		body += fmt.Sprintf(`<p>Reply to: %s</p>`, htmlEscape(replyTo))
	}
	return subject, body
}
