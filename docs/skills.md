# Skills System

The skills system is one of aidev's most powerful features. Skills are reusable prompt templates that encode expert knowledge for common coding tasks. They allow you to standardize workflows, share expertise with your team, and automate repetitive patterns.

---

## Overview

A skill is a Markdown file with YAML frontmatter that defines:

- **Metadata**: Name, description, version
- **Inputs**: Parameters the skill accepts
- **Prompt**: The instructions template sent to the AI model

Skills can be stored at two levels:

```
~/.aidev/skills/          # Global skills (available in all projects)
.aidev/skills/            # Project skills (available only in this project)
```

Project skills take precedence over global skills with the same name.

---

## Using Skills

### In the REPL

```bash
# List all available skills
/skills list

# Run a skill
/skills run code-review

# Run a skill with arguments
/skills run test-gen src/utils/parser.ts

# Run a skill with named arguments
/skills run refactor --file src/app.ts --style functional
```

### From the Command Line

```bash
# Run a skill directly
aidev skill run code-review src/auth.ts

# List skills
aidev skill list

# Show skill details
aidev skill show code-review
```

---

## Built-in Skills

### code-review

Comprehensive code review with severity ratings.

```bash
/skills run code-review src/index.ts
```

**Inputs:**
| Parameter | Type | Required | Description |
|---|---|---|---|
| `file` | `string` | Yes | File path to review |
| `focus` | `string` | No | Specific area to focus on (e.g., "security", "performance") |

### refactor

Suggest and apply refactoring improvements.

```bash
/skills run refactor src/utils/helper.ts --style functional
```

**Inputs:**
| Parameter | Type | Required | Description |
|---|---|---|---|
| `file` | `string` | Yes | File path to refactor |
| `style` | `string` | No | Target style (`functional`, `oop`, `minimal`) |

### test-gen

Generate unit tests with coverage analysis.

```bash
/skills run test-gen src/services/auth.ts --framework vitest
```

**Inputs:**
| Parameter | Type | Required | Description |
|---|---|---|---|
| `file` | `string` | Yes | File path to generate tests for |
| `framework` | `string` | No | Test framework (`vitest`, `jest`, `mocha`) |
| `coverage` | `boolean` | No | Include coverage analysis (default: true) |

### doc-gen

Generate documentation for functions and modules.

```bash
/skills run doc-gen src/lib/parser.ts
```

**Inputs:**
| Parameter | Type | Required | Description |
|---|---|---|---|
| `file` | `string` | Yes | File path to document |
| `style` | `string` | No | Documentation style (`jsdoc`, `tsdoc`, `plain`) |

### explain

Explain code with configurable detail levels.

```bash
/skills run explain src/core/engine.ts --level detailed
```

**Inputs:**
| Parameter | Type | Required | Description |
|---|---|---|---|
| `file` | `string` | Yes | File path to explain |
| `level` | `string` | No | Detail level (`brief`, `detailed`, `expert`) |

### optimize

Identify and fix performance bottlenecks.

```bash
/skills run optimize src/database/query.ts
```

**Inputs:**
| Parameter | Type | Required | Description |
|---|---|---|---|
| `file` | `string` | Yes | File path to optimize |
| `metric` | `string` | No | Target metric (`speed`, `memory`, `bundle-size`) |

---

## Creating Custom Skills

### File Format

Create a Markdown file with YAML frontmatter:

```markdown
---
name: my-skill
description: What this skill does
version: 1.0.0
author: your-name
tags:
  - typescript
  - testing
input:
  - name: file
    type: string
    description: Target file path
    required: true
  - name: style
    type: string
    description: Output style
    required: false
    default: standard
---

You are an expert TypeScript developer specializing in {{style}} patterns.

Analyze the file at {{file}} and:

1. Identify all exported functions and their signatures
2. Check for type safety issues
3. Identify potential runtime errors
4. Suggest improvements for readability

Provide your analysis in the following format:

## Summary
Brief overview of the file

## Issues Found
List each issue with severity (critical/warning/info)

## Recommendations
Actionable improvement suggestions

## Improved Code
Refactored version if applicable
```

### Frontmatter Reference

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` | Yes | Unique skill identifier |
| `description` | `string` | Yes | Brief description of what the skill does |
| `version` | `string` | No | Semantic version (default: `1.0.0`) |
| `author` | `string` | No | Skill author name |
| `tags` | `string[]` | No | Tags for categorization and discovery |
| `input` | `Input[]` | No | List of input parameters |

### Input Parameter Schema

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` | Yes | Parameter name (used in `{{name}}` templates) |
| `type` | `string` | Yes | Parameter type (`string`, `number`, `boolean`) |
| `description` | `string` | Yes | Description of the parameter |
| `required` | `boolean` | No | Whether the parameter is required (default: `false`) |
| `default` | `any` | No | Default value if not provided |

### Template Variables

Use `{{variable}}` syntax in the prompt body to reference input parameters:

```markdown
Analyze the code in {{file}} using {{style}} conventions.
```

### Special Variables

| Variable | Description |
|---|---|
| `{{file_content}}` | Contents of the file specified by the `file` input |
| `{{project_context}}` | Summary of the current project context |
| `{{git_diff}}` | Current staged git diff |
| `{{language}}` | Detected programming language |

---

## Skill Composition

Skills can reference other skills using the `{{skill:name}}` helper:

```markdown
---
name: full-review
description: Comprehensive code review with tests
version: 1.0.0
input:
  - name: file
    type: string
    description: Target file
    required: true
---

First, perform a code review:

{{skill:code-review}}

Then generate tests for any issues found:

{{skill:test-gen}}
```

---

## Sharing Skills

### With Your Team

Commit project skills to version control:

```bash
# Add to .gitignore pattern (optional: skills are safe to commit)
# .aidev/skills/     # <-- do NOT add this; commit your skills

git add .aidev/skills/
git commit -m "feat(skills): add team code review skill"
```

### Exporting and Importing

```bash
# Export a skill
aidev skill export code-review ./my-code-review.md

# Import a skill
aidev skill import ./shared-skill.md

# Import to project level
aidev skill import ./shared-skill.md --project
```

---

## Best Practices

1. **Be specific** -- The more specific the prompt, the better the output. Define the exact format you want.
2. **Use examples** -- Include example output in the skill prompt to guide the AI model.
3. **Version your skills** -- Bump the version when making changes so teams can track updates.
4. **Test your skills** -- Run your skill against various inputs to ensure consistent quality.
5. **Keep skills focused** -- Each skill should do one thing well. Use composition for complex workflows.
6. **Document inputs** -- Always provide clear descriptions for all input parameters.
