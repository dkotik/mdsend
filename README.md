# Mdsend

Project status: active development towads beta.

> Send markdown files as MIME-encoded electronic mail.

## Installation

- Go:
  ```sh
  go install github.com/dkotik/mdsend@latest
  ```
- MacOS:
  ```sh
  brew tap dkotik/tap
  brew install mdsend
  ```

## Example

```markdown
---
subject: "Test Email"
from: "Test Account <test@gmail.com>"
to: "Test Account <test@gmail.com>"
to: "./localfile.yaml|toml|json|csv|txt"
---

# Title

Message body.
```

EmailSend(file.md, templating engine, provider engine)

## Delivery Locking

## Distributors

- [x] Mailgun
- [ ] Resend
- [ ] Loops
- [ ] https://purelymail.com/

## Development Roadmap

- [ ] Anticipate circular imports for recipient lists.
- [ ] Bump `github.com/pelletier/go-toml` to v2.
- [ ] Text-part writer should minify html! minified html will have <img src=url> without quotes! (change inline detector)
- [ ] Should write a better markdown renderer that recognizes single-line youtube and image paragraphs, centers them.
- [ ] Email validation: <https://github.com/reacherhq/check-if-email-exists>.
- [ ] https://github.com/charmbracelet/glamour
- [ ] Beautify the default template. Add `dark.html` one.
- [ ] --log param
* Ensure that "From" is set up correctly with just one email address - MG queues it with 2 and never delivers! MG sux, write AWS SNS queue driver?
* // TODO: look into https://github.com/modoboa/modoboa server.
* Event invitations markup: https://developers.google.com/gmail/markup/reference/event-reservation#basic_event_reminder_without_a_ticket
* alternative API: https://docs.sendwithses.com/ - add it as a driver to "providers"
- support Cuelang! https://github.com/cuelang/cue
- add to https://github.com/rothgar/awesome-tuis and bubbletea list of apps
- https://mailtrap.io/
- SendGrid
- Amazon SES
- Postmark
- Sparkpost
- Brevo
- SendGrid
- Mailchimp
- Resend
- HubSpot
- Loops
- https://emaillabs.io/en/product/

```
// multipart/mixed
// |- multipart/alternative
// | |- text/plain
// | `- multipart/related // | |- text/html // |`- image/png
// `- attachments..
```

## Tools

* OBEY THE LAW. The CAN-SPAM act became law on Jan. 1, 2004. It says there many things you must do as a commercial email-er. Highlights are basically don't be deceptive, and that you MUST include a physical mailing address as well as a working unsubscribe link.
* unsubscribe button: https://blog.leavemealone.app/how-does-the-gmail-unsubscribe-button-work/

Standard U.S. (#10 envelope 4 1/8in. by 9 1/2in.) or European standard C4 (229mm by 324mm) template, depending on your envelope printer setting. There is also DL envelope 110mm x 220mm.

- <https://www.caniemail.com/> - check what is template-appropriate
- <https://github.com/AfterShip/email-verifier>
- <https://www.htmlemailcheck.com/check/>
- Use <https://www.mail-tester.com/> to check the deliverability of your mail.

## Similar Projects

- <https://github.com/domodwyer/mailyak>
- <https://github.com/charmbracelet/pop>
- <https://github.com/wneessen/go-mail>
