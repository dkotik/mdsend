---
subject: example message for {{ .Recipient.name }}
from: Test Author <joedoe@test.com>
to:
  - name: Full Name
    email: to@test.com
    first_name: FirstName
someVariable: test
templates:
  - ../internal/html/templates/default.html
  - mdsend://default.html
headers:
  X-Template-Test: "{{ .Frontmatter.someVariable }}"
---

Hello {{ .Recipient.first_name }},

The Markdown content itself is a `Golang` text template.
To wrap the output of a letter into an `HTML` template, specify
the list of files to load using the `template` field in the
frontmatter. The last template in the list becomes the root
template. Other templates can be called by their base file
name:

```
{{ if and "render only if true" false }}
{{ template "default.html" . }}
{{ end }}
```

Any sub-templates and blocks defined here, in the contents
of the letter are available in all other templates.

The frontmatter subject and the header values are also treated
as text templates. If the header template is rendered empty,
it will be ommitted from the sent message.

## Context

Template execution context contains the following fields:

- **Frontmatter:** the raw Markdown frontmatter as `map[string]any`.
- **Recipient:** the raw message target as `map[string]any`. It
  will contain `name` and `email` fields and any other fields
  fromt the contact source.
- **Content:** rendered Markdown body.
- **Schedule:** parsed scheduling directives.
- **IsPlainText:** true value, when the rendered content is
  alternative plain text of the electronic mail message.

## HTML or PlainText

Let's insert a centered large text title at the top of this letter:

{{ define "title" }}
  {{ titlecase .Frontmatter.subject }}
{{ end }}

Title sub-template is supported by all default _mdsend_ templates,
but the title text is not automatically included into the body
of the alternative plain text message.

The same is true of any text present in the
HTML template or inserted there by template execution.
You may add those elements back by using `.IsPlainText` context value.

{{ if .IsPlainText }}
  {{ template "title" . }}
{{ end }}

## Template Functions

- **skipMessageIfTrue:** skip current recipient is the argument to this function call evaluates to an unempty value.

## Reify Templates

Special `reify` function prints and stores the output of a named `Golang` block
or template. Calling `reify` again on the same template, prints
the initial output. Each recipient starts with an empty output cache.

Here is an example of declaring a template that encodes recipient
address as a URL-safe field once to be used several times:

```
{{ define "token" }}
  {{ urlQuery .Recipient.email }}
{{ end }}

URL: https://test.com/?unsubscribe={{ reify "token" }}
```

Reified values are marked as safe HTML. This prevents Go templating
engine from aggressively sanitizing their contents. For example, angle
brackets around a template value or inside it become `&gt;` and `&lt;`
to prevent cross-site scripting exploits. Put the angle brackets
inside the reified template, and they will render correctly.
