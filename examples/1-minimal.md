---
subject: minimal example message
from: Test Author <joedoe@test.com>
to: to@test.com
---

# Introduction

This is the simplest letter that can be sent. It has one recipient.

- _Letter_: this file.
- _Message_: a rendered copy of this file for each recipient.

Sending requires a set of mailing service credentials
provided as environment variables.

```sh
export SMTP_HOST=...
export SMTP_PORT=...
export SMTP_USERNAME=...
export SMTP_PASSWORD=...
mdsend 1-minimal.md
```

## Separate Queuing

The `continue` command will resume delivery of any 
remaining queued messages. Messages can added to the
queue with a separate command:

```sh
# add first letter
mdsend queue add 1-minimal.md
# add second letter
mdsend queue add 2-attachments.md
# send everything that is in the queue and stop when done
mdsend continue
# send everything that is in the queue and wait for more messages forever
mdsend continue --forever
```

## Command Line Flags

- `--queue`: change database where the messages are stored.
- `--from`: override the author of the messages. 
- `--to`: add another recipient to each queued letter. Repeat the flag
  as many times as needed.
- `--cc`: equivalent to `--to` flag. Kept for historic reasons.
- `--bcc`: equivalent to `--to` flag. Kept for historic reasons.
- `--forever`: keep the process perpetually alive. If more
  letters are added to the queue at any time, they will be sent.

# Sample Content

> A Markdown file requires valid 'from', 'to', and 'subject' fields
> in the frontmatter.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce
commodo sapien sed magna eleifend, a aliquam nunc condimentum.
Sed feugiat nibh at lorem malesuada, lobortis tincidunt urna
bibendum. Maecenas pulvinar, quam id tincidunt auctor, sem est
tincidunt nulla, ac fringilla velit sapien eget nisl.

Integer fringilla ipsum in elit `tempus varius`. Phasellus viverra
velit in pulvinar bibendum. Pellentesque non varius est, in faucibus
felis. Praesent sem tellus, consequat id sem nec, tempor efficitur
mi. Proin et sem condimentum, dictum dolor vitae, blandit mi.
Vestibulum eget diam ante. Etiam facilisis mi non augue malesuada
placerat. Maecenas porttitor nulla a sem mattis eleifend. Fusce at
neque augue.

Integer at odio laoreet orci iaculis hendrerit. Curabitur
sollicitudin volutpat dui at sodales. Vestibulum elit magna, semper
nec dolor eget, pharetra tempus augue.

Suspendisse iaculis tortor at massa congue condimentum. Ut interdum
leo vel dignissim feugiat. Pellentesque volutpat molestie interdum.
Vivamus vel purus sem. In eu hendrerit purus, vel gravida libero.
