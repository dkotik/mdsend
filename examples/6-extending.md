---
subject: extend one letter from letter
from: Test Author <joedoe@test.com>
to: extends@test.com
extends: 3-scheduling.md
---

# Extend One Letter with Another

Merge the content and frontmatter fields into a letter from another file by using the `extends` field.

```yaml
extends:
  # this letter will inherit the content
  # and configuration values of `template.md`
  - template.md
  # load a configuration file
  - config.yaml
```

Any Markdown content below the last horizontal rule is added as a footer to the current letter. You may nest the extensions deep as long as there are no circular references.

## Content

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce commodo sapien sed magna eleifend, a aliquam nunc condimentum. Sed feugiat nibh at lorem malesuada, lobortis tincidunt urna bibendum. Maecenas pulvinar, quam id tincidunt auctor, sem est tincidunt nulla, ac fringilla velit sapien eget nisl.

Integer fringilla ipsum in elit tempus varius. Phasellus viverra velit in pulvinar bibendum. Pellentesque non varius est, in faucibus felis. Praesent sem tellus, consequat id sem nec, tempor efficitur mi. Proin et sem condimentum, dictum dolor vitae, blandit mi. Vestibulum eget diam ante. Etiam facilisis mi non augue malesuada placerat. Maecenas porttitor nulla a sem mattis eleifend. Fusce at neque augue.

Integer at odio laoreet orci iaculis hendrerit. Curabitur sollicitudin volutpat dui at sodales. Vestibulum elit magna, semper nec dolor eget, pharetra tempus augue.

Suspendisse iaculis tortor at massa congue condimentum. Ut interdum leo vel dignissim feugiat. Pellentesque volutpat molestie interdum. Vivamus vel purus sem. In eu hendrerit purus, vel gravida libero.
