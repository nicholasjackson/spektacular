# Bob-Shell Task Delegation Feature

Bob-Shell's delegation feature allows you to break down complex tasks into smaller, focused sub-tasks using the `new_task` tool.

## What Triggers Delegation

Bob-Shell will consider using delegation when you provide prompts that involve:

### Multi-Component Tasks
```
"Build a complete user authentication system with login, registration, 
password reset, and email verification"
```

### Parallel Workstreams
```
"Refactor this codebase: update all API endpoints to async/await, 
add error handling to services, and update tests to match"
```

### Mode-Switching Requirements
```
"First plan the architecture for a new feature, then implement it, 
and finally write comprehensive tests"
```

### Large-Scale Changes
```
"Analyze this monorepo and improve code quality in each package: 
fix linting issues, add missing tests, and update documentation"
```

### Explicit Delegation Requests
```
"Break this down into sub-tasks: create the frontend components, 
build the API endpoints, and set up the database schema"
```

## How to Use Delegation

### Using the new_task Tool

```xml
<new_task>
<prompt>Implement the login page for the website</prompt>
<mode>code</mode>
<files>src/auth/login.js,src/components/LoginForm.jsx</files>
<description>Create login UI and authentication logic</description>
</new_task>
```

### Parameters

- **prompt** (required): Clear instructions for the sub-task with specific success criteria
- **mode** (optional): Which mode to use - `plan`, `code`, `advanced`, or `ask`
- **files** (optional): Array of relative file paths to pass as context to the sub-task
- **description** (optional): Brief description for tracking purposes

## When to Use Delegation

**Good Use Cases:**
- Complex tasks with multiple independent components
- Tasks requiring different modes (planning vs. implementation)
- Large codebases where focused attention on specific modules is needed
- Parallel workflows without dependencies

**Not Recommended For:**
- Simple, single-step tasks
- Tasks requiring continuous shared context
- Highly interdependent operations

## How It Works

1. Main task identifies work that can be delegated
2. Sub-task is created with specific instructions and optional file context
3. Sub-task runs independently in its own session
4. Sub-task completes using `attempt_completion` tool
5. Results are returned to the main task for integration

## Example Workflow

```
Main Task: "Build a user authentication system"
  ├─ Sub-task 1: "Create login UI components" (mode: code)
  ├─ Sub-task 2: "Implement JWT authentication" (mode: code)
  └─ Sub-task 3: "Write integration tests" (mode: code)
```

Each sub-task receives only the files it needs and reports results back to coordinate the overall implementation.

## Best Practices

### Clear Success Criteria
Always provide specific, measurable success criteria in your sub-task prompts:
- ✅ "Create a login form with email/password fields, validation, and submit handler"
- ❌ "Make a login page"

### Appropriate Context
Only pass files that are directly relevant to the sub-task:
```xml
<files>src/auth/login.js,src/components/LoginForm.jsx</files>
```

### Mode Selection
Choose the appropriate mode for each sub-task:
- `plan`: For design and architecture decisions
- `code`: For implementation work
- `advanced`: For complex coding with more autonomy
- `ask`: For research and documentation

## Technical Details

- Sub-tasks run in isolated sessions
- File context is passed in a structured format
- Sub-tasks inherit parent configuration
- Results are aggregated back to the main task
- Each sub-task must complete with `attempt_completion`
