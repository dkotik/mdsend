# Mdsend

Send markdown files as mail.

**Project status: active development towads beta.**

## Features

- **Durable:** mail queues are fault tolerant and atomic, brokered by <Watermill.io> over SQLite3.
- **Flexible:** select the mailer backend based on highest deliverability. Swap it out later without changing anything in your letters or templates.
- **Portable:** runs on many kinds of systems. Will include an embeddable HTTP service and Posgres support in the future.

## Installation

- MacOS:
  ```sh
  brew tap dkotik/tap
  brew install mdsend
  ```
- Build from source:
  ```sh
  go install github.com/dkotik/mdsend@latest
  ```

## Example

```markdown
---
subject: "Test Email"
from: "Test Account <test@gmail.com>"
to: "Test Account <test@gmail.com>"
---

# Title

Message body.
```

## Development Roadmap

- [ ] Anticipate circular imports for recipient lists.
- [ ] Text-part writer should minify html! minified html will have <img src=url> without quotes! (change inline detector)
- [ ] Should write a better markdown renderer that recognizes single-line youtube and image paragraphs, centers them.
- [ ] https://github.com/charmbracelet/glamour
- [ ] Beautify the default template. Add `dark.html` one.
- [ ] Event invitations markup: https://developers.google.com/gmail/markup/reference/event-reservation#basic_event_reminder_without_a_ticket
- [ ] alternative API: https://docs.sendwithses.com/ - add it as a driver to "providers"
- [ ] support Cuelang! https://github.com/cuelang/cue
- [ ] Snailmail support:
  - [ ] Markdown to PDF rendering.
  - [ ] Envelope face PDF generation.
- [x] Mailgun
- [ ] Resend
- [ ] Loops
- [ ] https://purelymail.com/
- [ ] SendGrid
- [ ] Amazon SES
- [ ] Postmark
- [ ] Sparkpost
- [ ] Brevo
- [ ] SendGrid
- [ ] Mailchimp
- [ ] HubSpot
- [ ] Twilio
- [ ] <https://emaillabs.io/en/product/>
- [ ] <https://mailtrap.io/>
- [ ] add to https://github.com/rothgar/awesome-tuis and bubbletea list of apps

## Tools

* OBEY THE LAW. The CAN-SPAM act became law on Jan. 1, 2004. It says there many things you must do as a commercial email-er. Highlights are basically don't be deceptive, and that you MUST include a physical mailing address as well as a working unsubscribe link.
* unsubscribe button: https://blog.leavemealone.app/how-does-the-gmail-unsubscribe-button-work/

- <https://www.caniemail.com/> - check what is template-appropriate
- <https://www.htmlemailcheck.com/check/>
- Use <https://www.mail-tester.com/> to check the deliverability of your mail.
- Address verification:
  - <https://github.com/AfterShip/email-verifier>
  - <https://github.com/reacherhq/check-if-email-exists>
  - <https://hunter.io>
  - <https://verify-email.org>
  - <https://email-checker.net>
  - <https://github.com/mailcheck/mailcheck>
  - <https://github.com/ivolo/disposable-email-domains>
  - <https://github.com/willwhite/freemail>
- Hosting:
  - <https://github.com/modoboa/modoboa>

## Similar Projects

- <https://github.com/domodwyer/mailyak>
- <https://github.com/charmbracelet/pop>
- <https://github.com/wneessen/go-mail>
