## TODO

- [ ] MailYak can help? https://pkg.go.dev/github.com/domodwyer/mailyak?utm_source=godoc#MailYak.MimeBuf
- [ ] implement unsubscribe API using encore functions over etcd for state management
  - [ ] does encore free also provide a database?
- [ ] Use badger for locking https://dgraph.io/docs/badger/get-started/
- [ ] write mailer using encore? https://encore.dev/docs/how-to/secrets
- isolate template selection - seems like a template is re-compiled for each send

* bad yaml crashes the program! - catch the error and recover gracefully

- hook up github.com/leaanthony/spinner

* OBEY THE LAW. The CAN-SPAM act became law on Jan. 1, 2004. It says there many things you must do as a commercial email-er. Highlights are basically don't be deceptive, and that you MUST include a physical mailing address as well as a working unsubscribe link.
* unsubscribe button: https://blog.leavemealone.app/how-does-the-gmail-unsubscribe-button-work/

Standard U.S. (#10 envelope 4 1/8in. by 9 1/2in.) or European standard C4 (229mm by 324mm) template, depending on your envelope printer setting. There is also DL envelope 110mm x 220mm.

// multipart/mixed
// |- multipart/alternative
// | |- text/plain
// | `- multipart/related // | |- text/html // |`- image/png
// `- attachments..

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
return quoteEscaper.Replace(s)
}

## Roadmap

- proper command line response when no parameters are set

* --log param
* make loggers Logger.Log??? interface methods more precise, example: LogSent(email, confirmation string) | this way the logger can determine verbosity
* https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
* reset locks on distributors - have .Reset() interface function for them
* make low-level MIME writer! cache resizedReadCloser, also cache all attachments as base64! into UserCacheDir?
  ++ use message.Validate() to uncover errors! including checking if From has 2 addresses!
* Ensure that "From" is set up correctly with just one email address - MG queues it with 2 and never delivers! MG sux, write AWS SNS queue driver?
* // TODO: look into https://github.com/modoboa/modoboa server.
* Event invitations markup: https://developers.google.com/gmail/markup/reference/event-reservation#basic_event_reminder_without_a_ticket

* Multiple-program lock prevention for the same file?
* my personal info is in git repo - has to be nuked before moving to github
* !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
* progress function should take a pointer to the message?
* alternative API: https://docs.sendwithses.com/ - add it as a driver to "providers"

- support Cuelang! https://github.com/cuelang/cue
