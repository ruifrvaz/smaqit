Build a one-page website that greets Mario fans.

Actors:
- Mario Fan: a visitor who loves Nintendo's Mario franchise and wants a fun greeting.

Use case:
- As a Mario Fan, I open the site's home page and see the heading "Hello, Mario!" and a button labelled "Let's-a go!". When I click the button, the page shows the message "Wahoo!".

Success metrics:
- The home page heading reads exactly "Hello, Mario!".
- Clicking the "Let's-a go!" button shows "Wahoo!" on the page.
- The automated tests pass with `node --test` run from the project root.

Technology (fixed; do not substitute):
- Plain HTML, CSS, and vanilla JavaScript. No frameworks, no build step, no npm packages.
- `index.html` at the project root, with `styles.css` and `app.js` next to it.
- The greeting logic lives in an ES module, `greeting.mjs`, that exports a function `wahoo()` returning "Wahoo!". `app.js` uses it.
- Tests use Node's built-in test runner (`node:test` and `node:assert`) in `tests/site.test.mjs`. They check the result of `wahoo()` and that `index.html` contains the "Hello, Mario!" heading.
- Node.js 22 or newer.

Constraints: none beyond the above. Hosting and deployment are out of scope.
