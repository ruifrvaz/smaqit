Here are the answers to the open questions, keyed by question id:

- actors: Mario Fan, a visitor who loves Nintendo's Mario franchise and wants a fun greeting.
- use-cases: As a Mario Fan, I open the site's home page and see the heading "Hello, Mario!" and a button labelled "Let's-a go!". When I click the button, the page shows the message "Wahoo!".
- success-metrics: The heading reads exactly "Hello, Mario!"; clicking "Let's-a go!" shows "Wahoo!"; the automated tests pass with `node --test` from the project root.
- add-or-revise: add new specifications; none exist yet.

If any other question is open, these are the complete requirements:

- Technology (fixed): plain HTML, CSS, and vanilla JavaScript, with no frameworks, build step, or npm packages. `index.html` at the project root, with `styles.css` and `app.js` next to it.
- Greeting logic: an ES module, `greeting.mjs`, exporting `wahoo()`, which returns "Wahoo!".
- Tests: Node's built-in test runner in `tests/site.test.mjs`.
- Runtime: Node.js 22 or newer.
- Hosting and deployment are out of scope.

Continue the Development phase with these answers.
