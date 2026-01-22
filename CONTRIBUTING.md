# Contributing to smaqit

Thanks for helping improve smaqit.

## Ways to Contribute

- Report bugs and request features via GitHub Issues
- Improve documentation (README, wiki, prompts, templates)
- Submit pull requests for fixes and enhancements

## Development Setup

**Prerequisites:** Go 1.25+

```bash
git clone https://github.com/ruifrvaz/smaqit.git
cd smaqit/installer
make build
```

## Testing

- Manual and workflow guidance: `docs/wiki/workflows/testing-smaqit.md`
- Standardized test directory: `installer/test/`

## Making Changes

### Scope and Style

- Keep changes focused and consistent with existing structure
- Prefer small PRs over large refactors
- Avoid introducing new concepts/pages unless needed

### Documentation vs Framework

- `framework/` is **agent-facing** execution instructions
- `docs/wiki/` is **human-facing** context and rationale

## Pull Requests

1. Create a branch from `main`
2. Make changes with clear commit messages
3. Ensure build/test workflows still pass
4. Open a PR describing:
   - What changed and why
   - How you tested
   - Any breaking changes

## License

By contributing, you agree that your contributions are licensed under the MIT License (see `LICENSE`).
