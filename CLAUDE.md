# Claude Guidelines for Polykeys

## Commit Guidelines

This project uses **Conventional Commits** for commit messages. All commits should follow this format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `test`: Adding or updating tests
- `refactor`: Code refactoring without changing behavior
- `docs`: Documentation changes
- `chore`: Maintenance tasks (dependencies, config, etc.)
- `perf`: Performance improvements
- `style`: Code style changes (formatting, missing semi-colons, etc.)

### Scopes

- `domain`: Domain layer (entities, interfaces)
- `usecases`: Use cases layer
- `adapters`: Adapters layer (config, devices, layouts, storage, tray)
- `cli`: CLI tool (polykeys command)
- `daemon`: Daemon (polykeysd)
- `config`: Configuration management
- `deps`: Dependencies

### Examples

```bash
feat(domain): add Device, Layout, and Mapping entities
test(domain): add comprehensive tests for domain entities
feat(config): add Lua configuration loader
docs: add README and LICENSE files
chore: initialize Go module and project structure
```

### Important Rules

1. **Atomic commits**: Each commit should represent one logical change
2. **No Claude Code references**: Do not mention "Claude Code" in commit messages
3. **Clear descriptions**: Use imperative mood ("add feature" not "added feature")
4. **Test commits**: Separate test commits from implementation commits when relevant
5. **Scope specificity**: Use appropriate scopes to organize changes

### Commit Strategy

When working on a feature:
1. Start with the core implementation
2. Add tests in a separate commit if substantial
3. Update documentation if needed
4. Refactor if necessary (separate commit)

Example sequence:
```bash
feat(domain): add Device entity with ID and metadata
test(domain): add tests for Device entity
feat(usecases): add SwitchLayoutUseCase
feat(adapters): add Lua config loader
test(adapters): add tests for Lua config loader
docs: update README with configuration instructions
```

## Architecture Principles

- Follow **Clean Architecture** (domain → usecases → adapters → infrastructure)
- Write **tests first** (TDD approach)
- Keep **dependencies pointing inward** (domain has no external deps)
- Use **interfaces** for all external interactions
- Maintain **platform separation** in adapters (Linux/macOS/Windows)

## Testing

- All domain logic must be tested
- Use table-driven tests when appropriate
- Mock external dependencies in use case tests
- Integration tests for adapters when possible

## Code Style

- Follow standard Go conventions
- Use meaningful variable names
- Keep functions small and focused
- Document exported functions and types
- Avoid premature optimization
