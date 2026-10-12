Please build me a one-page website that greets Mario fans.

It's for a Mario Fan: a visitor who loves Nintendo's Mario franchise and wants a fun greeting. When they open the home page, they should see the heading "Hello, Mario!" and a button labelled "Let's-a go!". Clicking the button shows "Wahoo!" on the page. The heading must read exactly "Hello, Mario!", the button must show "Wahoo!", and the automated tests must pass with `node --test` run from the project root.

Use plain HTML, CSS, and vanilla JavaScript, with no frameworks, build step, or npm packages. Put `index.html` at the project root, with `styles.css` and `app.js` next to it. The greeting logic goes in an ES module, `greeting.mjs`, that exports `wahoo()`, which returns "Wahoo!", and `app.js` uses it. Write the tests with Node's built-in test runner in `tests/site.test.mjs`, checking `wahoo()` and that `index.html` contains the heading. Node.js 22 or newer. Hosting and deployment are out of scope.
